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

// serveSamples starts an httptest.Server that serves a GitHub tarball.
func serveSamples(t *testing.T, tarball []byte) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/tar.gz/") {
			w.Header().Set("Content-Type", "application/gzip")
			_, _ = w.Write(tarball)
		} else {
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

const testPythonHelloManifest = `description: Basic hello world samples
dependencies:
  - "temporalio>=1.23.0,<2"
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
		{"hello/temporal-sample.yaml", testPythonHelloManifest},
		{"hello/__init__.py", ""},
		{"hello/my_activity.py", "from temporalio import activity\n"},
		{"hello/worker.py", "import asyncio\n"},
		{"hello/README.md", "# Hello Sample\n\nRun with: uv run hello/worker.py\n"},
		{"encryption/temporal-sample.yaml", "description: End-to-end encryption with a custom codec\n"},
		{"encryption/__init__.py", ""},
		{"encryption/worker.py", "# worker\n"},
	})
}

func TestSample_List(t *testing.T) {
	srv := serveSamples(t, testPythonTarball(t))
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

func TestSample_Init_Python(t *testing.T) {
	srv := serveSamples(t, testPythonTarball(t))
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

	assert.Contains(t, res.Stdout.String(), "hello")
}

func TestSample_Init_TypeScript_FlatCopy(t *testing.T) {
	// No scaffold → flat copy.
	tarball := buildGitHubTarball(t, "temporalio-samples-typescript-def5678/", []tarEntry{
		{"hello-world/temporal-sample.yaml", "description: Basic hello world workflow\n"},
		{"hello-world/package.json", `{"name": "hello-world"}`},
		{"hello-world/src/workflows.ts", "export async function greet() { return 'hello'; }\n"},
		{"hello-world/src/activities.ts", "export async function sayHello() { return 'hello'; }\n"},
		{"hello-world/README.md", "# Hello World\n"},
	})

	srv := serveSamples(t, tarball)
	t.Setenv("TEMPORAL_SAMPLES_BASE_URL", srv.URL)
	t.Chdir(t.TempDir())

	h := NewCommandHarness(t)
	res := h.Execute("sample", "init", "typescript", "hello-world")

	require.NoError(t, res.Err)

	assert.FileExists(t, filepath.Join("hello-world", "package.json"))
	assert.FileExists(t, filepath.Join("hello-world", "src", "workflows.ts"))
	assert.FileExists(t, filepath.Join("hello-world", "src", "activities.ts"))
	assert.FileExists(t, filepath.Join("hello-world", "README.md"))
	assert.NoDirExists(t, filepath.Join("hello-world", "hello-world"))
	assert.NoFileExists(t, filepath.Join("hello-world", "temporal-sample.yaml"))
}

func TestSample_Init_GitHubURL(t *testing.T) {
	srv := serveSamples(t, testPythonTarball(t))
	t.Setenv("TEMPORAL_SAMPLES_BASE_URL", srv.URL)
	t.Chdir(t.TempDir())

	h := NewCommandHarness(t)
	res := h.Execute("sample", "init",
		"https://github.com/temporalio/samples-python/tree/main/hello")

	require.NoError(t, res.Err)
	assert.FileExists(t, filepath.Join("hello", "hello", "__init__.py"))
}

func TestSample_Init_GitHubURL_DeepPath(t *testing.T) {
	tarball := buildGitHubTarball(t, "org-repo-abc1234/", []tarEntry{
		{"deep/path/mysample/temporal-sample.yaml", "description: Deep sample\nscaffold:\n  config.yaml: |\n    name: {{name}}\n"},
		{"deep/path/mysample/main.py", "print('hello')\n"},
		{"deep/path/mysample/README.md", "# Deep Sample\n"},
	})
	srv := serveSamples(t, tarball)
	t.Setenv("TEMPORAL_SAMPLES_BASE_URL", srv.URL)
	t.Chdir(t.TempDir())

	h := NewCommandHarness(t)
	res := h.Execute("sample", "init",
		"https://github.com/org/repo/tree/main/deep/path/mysample")

	require.NoError(t, res.Err)

	// Scaffold present → nested.
	assert.FileExists(t, filepath.Join("mysample", "mysample", "main.py"))
	assert.FileExists(t, filepath.Join("mysample", "README.md"))

	config, err := os.ReadFile(filepath.Join("mysample", "config.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(config), "name: mysample")
}

func TestSample_Init_Go_ImportRewrite(t *testing.T) {
	goManifest := `description: Basic hello world workflow
scaffold:
  go.mod: |
    module {{name}}
    go 1.23.0
    require go.temporal.io/sdk v1.41.0
rewrite_imports:
  from: github.com/temporalio/samples-go
  glob: "*.go"
`
	tarball := buildGitHubTarball(t, "temporalio-samples-go-abc1234/", []tarEntry{
		{"helloworld/temporal-sample.yaml", goManifest},
		{"helloworld/helloworld.go", "package helloworld\n\nimport \"go.temporal.io/sdk/workflow\"\n"},
		{"helloworld/worker/main.go", "package main\n\nimport (\n\t\"github.com/temporalio/samples-go/helloworld\"\n)\n"},
		{"helloworld/starter/main.go", "package main\n\nimport (\n\t\"github.com/temporalio/samples-go/helloworld\"\n)\n"},
		{"helloworld/README.md", "# Hello World\n"},
	})

	srv := serveSamples(t, tarball)
	t.Setenv("TEMPORAL_SAMPLES_BASE_URL", srv.URL)
	t.Chdir(t.TempDir())

	h := NewCommandHarness(t)
	res := h.Execute("sample", "init", "go", "helloworld")

	require.NoError(t, res.Err)

	gomod, err := os.ReadFile(filepath.Join("helloworld", "go.mod"))
	require.NoError(t, err)
	assert.Contains(t, string(gomod), "module helloworld")

	worker, err := os.ReadFile(filepath.Join("helloworld", "helloworld", "worker", "main.go"))
	require.NoError(t, err)
	assert.Contains(t, string(worker), `"helloworld/helloworld"`)
	assert.NotContains(t, string(worker), "github.com/temporalio/samples-go")
}

func TestSample_Init_Java(t *testing.T) {
	javaManifest := `description: Basic hello world samples
sdk_version: "1.32.1"
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
dest_prefix: src/main/java/io/temporal/samples
root_files:
  - gradlew
  - gradle/
`
	tarball := buildGitHubTarball(t, "temporalio-samples-java-abc1234/", []tarEntry{
		{"core/src/main/java/io/temporal/samples/hello/temporal-sample.yaml", javaManifest},
		{"core/src/main/java/io/temporal/samples/hello/HelloActivity.java", "package io.temporal.samples.hello;\npublic class HelloActivity {}\n"},
		{"core/src/main/java/io/temporal/samples/hello/README.md", "# Hello\n"},
		{"gradlew", "#!/bin/sh\nexec gradle \"$@\"\n"},
		{"gradle/wrapper/gradle-wrapper.properties", "distributionUrl=https://services.gradle.org/distributions/gradle-8.10.2-bin.zip\n"},
	})

	srv := serveSamples(t, tarball)
	t.Setenv("TEMPORAL_SAMPLES_BASE_URL", srv.URL)
	t.Chdir(t.TempDir())

	h := NewCommandHarness(t)
	res := h.Execute("sample", "init", "java", "hello")

	require.NoError(t, res.Err)

	// Scaffold: build.gradle with sdk_version expanded.
	gradle, err := os.ReadFile(filepath.Join("hello", "build.gradle"))
	require.NoError(t, err)
	assert.Contains(t, string(gradle), `temporal-sdk:1.32.1`)

	// Java source at deep path under project root.
	assert.FileExists(t, filepath.Join("hello",
		"src", "main", "java", "io", "temporal", "samples", "hello",
		"HelloActivity.java"))

	// README at project root.
	assert.FileExists(t, filepath.Join("hello", "README.md"))

	// Root files copied.
	assert.FileExists(t, filepath.Join("hello", "gradlew"))
	assert.FileExists(t, filepath.Join("hello", "gradle", "wrapper", "gradle-wrapper.properties"))
}

func TestSample_Init_DotNet(t *testing.T) {
	dotnetManifest := `description: Simple activity execution from a workflow
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
`
	tarball := buildGitHubTarball(t, "temporalio-samples-dotnet-abc1234/", []tarEntry{
		{"src/ActivitySimple/temporal-sample.yaml", dotnetManifest},
		{"src/ActivitySimple/TemporalioSamples.ActivitySimple.csproj", "<Project Sdk=\"Microsoft.NET.Sdk\">\n  <PropertyGroup><OutputType>Exe</OutputType></PropertyGroup>\n</Project>\n"},
		{"src/ActivitySimple/Program.cs", "using Temporalio.Client;\n"},
		{"src/ActivitySimple/README.md", "# Activity Simple\n"},
	})

	srv := serveSamples(t, tarball)
	t.Setenv("TEMPORAL_SAMPLES_BASE_URL", srv.URL)
	t.Chdir(t.TempDir())

	h := NewCommandHarness(t)
	res := h.Execute("sample", "init", "dotnet", "ActivitySimple")

	require.NoError(t, res.Err)

	props, err := os.ReadFile(filepath.Join("ActivitySimple", "Directory.Build.props"))
	require.NoError(t, err)
	assert.Contains(t, string(props), "net8.0")

	assert.FileExists(t, filepath.Join("ActivitySimple", "ActivitySimple",
		"TemporalioSamples.ActivitySimple.csproj"))
	assert.FileExists(t, filepath.Join("ActivitySimple", "ActivitySimple", "Program.cs"))
	assert.FileExists(t, filepath.Join("ActivitySimple", "README.md"))
}

func TestSample_Init_Ruby(t *testing.T) {
	rubyManifest := `description: Simple activity execution from a workflow
scaffold:
  Gemfile: |
    source 'https://rubygems.org'
    gem 'temporalio'
`
	tarball := buildGitHubTarball(t, "temporalio-samples-ruby-abc1234/", []tarEntry{
		{"activity_simple/temporal-sample.yaml", rubyManifest},
		{"activity_simple/my_workflow.rb", "module ActivitySimple; end\n"},
		{"activity_simple/my_activities.rb", "module ActivitySimple; end\n"},
		{"activity_simple/worker.rb", "require_relative 'my_workflow'\n"},
		{"activity_simple/starter.rb", "require_relative 'my_workflow'\n"},
		{"activity_simple/README.md", "# Activity Simple\n"},
	})

	srv := serveSamples(t, tarball)
	t.Setenv("TEMPORAL_SAMPLES_BASE_URL", srv.URL)
	t.Chdir(t.TempDir())

	h := NewCommandHarness(t)
	res := h.Execute("sample", "init", "ruby", "activity_simple")

	require.NoError(t, res.Err)

	gemfile, err := os.ReadFile(filepath.Join("activity_simple", "Gemfile"))
	require.NoError(t, err)
	assert.Contains(t, string(gemfile), "gem 'temporalio'")

	assert.FileExists(t, filepath.Join("activity_simple", "activity_simple", "worker.rb"))
	assert.FileExists(t, filepath.Join("activity_simple", "activity_simple", "my_workflow.rb"))
	assert.FileExists(t, filepath.Join("activity_simple", "README.md"))
}

// TestSample_Init_NoManifest_URL verifies that the URL form works even when
// the sample has no temporal-sample.yaml — files are extracted flat.
func TestSample_Init_NoManifest_URL(t *testing.T) {
	tarball := buildGitHubTarball(t, "org-repo-abc1234/", []tarEntry{
		{"some/path/hello/__init__.py", ""},
		{"some/path/hello/worker.py", "import asyncio\n"},
		{"some/path/hello/README.md", "# Hello Sample\n"},
	})
	srv := serveSamples(t, tarball)
	t.Setenv("TEMPORAL_SAMPLES_BASE_URL", srv.URL)
	t.Chdir(t.TempDir())

	h := NewCommandHarness(t)
	res := h.Execute("sample", "init",
		"https://github.com/org/repo/tree/main/some/path/hello")

	require.NoError(t, res.Err)

	// Flat copy: files directly under hello/, no nesting.
	assert.FileExists(t, filepath.Join("hello", "worker.py"))
	assert.FileExists(t, filepath.Join("hello", "README.md"))
	assert.NoDirExists(t, filepath.Join("hello", "hello"))
}

func TestSample_Init_NotFound(t *testing.T) {
	srv := serveSamples(t, testPythonTarball(t))
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
