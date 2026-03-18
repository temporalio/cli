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

const manifestFile = "temporal-sample.yaml"

// serveSamples starts an httptest.Server that serves a root manifest at any
// URL ending in /temporal-sample.yaml, and a tarball at any URL containing
// /tar.gz/.
func serveSamples(t *testing.T, manifest string, tarball []byte) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/tar.gz/"):
			w.Header().Set("Content-Type", "application/gzip")
			_, _ = w.Write(tarball)
		case strings.HasSuffix(r.URL.Path, "/"+manifestFile):
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
samples:
  - path: hello
    description: Basic hello world samples
    dependencies:
      - "temporalio>=1.23.0,<2"
    commands:
      - cmd: uv run hello/worker.py
      - cmd: uv run hello/starter.py
        new_terminal: true
  - path: encryption
    description: End-to-end encryption with a custom codec
`

func testPythonTarball(t *testing.T) []byte {
	t.Helper()
	return buildGitHubTarball(t, "temporalio-samples-python-abc1234/", []tarEntry{
		{"hello/__init__.py", ""},
		{"hello/my_activity.py", "from temporalio import activity\n"},
		{"hello/worker.py", "import asyncio\n"},
		{"hello/README.md", "# Hello Sample\n\nRun with: uv run hello/worker.py\n"},
		{"encryption/__init__.py", ""},
		{"encryption/worker.py", "# worker\n"},
	})
}

// TestSample_List verifies that `temporal sample list <language>` fetches
// the manifest and prints sample names with descriptions.
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
	assert.NoFileExists(t, filepath.Join("hello", manifestFile))
	assert.NoFileExists(t, filepath.Join("hello", "hello", manifestFile))

	// Stdout shows run commands from manifest.
	out := res.Stdout.String()
	assert.Contains(t, out, "uv run hello/worker.py")
	assert.Contains(t, out, "In another terminal")
	assert.Contains(t, out, "uv run hello/starter.py")
	assert.Contains(t, out, "temporal server start-dev")
	assert.Contains(t, out, "http://localhost:8233")
	assert.NotContains(t, out, "cat README.md")
}

// TestSample_Init_TypeScript_FlatCopy verifies that when the scaffold is empty
// (TypeScript), sample files are copied flat — no nesting.
func TestSample_Init_TypeScript_FlatCopy(t *testing.T) {
	manifest := `version: 1
language: typescript
scaffold: {}
samples:
  - path: hello-world
    description: Basic hello world workflow
`
	tarball := buildGitHubTarball(t, "temporalio-samples-typescript-def5678/", []tarEntry{
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
	assert.NoFileExists(t, filepath.Join("hello-world", manifestFile))
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

// TestSample_Init_GitHubURL_TrailingSlash verifies trailing slash is handled.
func TestSample_Init_GitHubURL_TrailingSlash(t *testing.T) {
	srv := serveSamples(t, testPythonManifest, testPythonTarball(t))
	t.Setenv("TEMPORAL_SAMPLES_BASE_URL", srv.URL)
	t.Chdir(t.TempDir())

	h := NewCommandHarness(t)
	res := h.Execute("sample", "init",
		"https://github.com/temporalio/samples-python/tree/main/hello/")

	require.NoError(t, res.Err)
	assert.FileExists(t, filepath.Join("hello", "hello", "__init__.py"))
}

// TestSample_Init_GitHubURL_RefWithSlash verifies refs containing slashes
// (e.g. feature branches) are parsed correctly.
func TestSample_Init_GitHubURL_RefWithSlash(t *testing.T) {
	srv := serveSamples(t, testPythonManifest, testPythonTarball(t))
	t.Setenv("TEMPORAL_SAMPLES_BASE_URL", srv.URL)
	t.Chdir(t.TempDir())

	h := NewCommandHarness(t)
	res := h.Execute("sample", "init",
		"https://github.com/temporalio/samples-python/tree/feature/foo/hello")

	require.NoError(t, res.Err)
	assert.FileExists(t, filepath.Join("hello", "hello", "__init__.py"))
}

// TestSample_Init_Go_ImportRewrite verifies that Go import paths are rewritten
// from the monorepo module path to the new standalone module name.
func TestSample_Init_Go_ImportRewrite(t *testing.T) {
	manifest := `version: 1
language: go
scaffold:
  go.mod: |
    module {{name}}
    go 1.23.0
    require go.temporal.io/sdk v1.41.0
