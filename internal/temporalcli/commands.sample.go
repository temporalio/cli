package temporalcli

import (
	"archive/tar"
	"bytes"
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

type rewriteRule struct {
	From string `yaml:"from"`
	Glob string `yaml:"glob"`
}

// sampleManifest is the per-sample temporal-sample.yaml. It is the sole
// manifest: there is no repo-level manifest. Each sample is self-contained.
type sampleManifest struct {
	Description    string            `yaml:"description"`
	Dependencies   []string          `yaml:"dependencies"`
	Scaffold       map[string]string `yaml:"scaffold"`
	RewriteImports *rewriteRule      `yaml:"rewrite_imports"`
	RootFiles      []string          `yaml:"root_files"`
	DestPrefix     string            `yaml:"dest_prefix"`
	// Extra captures additional template variables (e.g. sdk_version).
	Extra map[string]any `yaml:",inline"`
}

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

func defaultRef() string {
	if v := os.Getenv("TEMPORAL_SAMPLES_REF"); v != "" {
		return v
	}
	return "main"
}

func tarballURL(repo, ref string) string {
	if base := samplesBaseURL(); base != "" {
		return base + "/" + repo + "/tar.gz/" + ref
	}
	return "https://codeload.github.com/" + repo + "/tar.gz/" + ref
}

// parseGitHubURL extracts (repo, ref, samplePath) from a URL like
// https://github.com/temporalio/samples-python/tree/main/hello
// or a deep path like
// https://github.com/org/repo/tree/main/deep/path/sample
func parseGitHubURL(rawURL string) (repo, ref, samplePath string, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid URL: %w", err)
	}
	parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	if len(parts) < 5 || parts[2] != "tree" {
		return "", "", "", fmt.Errorf("expected URL like https://github.com/OWNER/REPO/tree/REF/SAMPLE")
	}
	return parts[0] + "/" + parts[1], parts[3], strings.Join(parts[4:], "/"), nil
}

// stripTarPrefix removes the top-level GitHub directory from a tar entry name.
func stripTarPrefix(name string) string {
	if i := strings.Index(name, "/"); i >= 0 {
		return name[i+1:]
	}
	return name
}

func downloadTarballBytes(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func newTarReader(data []byte) (*tar.Reader, error) {
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return tar.NewReader(gr), nil
}

type sampleEntry struct {
	Name        string
	Description string
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

	data, err := downloadTarballBytes(cctx, tarballURL(repo, defaultRef()))
	if err != nil {
		return fmt.Errorf("downloading samples: %w", err)
	}

	tr, err := newTarReader(data)
	if err != nil {
		return err
	}

	var samples []sampleEntry
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading tarball: %w", err)
		}
		rel := stripTarPrefix(hdr.Name)
		if !strings.HasSuffix(rel, "/temporal-sample.yaml") {
			continue
		}
		var sm sampleManifest
		if err := yaml.NewDecoder(tr).Decode(&sm); err != nil {
			continue
		}
		dir := strings.TrimSuffix(rel, "/temporal-sample.yaml")
		name := filepath.Base(dir)
		samples = append(samples, sampleEntry{Name: name, Description: sm.Description})
	}

	sort.Slice(samples, func(i, j int) bool { return samples[i].Name < samples[j].Name })

	maxName := 0
	for _, s := range samples {
		if len(s.Name) > maxName {
			maxName = len(s.Name)
		}
	}
	for _, s := range samples {
		fmt.Fprintf(cctx.Options.Stdout, "%-*s  %s\n", maxName, s.Name, s.Description)
	}
	fmt.Fprintf(cctx.Options.Stdout, "\nhttps://github.com/%s\n", repo)
	return nil
}

