package temporalcli

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-isatty"
	"gopkg.in/yaml.v3"
)

// sampleManifest is parsed from temporal-sample.yaml. A single file holds
// shared config (scaffold templates, rewrite rules) alongside a list of
// sample entries. Official repos place one at the repo root listing all
// samples; third-party repos may place one next to a single sample with
// path: "." to refer to the enclosing directory.
// TODO: support discovering multiple temporal-sample.yaml files in a repo.
type sampleManifest struct {
	Version        int               `yaml:"version"`
	Language       string            `yaml:"language"`
	Scaffold       map[string]string `yaml:"scaffold"`
	RewriteImports *rewriteRule      `yaml:"rewrite_imports"`
	RootFiles      []string          `yaml:"root_files"`
	Samples        []sampleSpec      `yaml:"samples"`
}

type rewriteRule struct {
	From string `yaml:"from"`
	Glob string `yaml:"glob"`
}

// sampleSpec describes a single sample within a manifest.
// Path is relative to the manifest file's directory.
// Dest, if set, overrides the default destination prefix (the sample name)
// for nested extraction. This is needed when the repo layout differs from
// the standalone project layout (e.g. Java's "core/..." prefix).
type sampleSpec struct {
	Path         string         `yaml:"path"`
	Dest         string         `yaml:"dest"`
	Description  string         `yaml:"description"`
	Dependencies []string       `yaml:"dependencies"`
	Extra        map[string]any `yaml:",inline"`
}

const manifestFile = "temporal-sample.yaml"

var langRepos = map[string]string{
	"go":         "temporalio/samples-go",
	"java":       "temporalio/samples-java",
	"python":     "temporalio/samples-python",
	"typescript": "temporalio/samples-typescript",
	"dotnet":     "temporalio/samples-dotnet",
	"ruby":       "temporalio/samples-ruby",
}

func samplesBaseURL() string {
	return os.Getenv("TEMPORAL_SAMPLES_BASE_URL")
}

// samplesRef is the git ref used when fetching from official sample repos.
// This will be changed to "main" before merge, once manifests land on main.
const samplesRef = "cli-sample"

func rawContentURL(repo, ref, path string) string {
	if base := samplesBaseURL(); base != "" {
		return base + "/" + repo + "/" + ref + "/" + path
	}
	return "https://raw.githubusercontent.com/" + repo + "/" + ref + "/" + path
}

func tarballURL(repo, ref string) string {
	if base := samplesBaseURL(); base != "" {
		return base + "/" + repo + "/tar.gz/" + ref
	}
	return "https://codeload.github.com/" + repo + "/tar.gz/" + ref
}

// parseGitHubURL extracts (repo, ref, sample) from a URL like
// https://github.com/temporalio/samples-python/tree/main/hello
// The ref may contain slashes (e.g. feature/foo); the last path component
// is always treated as the sample name.
func parseGitHubURL(rawURL string) (repo, ref, sample string, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid URL: %w", err)
	}
	p := strings.TrimSuffix(strings.TrimPrefix(u.Path, "/"), "/")
	parts := strings.Split(p, "/")
	// Minimum: owner/repo/tree/ref/sample = 5 parts.
	if len(parts) < 5 || parts[2] != "tree" {
		return "", "", "", fmt.Errorf("expected URL like https://github.com/OWNER/REPO/tree/REF/SAMPLE")
	}
	return parts[0] + "/" + parts[1], strings.Join(parts[3:len(parts)-1], "/"), parts[len(parts)-1], nil
}

// stripTarPrefix removes the top-level GitHub directory from a tar entry name.
func stripTarPrefix(name string) string {
	if i := strings.Index(name, "/"); i >= 0 {
		return name[i+1:]
	}
	return name
}

func downloadTarball(ctx context.Context, url string) (io.ReadCloser, *tar.Reader, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, nil, fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}
	gr, err := gzip.NewReader(resp.Body)
	if err != nil {
		resp.Body.Close()
		return nil, nil, err
	}
	rc := &multiCloser{closers: []io.Closer{gr, resp.Body}}
	return rc, tar.NewReader(gr), nil
}

type multiCloser struct {
	closers []io.Closer
}