rewrite_imports:
  from: github.com/temporalio/samples-go
  glob: "*.go"
samples:
  - path: helloworld
    description: Basic hello world workflow
`
	tarball := buildGitHubTarball(t, "temporalio-samples-go-abc1234/", []tarEntry{
		{"helloworld/helloworld.go", "package helloworld\n\nimport \"go.temporal.io/sdk/workflow\"\n"},
		{"helloworld/worker/main.go", "package main\n\nimport (\n\t\"github.com/temporalio/samples-go/helloworld\"\n)\n"},
		{"helloworld/starter/main.go", "package main\n\nimport (\n\t\"github.com/temporalio/samples-go/helloworld\"\n)\n"},
		{"helloworld/README.md", "# Hello World\n"},
	})

	srv := serveSamples(t, manifest, tarball)
	t.Setenv("TEMPORAL_SAMPLES_BASE_URL", srv.URL)
	t.Chdir(t.TempDir())

	h := NewCommandHarness(t)
	res := h.Execute("sample", "init", "go", "helloworld")

	require.NoError(t, res.Err)

	// Scaffold: go.mod generated with module name.
	gomod, err := os.ReadFile(filepath.Join("helloworld", "go.mod"))
	require.NoError(t, err)
	assert.Contains(t, string(gomod), "module helloworld")

	// Import paths rewritten from monorepo to standalone.
	worker, err := os.ReadFile(filepath.Join("helloworld", "helloworld", "worker", "main.go"))
	require.NoError(t, err)
	assert.Contains(t, string(worker), `"helloworld/helloworld"`)
	assert.NotContains(t, string(worker), "github.com/temporalio/samples-go")
}

// TestSample_Init_Java verifies that Java samples are extracted from the deep
// path, with Gradle scaffold files generated at the project root.
func TestSample_Init_Java(t *testing.T) {
	manifest := `version: 1
language: java
scaffold:
  build.gradle: |
    plugins { id 'java' }
    repositories { mavenCentral() }
    java { sourceCompatibility = JavaVersion.VERSION_11 }
    dependencies {
        implementation "io.temporal:temporal-sdk:{{sdk_version}}"
        implementation "io.temporal:temporal-envconfig:{{sdk_version}}"
        implementation "ch.qos.logback:logback-classic:1.5.6"
        implementation "commons-lang:commons-lang:2.6"
    }
    task execute(type: JavaExec) {
        mainClass = findProperty("mainClass") ?: ""
        classpath = sourceSets.main.runtimeClasspath
    }
root_files:
  - gradlew
  - gradlew.bat
  - gradle/
samples:
  - path: core/src/main/java/io/temporal/samples/hello
    dest: src/main/java/io/temporal/samples/hello
    description: Basic hello world samples
    sdk_version: "1.32.1"
`
	tarball := buildGitHubTarball(t, "temporalio-samples-java-abc1234/", []tarEntry{
		{"core/src/main/java/io/temporal/samples/hello/HelloActivity.java", "package io.temporal.samples.hello;\npublic class HelloActivity {}\n"},
		{"core/src/main/java/io/temporal/samples/hello/README.md", "# Hello\n"},
		{"gradlew", "#!/bin/sh\nexec gradle \"$@\"\n"},
		{"gradlew.bat", "@echo off\ngradle %*\n"},
		{"gradle/wrapper/gradle-wrapper.properties", "distributionUrl=https\\://services.gradle.org/distributions/gradle-8.5-bin.zip\n"},
	})

	srv := serveSamples(t, manifest, tarball)
	t.Setenv("TEMPORAL_SAMPLES_BASE_URL", srv.URL)
	t.Chdir(t.TempDir())

	h := NewCommandHarness(t)
	res := h.Execute("sample", "init", "java", "hello")

	require.NoError(t, res.Err)

	// Scaffold: build.gradle with sdk_version expanded.
	gradle, err := os.ReadFile(filepath.Join("hello", "build.gradle"))
	require.NoError(t, err)
	assert.Contains(t, string(gradle), `temporal-sdk:1.32.1`)

	// Java source preserved at deep path under project root.
	assert.FileExists(t, filepath.Join("hello",
		"src", "main", "java", "io", "temporal", "samples", "hello",
		"HelloActivity.java"))

	// README at project root.
	assert.FileExists(t, filepath.Join("hello", "README.md"))

	// root_files copied.
	assert.FileExists(t, filepath.Join("hello", "gradlew"))
	assert.FileExists(t, filepath.Join("hello", "gradlew.bat"))
	assert.FileExists(t, filepath.Join("hello", "gradle", "wrapper", "gradle-wrapper.properties"))
}

// TestSample_Init_DotNet verifies that .NET samples are extracted from src/,
// with a generated Directory.Build.props at the project root.
func TestSample_Init_DotNet(t *testing.T) {
	manifest := `version: 1
