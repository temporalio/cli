//go:build sample_integration

package temporalcli_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func samplesRef() string {
	if v := os.Getenv("TEMPORAL_SAMPLES_REF"); v != "" {
		return v
	}
	return "cli-sample"
}

func initSampleURL(t *testing.T, url string) *CommandResult {
	t.Helper()
	t.Chdir(t.TempDir())
	h := NewCommandHarness(t)
	return h.Execute("sample", "init", "--url", url)
}

func listSamples(t *testing.T, lang string) *CommandResult {
	t.Helper()
	h := NewCommandHarness(t)
	return h.Execute("sample", "list", "--language", lang)
}

func runCmd(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_PAGER=cat")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "command %s %v failed in %s:\n%s", name, args, dir, string(out))
}

// --- Init tests ---

func TestSampleIntegration_Init_Python(t *testing.T) {
	ref := samplesRef()
	res := initSampleURL(t,
		"https://github.com/temporalio/samples-python/tree/"+ref+"/hello")
	require.NoError(t, res.Err)

	pyproject, err := os.ReadFile(filepath.Join("hello", "pyproject.toml"))
	require.NoError(t, err)
	assert.Contains(t, string(pyproject), `name = "hello"`)
	assert.Contains(t, string(pyproject), "temporalio")

	assert.FileExists(t, filepath.Join("hello", "README.md"))
	assert.FileExists(t, filepath.Join("hello", "hello", "__init__.py"))

	runCmd(t, "hello", "uv", "sync")
}

func TestSampleIntegration_Init_Go(t *testing.T) {
	ref := samplesRef()
	res := initSampleURL(t,
		"https://github.com/temporalio/samples-go/tree/"+ref+"/helloworld")
	require.NoError(t, res.Err)

	gomod, err := os.ReadFile(filepath.Join("helloworld", "go.mod"))
	require.NoError(t, err)
	assert.Contains(t, string(gomod), "module helloworld")

	// Check import rewriting in worker.
	workerPath := filepath.Join("helloworld", "helloworld", "worker", "main.go")
	if _, err := os.Stat(workerPath); err == nil {
		worker, err := os.ReadFile(workerPath)
		require.NoError(t, err)
		assert.Contains(t, string(worker), `"helloworld/helloworld"`)
		assert.NotContains(t, string(worker), "github.com/temporalio/samples-go")
	}

	runCmd(t, "helloworld", "go", "mod", "tidy")
	runCmd(t, "helloworld", "go", "build", "./...")
}

func TestSampleIntegration_Init_TypeScript(t *testing.T) {
	ref := samplesRef()
	res := initSampleURL(t,
		"https://github.com/temporalio/samples-typescript/tree/"+ref+"/hello-world")
	require.NoError(t, res.Err)

	assert.FileExists(t, filepath.Join("hello-world", "package.json"))
	assert.FileExists(t, filepath.Join("hello-world", "src", "workflows.ts"))
	assert.NoDirExists(t, filepath.Join("hello-world", "hello-world"))

	pkg, err := os.ReadFile(filepath.Join("hello-world", "package.json"))
	require.NoError(t, err)
	assert.Contains(t, string(pkg), "@temporalio")

	runCmd(t, "hello-world", "npm", "install")
	runCmd(t, "hello-world", "npx", "tsc", "--noEmit")
}

func TestSampleIntegration_Init_Java(t *testing.T) {
	ref := samplesRef()
	res := initSampleURL(t,
		"https://github.com/temporalio/samples-java/tree/"+ref+"/hello")
	require.NoError(t, res.Err)

	gradle, err := os.ReadFile(filepath.Join("hello", "build.gradle"))
	require.NoError(t, err)
	assert.Contains(t, string(gradle), "temporal-sdk")

	// Java files at deep path.
	javaDir := filepath.Join("hello", "src", "main", "java", "io", "temporal", "samples", "hello")
	entries, err := os.ReadDir(javaDir)
	require.NoError(t, err)
	javaFound := false
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".java") {
			javaFound = true
			break
		}
	}
	assert.True(t, javaFound, "expected .java files in %s", javaDir)

	runCmd(t, "hello", "gradle", "compileJava")
}