func (c *TemporalSampleInitCommand) run(cctx *CommandContext, args []string) error {
	var repo, ref, sample, samplePath string

	switch len(args) {
	case 0:
		return fmt.Errorf("provide a language and sample name, or a GitHub URL")
	case 1:
		if !strings.HasPrefix(args[0], "https://") {
			return fmt.Errorf("provide a language and sample name, or a GitHub URL")
		}
		var err error
		repo, ref, samplePath, err = parseGitHubURL(args[0])
		if err != nil {
			return err
		}
		sample = filepath.Base(samplePath)
	case 2:
		lang := strings.ToLower(args[0])
		var ok bool
		repo, ok = langRepos[lang]
		if !ok {
			return fmt.Errorf("unsupported language %q (supported: go, java, python, typescript, dotnet, ruby)", lang)
		}
		sample = args[1]
		ref = defaultRef()
	}

	outputDir := c.OutputDir
	if outputDir == "" {
		outputDir = sample
	}
	if _, err := os.Stat(outputDir); err == nil {
		return fmt.Errorf("directory %q already exists", outputDir)
	}

	spin := newSpinner(cctx.Options.Stdout, fmt.Sprintf("Downloading %s from %s", sample, repo))
	spin.Start()

	data, err := downloadTarballBytes(cctx, tarballURL(repo, ref))
	if err != nil {
		spin.Stop()
		return fmt.Errorf("downloading samples: %w", err)
	}

	// Locate the sample's temporal-sample.yaml in the tarball.
	var sm sampleManifest
	var samplePrefix string
	foundManifest := false

	if samplePath != "" {
		samplePrefix = samplePath + "/"
	}

	tr, err := newTarReader(data)
	if err != nil {
		spin.Stop()
		return err
	}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			spin.Stop()
			return fmt.Errorf("reading tarball: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		rel := stripTarPrefix(hdr.Name)
		if !strings.HasSuffix(rel, "/temporal-sample.yaml") {
			continue
		}
		if samplePath != "" {
			// URL form: exact path match.
			if rel != samplePath+"/temporal-sample.yaml" {
				continue
			}
		} else {
			// <lang> <sample> form: match by directory basename.
			dir := strings.TrimSuffix(rel, "/temporal-sample.yaml")
			if filepath.Base(dir) != sample {
				continue
			}
			samplePrefix = dir + "/"
		}
		if err := yaml.NewDecoder(tr).Decode(&sm); err != nil {
			spin.Stop()
			return fmt.Errorf("parsing sample manifest: %w", err)
		}
		foundManifest = true
		break
	}

	if !foundManifest && samplePath == "" {
		spin.Stop()
		return fmt.Errorf("sample %q not found in %s", sample, repo)
	}

	nested := len(sm.Scaffold) > 0
	var destPrefix string
	if nested {
		if sm.DestPrefix != "" {
			destPrefix = sm.DestPrefix + "/" + sample
		} else {
			destPrefix = sample
		}
	}

	// Second pass: extract files.
	tr, err = newTarReader(data)
	if err != nil {
		spin.Stop()
		return err
	}
	filesWritten := 0
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			spin.Stop()
			return fmt.Errorf("reading tarball: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		rel := stripTarPrefix(hdr.Name)

		// Check root_files entries.
		if rootFileDest := matchRootFile(rel, sm.RootFiles); rootFileDest != "" {
			outPath := filepath.Join(outputDir, rootFileDest)
			if err := writeFileFromTar(outPath, tr, hdr.Mode); err != nil {
				spin.Stop()
				return err
			}
			continue
		}

		if !strings.HasPrefix(rel, samplePrefix) {
			continue
		}
		relToSample := strings.TrimPrefix(rel, samplePrefix)

		// Skip manifest files.
		base := filepath.Base(rel)
		if base == "temporal-sample.yaml" || base == "temporal-samples.yaml" {
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
			spin.Stop()
			return err
		}
		filesWritten++
	}

	spin.Stop()

	if filesWritten == 0 {
		return fmt.Errorf("sample %q not found in %s", sample, repo)
	}

	// Write scaffold files.
	for filename, tmpl := range sm.Scaffold {
		content := expandTemplate(tmpl, sample, &sm)
		outPath := filepath.Join(outputDir, filename)
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(outPath, []byte(content), 0o644); err != nil {
			return err
		}
	}

	// Rewrite imports if configured.
	if sm.RewriteImports != nil {
		oldPrefix := sm.RewriteImports.From + "/" + sample
		newPrefix := filepath.Base(outputDir) + "/" + sample
		if err := rewriteImports(outputDir, *sm.RewriteImports, oldPrefix, newPrefix); err != nil {
			return err
		}
	}

	fmt.Fprintf(cctx.Options.Stdout, "Created ./%s/\n\n  cd %s\n  cat README.md\n", outputDir, outputDir)
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

func expandTemplate(tmpl, name string, sm *sampleManifest) string {
	s := strings.ReplaceAll(tmpl, "{{name}}", name)
	if sm != nil {
		quoted := make([]string, len(sm.Dependencies))
		for i, d := range sm.Dependencies {
			quoted[i] = `"` + d + `"`
		}
		deps := strings.Join(quoted, ", ")
		s = strings.ReplaceAll(s, "{{dependencies}}", deps)
		for k, v := range sm.Extra {
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
	go func() {
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
	if s.tty {
		time.Sleep(10 * time.Millisecond)
	}
}