language: dotnet
scaffold:
  Directory.Build.props: |
    <Project>
      <PropertyGroup>
        <TargetFramework>net8.0</TargetFramework>
      </PropertyGroup>
      <ItemGroup>
        <PackageReference Include="Temporalio" Version="1.11.1" />
      </ItemGroup>
    </Project>
samples:
  - path: src/ActivitySimple
    description: Simple activity execution from a workflow
`
	tarball := buildGitHubTarball(t, "temporalio-samples-dotnet-abc1234/", []tarEntry{
		{"src/ActivitySimple/TemporalioSamples.ActivitySimple.csproj", "<Project Sdk=\"Microsoft.NET.Sdk\">\n  <PropertyGroup><OutputType>Exe</OutputType></PropertyGroup>\n</Project>\n"},
		{"src/ActivitySimple/Program.cs", "using Temporalio.Client;\n"},
		{"src/ActivitySimple/README.md", "# Activity Simple\n"},
	})

	srv := serveSamples(t, manifest, tarball)
	t.Setenv("TEMPORAL_SAMPLES_BASE_URL", srv.URL)
	t.Chdir(t.TempDir())

	h := NewCommandHarness(t)
	res := h.Execute("sample", "init", "dotnet", "ActivitySimple")

	require.NoError(t, res.Err)

	// Scaffold: Directory.Build.props generated.
	props, err := os.ReadFile(filepath.Join("ActivitySimple", "Directory.Build.props"))
	require.NoError(t, err)
	assert.Contains(t, string(props), "net8.0")

	// Sample files nested under project subdir.
	assert.FileExists(t, filepath.Join("ActivitySimple", "ActivitySimple",
		"TemporalioSamples.ActivitySimple.csproj"))
	assert.FileExists(t, filepath.Join("ActivitySimple", "ActivitySimple", "Program.cs"))

	// README at project root.
	assert.FileExists(t, filepath.Join("ActivitySimple", "README.md"))
}

// TestSample_Init_Ruby verifies that Ruby samples get a generated Gemfile
// at the project root.
func TestSample_Init_Ruby(t *testing.T) {
	manifest := `version: 1
language: ruby
scaffold:
  Gemfile: |
    source 'https://rubygems.org'
    gem 'temporalio'
samples:
  - path: activity_simple
    description: Simple activity execution from a workflow
`
	tarball := buildGitHubTarball(t, "temporalio-samples-ruby-abc1234/", []tarEntry{
		{"activity_simple/my_workflow.rb", "module ActivitySimple; end\n"},
		{"activity_simple/my_activities.rb", "module ActivitySimple; end\n"},
		{"activity_simple/worker.rb", "require_relative 'my_workflow'\n"},
		{"activity_simple/starter.rb", "require_relative 'my_workflow'\n"},
		{"activity_simple/README.md", "# Activity Simple\n"},
	})

	srv := serveSamples(t, manifest, tarball)
	t.Setenv("TEMPORAL_SAMPLES_BASE_URL", srv.URL)
	t.Chdir(t.TempDir())

	h := NewCommandHarness(t)
	res := h.Execute("sample", "init", "ruby", "activity_simple")

	require.NoError(t, res.Err)

	// Scaffold: Gemfile generated.
	gemfile, err := os.ReadFile(filepath.Join("activity_simple", "Gemfile"))
	require.NoError(t, err)
	assert.Contains(t, string(gemfile), "gem 'temporalio'")

	// Sample files present (Ruby uses require_relative, no nesting needed,
	// but scaffold is non-empty so files are nested under sample subdir).
	assert.FileExists(t, filepath.Join("activity_simple", "activity_simple", "worker.rb"))
	assert.FileExists(t, filepath.Join("activity_simple", "activity_simple", "my_workflow.rb"))

	// README at project root.
	assert.FileExists(t, filepath.Join("activity_simple", "README.md"))
}

// TestSample_Init_PathDot verifies that a manifest next to the sample with
// path: "." correctly identifies the enclosing directory as the sample.
func TestSample_Init_PathDot(t *testing.T) {
	sampleManifest := `version: 1