func TestSampleIntegration_Init_DotNet(t *testing.T) {
	ref := samplesRef()
	res := initSampleURL(t,
		"https://github.com/temporalio/samples-dotnet/tree/"+ref+"/ActivitySimple")
	require.NoError(t, res.Err)

	props, err := os.ReadFile(filepath.Join("ActivitySimple", "Directory.Build.props"))
	require.NoError(t, err)
	assert.Contains(t, string(props), "net8.0")
	assert.Contains(t, string(props), "Temporalio")

	// Check for .csproj file.
	csprojDir := filepath.Join("ActivitySimple", "ActivitySimple")
	entries, err := os.ReadDir(csprojDir)
	require.NoError(t, err)
	csprojFound := false
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".csproj") {
			csprojFound = true
			break
		}
	}
	assert.True(t, csprojFound, "expected .csproj file in %s", csprojDir)

	runCmd(t, filepath.Join("ActivitySimple", "ActivitySimple"), "dotnet", "build")
}

func TestSampleIntegration_Init_Ruby(t *testing.T) {
	ref := samplesRef()
	res := initSampleURL(t,
		"https://github.com/temporalio/samples-ruby/tree/"+ref+"/activity_simple")
	require.NoError(t, res.Err)

	gemfile, err := os.ReadFile(filepath.Join("activity_simple", "Gemfile"))
	require.NoError(t, err)
	assert.Contains(t, string(gemfile), "temporalio")

	assert.FileExists(t, filepath.Join("activity_simple", "activity_simple", "worker.rb"))

	runCmd(t, "activity_simple", "bundle", "install")
}

// TestSampleIntegration_Init_PathDot verifies that a manifest next to the
// sample with path: "." works against a real third-party repo.
func TestSampleIntegration_Init_PathDot(t *testing.T) {
	res := initSampleURL(t,
		"https://github.com/dandavison/etc/tree/temporal-samples/temporal-sample-test")
	require.NoError(t, res.Err)

	// Scaffold: pyproject.toml generated.
	pyproject, err := os.ReadFile(filepath.Join("temporal-sample-test", "pyproject.toml"))
	require.NoError(t, err)
	assert.Contains(t, string(pyproject), `name = "temporal-sample-test"`)
	assert.Contains(t, string(pyproject), "temporalio")

	// Sample files nested.
	assert.FileExists(t, filepath.Join("temporal-sample-test", "temporal-sample-test", "worker.py"))
	assert.FileExists(t, filepath.Join("temporal-sample-test", "README.md"))

	runCmd(t, "temporal-sample-test", "uv", "sync")
}

// --- List tests ---

func TestSampleIntegration_List_Python(t *testing.T) {
	res := listSamples(t, "python")
	require.NoError(t, res.Err)
	out := res.Stdout.String()
	assert.Contains(t, out, "hello")
	assert.Contains(t, out, "encryption")
	assert.Contains(t, out, "dsl")
}

func TestSampleIntegration_List_Go(t *testing.T) {
	res := listSamples(t, "go")
	require.NoError(t, res.Err)
	out := res.Stdout.String()
	assert.Contains(t, out, "helloworld")
	assert.Contains(t, out, "saga")
	assert.Contains(t, out, "encryption")
}

func TestSampleIntegration_List_TypeScript(t *testing.T) {
	res := listSamples(t, "typescript")
	require.NoError(t, res.Err)
	out := res.Stdout.String()
	assert.Contains(t, out, "hello-world")
	assert.Contains(t, out, "saga")
	assert.Contains(t, out, "encryption")
}

func TestSampleIntegration_List_Java(t *testing.T) {
	res := listSamples(t, "java")
	require.NoError(t, res.Err)
	out := res.Stdout.String()
	assert.Contains(t, out, "hello")
	assert.Contains(t, out, "bookingsaga")
	assert.Contains(t, out, "encryptedpayloads")
}

func TestSampleIntegration_List_DotNet(t *testing.T) {
	res := listSamples(t, "dotnet")
	require.NoError(t, res.Err)
	out := res.Stdout.String()
	assert.Contains(t, out, "ActivitySimple")
	assert.Contains(t, out, "Saga")
	assert.Contains(t, out, "Encryption")
}

func TestSampleIntegration_List_Ruby(t *testing.T) {
	res := listSamples(t, "ruby")
	require.NoError(t, res.Err)
	out := res.Stdout.String()
	assert.Contains(t, out, "activity_simple")
	assert.Contains(t, out, "polling")
}