func (mc *multiCloser) Read([]byte) (int, error) { return 0, io.EOF }
func (mc *multiCloser) Close() error {
	var firstErr error
	for _, c := range mc.closers {
		if err := c.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func fetchManifest(ctx context.Context, repo, ref, path string) (*sampleManifest, error) {
	u := rawContentURL(repo, ref, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching manifest: HTTP %d", resp.StatusCode)
	}
	var m sampleManifest
	if err := yaml.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}
	return &m, nil
}

// resolveSamplePath computes the tarball-relative path for a sample.
func resolveSamplePath(manifestDir string, spec *sampleSpec) string {
	if manifestDir != "" {
		return filepath.Clean(manifestDir + "/" + spec.Path)
	}
	return filepath.Clean(spec.Path)
}

// lookupSample searches for a sample by name in the manifest. The name is
// matched against filepath.Base of the resolved path (manifestDir + spec.Path).
func lookupSample(m *sampleManifest, name, manifestDir string) *sampleSpec {
	for i := range m.Samples {
		resolved := resolveSamplePath(manifestDir, &m.Samples[i])
		if filepath.Base(resolved) == name {
			return &m.Samples[i]
		}
	}
	return nil
}

func (c *TemporalSampleListCommand) run(cctx *CommandContext, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("provide a language (go, java, python, typescript, dotnet, ruby)")
	}
	lang := strings.ToLower(args[0])
	repo, ok := langRepos[lang]
	if !ok {
		return fmt.Errorf("unsupported language %q (supported: go, java, python, typescript, dotnet, ruby)", lang)
	}

	manifest, err := fetchManifest(cctx, repo, samplesRef, manifestFile)
	if err != nil {
		return err
	}
	if manifest == nil {
		return fmt.Errorf("no %s found in %s", manifestFile, repo)
	}

	type entry struct{ name, desc string }
	entries := make([]entry, 0, len(manifest.Samples))
	for i := range manifest.Samples {
		s := &manifest.Samples[i]
		entries = append(entries, entry{filepath.Base(resolveSamplePath("", s)), s.Description})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].name < entries[j].name })

	maxName := 0
	for _, e := range entries {
		if len(e.name) > maxName {
			maxName = len(e.name)
		}
	}
	for _, e := range entries {
		fmt.Fprintf(cctx.Options.Stdout, "%-*s  %s\n", maxName, e.name, e.desc)
	}
	fmt.Fprintf(cctx.Options.Stdout, "\nhttps://github.com/%s\n", repo)
	return nil
}

func (c *TemporalSampleInitCommand) run(cctx *CommandContext, args []string) error {
	var repo, ref, sample string

	switch len(args) {
	case 0:
		return fmt.Errorf("provide a language and sample name, or a GitHub URL")
	case 1:
		if !strings.HasPrefix(args[0], "https://") {
			return fmt.Errorf("provide a language and sample name, or a GitHub URL")
		}
		var err error
		repo, ref, sample, err = parseGitHubURL(args[0])
		if err != nil {
			return err
		}
	case 2:
		lang := strings.ToLower(args[0])
		var ok bool
		repo, ok = langRepos[lang]
		if !ok {
			return fmt.Errorf("unsupported language %q (supported: go, java, python, typescript, dotnet, ruby)", lang)
		}
		sample = args[1]
		ref = samplesRef
	}

	ctx := cctx

	// Look for manifest at the repo root first.
	manifest, err := fetchManifest(ctx, repo, ref, manifestFile)
	if err != nil {
		return err
	}
	var spec *sampleSpec
	var manifestDir string
	if manifest != nil {
		spec = lookupSample(manifest, sample, "")
		if spec == nil {
			return fmt.Errorf("sample %q not found in %s", sample, repo)
		}
	} else {
		// No root manifest; try next to the sample directory.
		manifest, err = fetchManifest(ctx, repo, ref, sample+"/"+manifestFile)
		if err != nil {
			return err
		}
		if manifest != nil {
			manifestDir = sample
			spec = lookupSample(manifest, sample, manifestDir)
		} else {
			// No manifest anywhere; fall back to flat extraction.
			fmt.Fprintf(cctx.Options.Stdout, "Warning: no %s found in %s (ref %s); extracting flat\n", manifestFile, repo, ref)
			manifest = &sampleManifest{}
		}
	}

	// Resolve the tarball path for this sample.
	var tarballPath string
	if spec != nil {
		tarballPath = resolveSamplePath(manifestDir, spec)
	} else {
		tarballPath = sample
	}
	sampleName := filepath.Base(tarballPath)

	outputDir := c.OutputDir
	if outputDir == "" {
		outputDir = sampleName
	}
	if _, err := os.Stat(outputDir); err == nil {
		return fmt.Errorf("directory %q already exists", outputDir)
	}

	spin := newSpinner(cctx.Options.Stdout, fmt.Sprintf("Downloading %s from %s", sampleName, repo))
	spin.Start()

	rc, tr, err := downloadTarball(ctx, tarballURL(repo, ref))
	if err != nil {
		spin.Stop()
		return fmt.Errorf("downloading samples: %w", err)
	}
	defer rc.Close()

	samplePrefix := tarballPath + "/"
	nested := len(manifest.Scaffold) > 0

	// Determine the destination prefix for sample files within outputDir.
	var destPrefix string
	if nested {
		if spec != nil && spec.Dest != "" {
			destPrefix = spec.Dest
		} else {
			destPrefix = sampleName
		}
	}

	filesWritten := 0

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading tarball: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		rel := stripTarPrefix(hdr.Name)

		// Copy root_files entries to project root.
		if rootFileDest := matchRootFile(rel, manifest.RootFiles); rootFileDest != "" {
			outPath := filepath.Join(outputDir, rootFileDest)
			if err := writeFileFromTar(outPath, tr, hdr.Mode); err != nil {
				return err
			}
			continue
		}

		if !strings.HasPrefix(rel, samplePrefix) {
			continue
		}
		relToSample := strings.TrimPrefix(rel, samplePrefix)

		// Skip manifest files.
		if filepath.Base(rel) == manifestFile {
			continue
		}

		var outPath string
		if nested {
			if strings.EqualFold(relToSample, "README.md") {
				outPath = filepath.Join(outputDir, relToSample)
			} else {
				outPath = filepath.Join(outputDir, destPrefix, relToSample)
			}
		} else {
			outPath = filepath.Join(outputDir, relToSample)
		}

		if err := writeFileFromTar(outPath, tr, hdr.Mode); err != nil {
			return err
		}
		filesWritten++
	}

	spin.Stop()

	if filesWritten == 0 {
		return fmt.Errorf("sample %q not found in %s", sample, repo)
	}

	projectName := filepath.Base(outputDir)

	// Write scaffold files.
	for filename, tmpl := range manifest.Scaffold {
		content := expandTemplate(tmpl, projectName, spec)
		outPath := filepath.Join(outputDir, filename)
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(outPath, []byte(content), 0o644); err != nil {
			return err
		}
	}

	// Rewrite imports if configured.
	if manifest.RewriteImports != nil {
		oldPrefix := manifest.RewriteImports.From + "/" + sampleName
		newPrefix := projectName + "/" + sampleName
		if err := rewriteImports(outputDir, *manifest.RewriteImports, oldPrefix, newPrefix); err != nil {
			return err
		}
	}

	displayDir := outputDir
	if !filepath.IsAbs(outputDir) {
		displayDir = "./" + outputDir
	}
	fmt.Fprintf(cctx.Options.Stdout, "Created %s/\n\n  cd %s\n  cat README.md\n", displayDir, outputDir)
	return nil
}

