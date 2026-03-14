package temporalcli_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tarEntry is a file to include in a synthetic GitHub tarball.
type tarEntry struct {
	Name    string
	Content string
}

// buildGitHubTarball creates a tar.gz in the format returned by
// codeload.github.com. Every path is prefixed with prefix (e.g.
// "org-repo-sha/").
func buildGitHubTarball(t *testing.T, prefix string, files []tarEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	// Collect unique directory entries.
	dirs := map[string]struct{}{prefix: {}}
	for _, f := range files {
		d := path.Dir(prefix + f.Name)
		for d != "." && d != "" {
			dirs[d+"/"] = struct{}{}
			d = path.Dir(d)
		}
	}
	sorted := make([]string, 0, len(dirs))
	for d := range dirs {
		sorted = append(sorted, d)
	}
	sort.Strings(sorted)
	for _, d := range sorted {
		require.NoError(t, tw.WriteHeader(&tar.Header{
			Typeflag: tar.TypeDir, Name: d, Mode: 0o755,
		}))
	}
	for _, f := range files {
		data := []byte(f.Content)
		require.NoError(t, tw.WriteHeader(&tar.Header{
			Typeflag: tar.TypeReg,
			Name:     prefix + f.Name,
			Size:     int64(len(data)),
			Mode:     0o644,
		}))
		_, err := tw.Write(data)
		require.NoError(t, err)
	}
	require.NoError(t, tw.Close())
	require.NoError(t, gw.Close())
	return buf.Bytes()
}

// serveSamples starts an httptest.Server that mimics the two GitHub endpoints
// the sample commands need: raw content (manifest) and codeload (tarball).
func serveSamples(t *testing.T, manifest string, tarball []byte) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/tar.gz/"):
			w.Header().Set("Content-Type", "application/gzip")
			_, _ = w.Write(tarball)
		case strings.HasSuffix(r.URL.Path, "temporal-samples.yaml"):
			_, _ = w.Write([]byte(manifest))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

const testPythonManifest = `version: 1
language: python
repo: temporalio/samples-python
scaffold:
  pyproject.toml: |
    [project]
    name = "{{name}}"
    version = "0.1.0"
    requires-python = ">=3.10"
    dependencies = [{{dependencies}}]
    [build-system]
    requires = ["hatchling"]
    build-backend = "hatchling.build"
`

func testPythonTarball(t *testing.T) []byte {
	t.Helper()
	return buildGitHubTarball(t, "temporalio-samples-python-abc1234/", []tarEntry{
		{"temporal-samples.yaml", testPythonManifest},
		{"hello/temporal-sample.yaml", "description: Basic hello world samples\ndependencies:\n  - \"temporalio>=1.23.0,<2\"\n"},
		{"hello/__init__.py", ""},
		{"hello/my_activity.py", "from temporalio import activity\n"},
		{"hello/worker.py", "import asyncio\n"},
		{"hello/README.md", "# Hello Sample\n\nRun with: uv run hello/worker.py\n"},
		{"encryption/temporal-sample.yaml", "description: End-to-end encryption with a custom codec\n"},
		{"encryption/__init__.py", ""},
		{"encryption/worker.py", "# worker\n"},
	})
}

// TestSample_List verifies that `temporal sample list <language>` downloads
// the repo tarball, discovers samples by scanning for temporal-sample.yaml,
// and prints names with descriptions.
func TestSample_List(t *testing.T) {
	srv := serveSamples(t, testPythonManifest, testPythonTarball(t))
	t.Setenv("TEMPORAL_SAMPLES_BASE_URL", srv.URL)

	h := NewCommandHarness(t)
	res := h.Execute("sample", "list", "python")

	require.NoError(t, res.Err)
	out := res.Stdout.String()
	assert.Contains(t, out, "hello")
	assert.Contains(t, out, "Basic hello world samples")
	assert.Contains(t, out, "encryption")
	assert.Contains(t, out, "End-to-end encryption with a custom codec")
}

