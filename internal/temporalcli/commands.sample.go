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

	"gopkg.in/yaml.v3"
)

// repoManifest is the repo-level temporal-samples.yaml.
type repoManifest struct {
	Version        int               `yaml:"version"`
	Language       string            `yaml:"language"`
	Repo           string            `yaml:"repo"`
	Scaffold       map[string]string `yaml:"scaffold"`
	RewriteImports *rewriteRule      `yaml:"rewrite_imports"`
	SamplePath     string            `yaml:"sample_path"`
}

type rewriteRule struct {
	From string `yaml:"from"`
	Glob string `yaml:"glob"`
}

// sampleManifest is the per-sample temporal-sample.yaml.
type sampleManifest struct {
	Description  string   `yaml:"description"`
	Dependencies []string `yaml:"dependencies"`
	// Extra captures any additional template variables.
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
func parseGitHubURL(rawURL string) (repo, ref, sample string, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid URL: %w", err)
	}
	// Path: /{owner}/{repo}/tree/{ref}/{sample}
	parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	if len(parts) < 5 || parts[2] != "tree" {
		return "", "", "", fmt.Errorf("expected URL like https://github.com/OWNER/REPO/tree/REF/SAMPLE")
	}
	return parts[0] + "/" + parts[1], parts[3], parts[4], nil
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
	// Wrap both closers so caller can close everything.
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

func fetchRepoManifest(ctx context.Context, repo, ref string) (*repoManifest, error) {
	u := rawContentURL(repo, ref, "temporal-samples.yaml")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching manifest: HTTP %d", resp.StatusCode)
	}
	var m repoManifest
	if err := yaml.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}
	return &m, nil
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

	ctx := cctx
	rc, tr, err := downloadTarball(ctx, tarballURL(repo, "main"))
	if err != nil {
		return fmt.Errorf("downloading samples: %w", err)
	}
	defer rc.Close()

	// Scan for temporal-sample.yaml entries.
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
		// Parse the sample manifest.
		var sm sampleManifest
		if err := yaml.NewDecoder(tr).Decode(&sm); err != nil {
			continue
		}
		// Extract sample name from path: either "sample/temporal-sample.yaml"
		// or "sample_path/sample/temporal-sample.yaml".
		dir := strings.TrimSuffix(rel, "/temporal-sample.yaml")
		name := filepath.Base(dir)
		samples = append(samples, sampleEntry{Name: name, Description: sm.Description})
	}

	sort.Slice(samples, func(i, j int) bool { return samples[i].Name < samples[j].Name })

	// Find max name width for alignment.
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
		ref = "main"
	}

	ctx := cctx

	// Fetch repo manifest.
	manifest, err := fetchRepoManifest(ctx, repo, ref)
	if err != nil {
		return err
	}

	outputDir := c.OutputDir
	if outputDir == "" {
		outputDir = sample
	}
	if _, err := os.Stat(outputDir); err == nil {
		return fmt.Errorf("directory %q already exists", outputDir)
	}

	fmt.Fprintf(cctx.Options.Stdout, "Downloading %s from %s...\n", sample, repo)

	// Download tarball.
	rc, tr, err := downloadTarball(ctx, tarballURL(repo, ref))
	if err != nil {
		return fmt.Errorf("downloading samples: %w", err)
	}
	defer rc.Close()

	// Determine where in the tarball the sample lives.
	samplePrefix := sample + "/"
	if manifest.SamplePath != "" {
		samplePrefix = manifest.SamplePath + "/" + sample + "/"
	}

	nested := len(manifest.Scaffold) > 0

	// Determine the destination prefix for sample files within outputDir.
	var destPrefix string
	if nested {
		if manifest.SamplePath != "" {
			// For deep sample_path (e.g. "core/src/main/java/io/temporal/samples"),
			// strip the first path component to get the source tree layout.
			parts := strings.SplitN(manifest.SamplePath, "/", 2)
			if len(parts) == 2 {
				destPrefix = parts[1] + "/" + sample
			} else {
				destPrefix = sample
			}
		} else {
			destPrefix = sample
		}
	}

	var sm sampleManifest
	parsedSampleManifest := false
	filesWritten := 0

	// Stream through tarball extracting matching files.
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
		if !strings.HasPrefix(rel, samplePrefix) {
			continue
		}
		// Path relative to the sample directory.
		relToSample := strings.TrimPrefix(rel, samplePrefix)

		// Parse sample manifest if found.
		if relToSample == "temporal-sample.yaml" && !parsedSampleManifest {
			if err := yaml.NewDecoder(tr).Decode(&sm); err != nil {
				return fmt.Errorf("parsing sample manifest: %w", err)
			}
			parsedSampleManifest = true
			continue
		}

		// Skip manifest files.
		base := filepath.Base(rel)
		if base == "temporal-sample.yaml" || base == "temporal-samples.yaml" {
			continue
		}

		// Determine output path.
		var outPath string
		if nested {
			if relToSample == "README.md" {
				// README goes to project root.
				outPath = filepath.Join(outputDir, "README.md")
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

	if filesWritten == 0 && !parsedSampleManifest {
		return fmt.Errorf("sample %q not found in %s", sample, repo)
	}

	// Write scaffold files.
	for filename, tmpl := range manifest.Scaffold {
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
	if manifest.RewriteImports != nil {
		oldPrefix := manifest.RewriteImports.From + "/" + sample
		newPrefix := filepath.Base(outputDir) + "/" + sample
		if err := rewriteImports(outputDir, *manifest.RewriteImports, oldPrefix, newPrefix); err != nil {
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
		// Build dependencies string: comma-joined with quotes.
		quoted := make([]string, len(sm.Dependencies))
		for i, d := range sm.Dependencies {
			quoted[i] = `"` + d + `"`
		}
		deps := strings.Join(quoted, ", ")
		s = strings.ReplaceAll(s, "{{dependencies}}", deps)
		// Expand extra keys.
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