func writeFileFromTar(path string, r io.Reader, mode int64) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, os.FileMode(mode))
}

func expandTemplate(tmpl, name string, spec *sampleSpec) string {
	s := strings.ReplaceAll(tmpl, "{{name}}", name)
	if spec != nil {
		quoted := make([]string, len(spec.Dependencies))
		for i, d := range spec.Dependencies {
			quoted[i] = `"` + d + `"`
		}
		deps := strings.Join(quoted, ", ")
		s = strings.ReplaceAll(s, "{{dependencies}}", deps)
		for k, v := range spec.Extra {
			s = strings.ReplaceAll(s, "{{"+k+"}}", fmt.Sprint(v))
		}
	}
	return s
}

func rewriteImports(dir string, rule rewriteRule, oldPrefix, newPrefix string) error {
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		matched, err := filepath.Match(rule.Glob, d.Name())
		if err != nil || !matched {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !strings.Contains(string(data), oldPrefix) {
			return nil
		}
		replaced := strings.ReplaceAll(string(data), oldPrefix, newPrefix)
		return os.WriteFile(path, []byte(replaced), 0o644)
	})
}

// matchRootFile checks if rel matches any root_files entry. Entries ending
// in "/" match as directory prefixes; others match exactly. Returns the
// relative destination path within the output dir, or "" if no match.
func matchRootFile(rel string, rootFiles []string) string {
	for _, rf := range rootFiles {
		if strings.HasSuffix(rf, "/") {
			if strings.HasPrefix(rel, rf) || rel+"/" == rf {
				return rel
			}
		} else if rel == rf {
			return rel
		}
	}
	return ""
}

// spinner shows a braille animation next to a message while work is in progress.
type spinner struct {
	w    io.Writer
	msg  string
	tty  bool
	done chan struct{}
	once sync.Once
	wg   sync.WaitGroup
}

var brailleFrames = [...]string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func newSpinner(w io.Writer, msg string) *spinner {
	tty := false
	if f, ok := w.(*os.File); ok {
		tty = isatty.IsTerminal(f.Fd())
	}
	return &spinner{w: w, msg: msg, tty: tty, done: make(chan struct{})}
}

func (s *spinner) Start() {
	if !s.tty {
		fmt.Fprintf(s.w, "%s...\n", s.msg)
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		i := 0
		t := time.NewTicker(80 * time.Millisecond)
		defer t.Stop()
		for {
			fmt.Fprintf(s.w, "\r%s %s", brailleFrames[i%len(brailleFrames)], s.msg)
			i++
			select {
			case <-s.done:
				fmt.Fprintf(s.w, "\r\033[K") // clear line
				return
			case <-t.C:
			}
		}
	}()
}

func (s *spinner) Stop() {
	s.once.Do(func() { close(s.done) })
	s.wg.Wait()
}