// TestSample_Init_Python verifies that `temporal sample init python <sample>`
// creates a project directory with generated scaffold files (pyproject.toml),
// README at the root, and sample files nested under a package subdirectory.
func TestSample_Init_Python(t *testing.T) {
	srv := serveSamples(t, testPythonManifest, testPythonTarball(t))
	t.Setenv("TEMPORAL_SAMPLES_BASE_URL", srv.URL)
	t.Chdir(t.TempDir())

	h := NewCommandHarness(t)
	res := h.Execute("sample", "init", "python", "hello")

	require.NoError(t, res.Err)

	// Scaffold: pyproject.toml generated with template vars expanded.
	pyproject, err := os.ReadFile(filepath.Join("hello", "pyproject.toml"))
	require.NoError(t, err)
	assert.Contains(t, string(pyproject), `name = "hello"`)
	assert.Contains(t, string(pyproject), `"temporalio>=1.23.0,<2"`)

	// README copied to project root.
	readme, err := os.ReadFile(filepath.Join("hello", "README.md"))
	require.NoError(t, err)
	assert.Contains(t, string(readme), "Hello Sample")

	// Sample files nested under package dir to preserve absolute imports.
	assert.FileExists(t, filepath.Join("hello", "hello", "__init__.py"))
	assert.FileExists(t, filepath.Join("hello", "hello", "my_activity.py"))
	assert.FileExists(t, filepath.Join("hello", "hello", "worker.py"))

	// Manifest files excluded from output.
	assert.NoFileExists(t, filepath.Join("hello", "temporal-sample.yaml"))
	assert.NoFileExists(t, filepath.Join("hello", "hello", "temporal-sample.yaml"))

	// Stdout shows next-step instructions.
	assert.Contains(t, res.Stdout.String(), "hello")
}

// TestSample_Init_TypeScript_FlatCopy verifies that when the scaffold is empty
// (TypeScript), sample files are copied flat — no nesting.
func TestSample_Init_TypeScript_FlatCopy(t *testing.T) {
	manifest := "version: 1\nlanguage: typescript\nrepo: temporalio/samples-typescript\nscaffold: {}\n"
	tarball := buildGitHubTarball(t, "temporalio-samples-typescript-def5678/", []tarEntry{
		{"temporal-samples.yaml", manifest},
		{"hello-world/temporal-sample.yaml", "description: Basic hello world workflow\n"},
		{"hello-world/package.json", `{"name": "hello-world"}`},
		{"hello-world/src/workflows.ts", "export async function greet() { return 'hello'; }\n"},
		{"hello-world/src/activities.ts", "export async function sayHello() { return 'hello'; }\n"},
		{"hello-world/README.md", "# Hello World\n"},
	})

	srv := serveSamples(t, manifest, tarball)
	t.Setenv("TEMPORAL_SAMPLES_BASE_URL", srv.URL)
	t.Chdir(t.TempDir())

	h := NewCommandHarness(t)
	res := h.Execute("sample", "init", "typescript", "hello-world")

	require.NoError(t, res.Err)

	// Empty scaffold: flat copy, no nesting.
	assert.FileExists(t, filepath.Join("hello-world", "package.json"))
	assert.FileExists(t, filepath.Join("hello-world", "src", "workflows.ts"))
	assert.FileExists(t, filepath.Join("hello-world", "src", "activities.ts"))
	assert.FileExists(t, filepath.Join("hello-world", "README.md"))
	assert.NoDirExists(t, filepath.Join("hello-world", "hello-world"))

	// Manifest excluded.
	assert.NoFileExists(t, filepath.Join("hello-world", "temporal-sample.yaml"))
}

// TestSample_Init_GitHubURL verifies that a full GitHub URL can be used
// instead of positional language + sample arguments.
func TestSample_Init_GitHubURL(t *testing.T) {
	srv := serveSamples(t, testPythonManifest, testPythonTarball(t))
	t.Setenv("TEMPORAL_SAMPLES_BASE_URL", srv.URL)
	t.Chdir(t.TempDir())

	h := NewCommandHarness(t)
	res := h.Execute("sample", "init",
		"https://github.com/temporalio/samples-python/tree/main/hello")

	require.NoError(t, res.Err)
	assert.FileExists(t, filepath.Join("hello", "hello", "__init__.py"))
}

func TestSample_Init_NoArgs(t *testing.T) {
	h := NewCommandHarness(t)
	res := h.Execute("sample", "init")
	assert.Error(t, res.Err)
}

func TestSample_List_NoArgs(t *testing.T) {
	h := NewCommandHarness(t)
	res := h.Execute("sample", "list")
	assert.Error(t, res.Err)
}
