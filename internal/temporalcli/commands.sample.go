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

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
	"github.com/temporalio/cli/internal/printer"
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
	Commands     []commandStep  `yaml:"commands"`
	Include      []includePath  `yaml:"include"`
	Extra        map[string]any `yaml:",inline"`
}

// includePath specifies an additional directory to extract alongside the sample.
type includePath struct {
	Path string `yaml:"path"`
	Dest string `yaml:"dest"`
}

type commandStep struct {
	Cmd         string `yaml:"cmd"`
	NewTerminal bool   `yaml:"new_terminal,omitempty"`
}

const manifestFile = "temporal-sample.yaml"

// sampleRow is the structured output record for `temporal sample list`.
type sampleRow struct {
	Language    string `json:"language"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// canonicalLanguages lists supported languages in display order.
var canonicalLanguages = []string{"dotnet", "go", "java", "python", "ruby", "typescript"}

var langRepos = map[string]string{
	"go":         "temporalio/samples-go",
	"java":       "temporalio/samples-java",
	"python":     "temporalio/samples-python",
	"py":         "temporalio/samples-python",
	"typescript": "temporalio/samples-typescript",
	"ts":         "temporalio/samples-typescript",
	"dotnet":     "temporalio/samples-dotnet",
	"csharp":     "temporalio/samples-dotnet",
	"cs":         "temporalio/samples-dotnet",
	"ruby":       "temporalio/samples-ruby",
	"rb":         "temporalio/samples-ruby",
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
	var languages []string
	if c.Language != "" {
		lang := strings.ToLower(c.Language)
		if _, ok := langRepos[lang]; !ok {
			return fmt.Errorf("unsupported language %q (supported: %s)", c.Language, strings.Join(canonicalLanguages, ", "))
		}
		languages = []string{lang}
	} else {
		languages = canonicalLanguages
	}

	rows, err := fetchSampleRows(cctx, languages)
	if err != nil {
		return err
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Language != rows[j].Language {
			return rows[i].Language < rows[j].Language
		}
		return rows[i].Name < rows[j].Name
	})

	opts := printer.StructuredOptions{Table: &printer.TableOptions{NoHeader: true}}
	if c.Language != "" {
		opts.Fields = []string{"Name", "Description"}
	}
	return cctx.Printer.PrintStructured(rows, opts)
}

// fetchSampleRows fetches manifests for the given languages concurrently and
// returns one sampleRow per sample entry.
func fetchSampleRows(ctx context.Context, languages []string) ([]sampleRow, error) {
	type result struct {
		rows []sampleRow
		err  error
	}
	results := make([]result, len(languages))
	var wg sync.WaitGroup
	for i, lang := range languages {
		wg.Add(1)
		go func(idx int, language string) {
			defer wg.Done()
			repo := langRepos[language]
			m, err := fetchManifest(ctx, repo, samplesRef, manifestFile)
			if err != nil {
				results[idx].err = fmt.Errorf("%s: %w", language, err)
				return
			}
			if m == nil {
				return
			}
			for k := range m.Samples {
				s := &m.Samples[k]
				results[idx].rows = append(results[idx].rows, sampleRow{
					Language:    language,
					Name:        filepath.Base(resolveSamplePath("", s)),
					Description: s.Description,
				})
			}
		}(i, lang)
	}
	wg.Wait()
	var rows []sampleRow
	for _, r := range results {
		if r.err != nil {
			return nil, r.err
		}
		rows = append(rows, r.rows...)
	}
	return rows, nil
}

func (c *TemporalSampleInitCommand) run(cctx *CommandContext, args []string) error {
	var repo, ref, sample string

	if c.Url != "" {
		var err error
		repo, ref, sample, err = parseGitHubURL(c.Url)
		if err != nil {
			return err
		}
	} else {
		if c.Name == "" || c.Language == "" {
			return fmt.Errorf("--name and --language are required (or use --url)")
		}
		lang := strings.ToLower(c.Language)
		var ok bool
		repo, ok = langRepos[lang]
		if !ok {
			return fmt.Errorf("unsupported language %q (supported: %s)", c.Language, strings.Join(canonicalLanguages, ", "))
		}
		sample = c.Name
		ref = samplesRef
	}

	ctx := cctx

	spin := newSpinner(cctx.Options.Stderr, fmt.Sprintf("Downloading %s from %s", sample, repo))
	spin.Start()
	defer spin.Stop()

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

	outputDir := c.Dir
	if outputDir == "" {
		outputDir = sampleName
	}
	if _, err := os.Stat(outputDir); err == nil {
		return fmt.Errorf("directory %q already exists", outputDir)
	}

	rc, tr, err := downloadTarball(ctx, tarballURL(repo, ref))
	if err != nil {
		return fmt.Errorf("downloading samples: %w", err)
	}
	defer rc.Close()

	nested := len(manifest.Scaffold) > 0

	// Build extraction rules: tarball prefix → dest prefix within outputDir.
	type extractRule struct {
		tarPrefix string
		destDir   string
		promote   bool // promote README.md to project root
	}
	var rules []extractRule
	if nested {
		destPrefix := sampleName
		if spec != nil && spec.Dest != "" {
			destPrefix = spec.Dest
		}
		rules = append(rules, extractRule{tarballPath + "/", destPrefix, true})
		if spec != nil {
			for _, inc := range spec.Include {
				rules = append(rules, extractRule{inc.Path + "/", inc.Dest, false})
			}
		}
	} else {
		rules = append(rules, extractRule{tarballPath + "/", "", true})
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

		// Skip manifest files.
		if filepath.Base(rel) == manifestFile {
			continue
		}

		// Match against extraction rules.
		var outPath string
		matched := false
		for _, rule := range rules {
			if !strings.HasPrefix(rel, rule.tarPrefix) {
				continue
			}
			relToDir := strings.TrimPrefix(rel, rule.tarPrefix)
			if nested {
				if rule.promote && strings.EqualFold(relToDir, "README.md") {
					outPath = filepath.Join(outputDir, relToDir)
				} else {
					outPath = filepath.Join(outputDir, rule.destDir, relToDir)
				}
			} else {
				outPath = filepath.Join(outputDir, relToDir)
			}
			matched = true
			break
		}
		if !matched {
			continue
		}

		if err := writeFileFromTar(outPath, tr, hdr.Mode); err != nil {
			return err
		}
		filesWritten++
	}

	if filesWritten == 0 {
		return fmt.Errorf("sample %q not found in %s", sample, repo)
	}

	projectName := filepath.Base(outputDir)

	// Write scaffold files.
	for filename, tmpl := range manifest.Scaffold {
		content := expandTemplate(tmpl, projectName, sampleName, spec)
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

	spin.Stop()

	fmt.Fprintln(cctx.Options.Stderr, color.GreenString("Created output directory: %s", outputDir))
	w := cctx.Options.Stdout
	comment := color.New(color.Faint).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()
	fmt.Fprintln(w)
	fmt.Fprintln(w, bold("cd "+outputDir))
	if spec != nil && len(spec.Commands) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, comment("# Local Temporal server: temporal server start-dev"))
		for _, step := range spec.Commands {
			if step.NewTerminal {
				fmt.Fprintln(w)
				fmt.Fprintln(w, comment("# In another terminal:"))
			}
			fmt.Fprintln(w, bold(step.Cmd))
		}
		fmt.Fprintln(w)
		fmt.Fprintln(w, comment("# Local Temporal UI: http://localhost:8233"))
	} else if readme, err := findReadme(outputDir); err == nil {
		fmt.Fprintln(w)
		fmt.Fprint(w, readme)
	}
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

// findReadme looks for a README file (case-insensitive) in dir and returns its content.
func findReadme(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if !e.IsDir() && strings.EqualFold(e.Name(), "README.md") {
			data, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				return "", err
			}
			return string(data), nil
		}
	}
	return "", fmt.Errorf("no README found")
}

func expandTemplate(tmpl, name, sample string, spec *sampleSpec) string {
	s := strings.ReplaceAll(tmpl, "{{name}}", name)
	s = strings.ReplaceAll(s, "{{sample}}", sample)
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