language: python
scaffold:
  pyproject.toml: |
    [project]
    name = "{{name}}"
    dependencies = [{{dependencies}}]
samples:
  - path: .
    description: A sample right here
    dependencies:
      - "temporalio>=1.23.0,<2"
`
	tarball := buildGitHubTarball(t, "dandavison-etc-abc1234/", []tarEntry{
		{"hello-sample/" + manifestFile, sampleManifest},
		{"hello-sample/__init__.py", ""},
		{"hello-sample/worker.py", "import asyncio\n"},
		{"hello-sample/README.md", "# Hello Sample\n"},
	})

	// Serve manifest only at the sample-adjacent path, not at root.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/tar.gz/"):
			w.Header().Set("Content-Type", "application/gzip")
			_, _ = w.Write(tarball)
		case strings.HasSuffix(r.URL.Path, "/hello-sample/"+manifestFile):
			_, _ = w.Write([]byte(sampleManifest))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	t.Setenv("TEMPORAL_SAMPLES_BASE_URL", srv.URL)
	t.Chdir(t.TempDir())

	h := NewCommandHarness(t)
	res := h.Execute("sample", "init",
		"https://github.com/dandavison/etc/tree/temporal-samples/hello-sample")

	require.NoError(t, res.Err)

	// Scaffold: pyproject.toml generated.
	pyproject, err := os.ReadFile(filepath.Join("hello-sample", "pyproject.toml"))
	require.NoError(t, err)
	assert.Contains(t, string(pyproject), `name = "hello-sample"`)
	assert.Contains(t, string(pyproject), `"temporalio>=1.23.0,<2"`)

	// README at project root.
	assert.FileExists(t, filepath.Join("hello-sample", "README.md"))

	// Sample files nested under package dir.
	assert.FileExists(t, filepath.Join("hello-sample", "hello-sample", "__init__.py"))
	assert.FileExists(t, filepath.Join("hello-sample", "hello-sample", "worker.py"))

	// Manifest excluded.
	assert.NoFileExists(t, filepath.Join("hello-sample", manifestFile))
	assert.NoFileExists(t, filepath.Join("hello-sample", "hello-sample", manifestFile))
}

func TestSample_Init_NoManifest(t *testing.T) {
	tarball := buildGitHubTarball(t, "temporalio-samples-python-abc1234/", []tarEntry{
		{"hello/__init__.py", ""},
		{"hello/worker.py", "import asyncio\n"},
		{"hello/README.md", "# Hello Sample\n"},
	})
	// Server returns 404 for all manifest requests.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/tar.gz/"):
			w.Header().Set("Content-Type", "application/gzip")
			_, _ = w.Write(tarball)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	t.Setenv("TEMPORAL_SAMPLES_BASE_URL", srv.URL)
	t.Chdir(t.TempDir())

	h := NewCommandHarness(t)
	res := h.Execute("sample", "init", "python", "hello")

	require.NoError(t, res.Err)
	out := res.Stdout.String()
	assert.Contains(t, out, "Warning")

	// Flat copy: files directly under hello/, no nesting.
	assert.FileExists(t, filepath.Join("hello", "worker.py"))
	assert.FileExists(t, filepath.Join("hello", "README.md"))
	assert.NoDirExists(t, filepath.Join("hello", "hello"))
}

func TestSample_Init_NotFound(t *testing.T) {
	srv := serveSamples(t, testPythonManifest, testPythonTarball(t))
	t.Setenv("TEMPORAL_SAMPLES_BASE_URL", srv.URL)
	t.Chdir(t.TempDir())

	h := NewCommandHarness(t)
	res := h.Execute("sample", "init", "python", "nonexistent")

	require.Error(t, res.Err)
	assert.Contains(t, res.Err.Error(), `sample "nonexistent" not found in temporalio/samples-python`)
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

func TestSample_LanguageAliases(t *testing.T) {
	srv := serveSamples(t, testPythonManifest, testPythonTarball(t))
	t.Setenv("TEMPORAL_SAMPLES_BASE_URL", srv.URL)

	for _, alias := range []string{"py", "ts", "cs", "csharp", "rb"} {
		t.Run(alias, func(t *testing.T) {
			h := NewCommandHarness(t)
			res := h.Execute("sample", "list", alias)
			// All aliases resolve to a valid repo, so the request hits our
			// test server. The server always returns the Python manifest,
			// which is fine — we're testing alias resolution, not content.
			require.NoError(t, res.Err)
		})
	}
}
