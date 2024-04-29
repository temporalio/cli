package temporalcli_test

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestEnv_Simple(t *testing.T) {
	h := NewCommandHarness(t)
	defer h.Close()

	// Non-existent file, no env found for get
	h.Options.EnvConfigFile = "does-not-exist"
	res := h.Execute("env", "get", "--env", "myenv1")
	h.ErrorContains(res.Err, `env "myenv1" not found`)

	// Temp file for env
	tmpFile, err := os.CreateTemp("", "")
	h.NoError(err)
	h.Options.EnvConfigFile = tmpFile.Name()
	defer os.Remove(h.Options.EnvConfigFile)

	// Store a key
	res = h.Execute("env", "set", "--env", "myenv1", "-k", "foo", "-v", "bar")
	h.NoError(res.Err)
	// Confirm file is YAML with expected values
	b, err := os.ReadFile(h.Options.EnvConfigFile)
	h.NoError(err)
	var yamlVals map[string]map[string]map[string]string
	h.NoError(yaml.Unmarshal(b, &yamlVals))
	h.Equal("bar", yamlVals["env"]["myenv1"]["foo"])

	// Store another key and another env
	res = h.Execute("env", "set", "--env", "myenv1", "-k", "baz", "-v", "qux")
	h.NoError(res.Err)
	res = h.Execute("env", "set", "--env", "myenv2", "-k", "foo", "-v", "baz")
	h.NoError(res.Err)

	// Get single prop
	res = h.Execute("env", "get", "--env", "myenv1", "-k", "baz")
	h.NoError(res.Err)
	h.ContainsOnSameLine(res.Stdout.String(), "baz", "qux")
	h.NotContains(res.Stdout.String(), "foo")

	// Get all props for env
	res = h.Execute("env", "get", "--env", "myenv1")
	h.NoError(res.Err)
	h.ContainsOnSameLine(res.Stdout.String(), "foo", "bar")
	h.ContainsOnSameLine(res.Stdout.String(), "baz", "qux")

	// List envs
	res = h.Execute("env", "list")
	h.NoError(res.Err)
	h.Contains(res.Stdout.String(), "myenv1")
	h.Contains(res.Stdout.String(), "myenv2")

	// Delete single env value
	res = h.Execute("env", "delete", "--env", "myenv1", "-k", "foo")
	h.NoError(res.Err)
	res = h.Execute("env", "get", "myenv1")
	h.NoError(res.Err)
	h.NotContains(res.Stdout.String(), "foo")

	// Delete entire env
	res = h.Execute("env", "delete", "myenv2")
	h.NoError(res.Err)
	res = h.Execute("env", "list")
	h.NoError(res.Err)
	h.NotContains(res.Stdout.String(), "myenv2")

	// Ensure env var overrides env file
	res = h.Execute("env", "set", "--env", "myenv1", "-k", "address", "-v", "something:1234")
	h.NoError(res.Err)
	h.NoError(os.Setenv("TEMPORAL_ADDRESS", "overridden:1235"))
	defer os.Unsetenv("TEMPORAL_ADDRESS")
	res = h.Execute("workflow", "list", "--env", "myenv1")
	h.Contains(res.Stderr.String(), "Env var overrode --env setting")
}

func TestEnv_InputValidation(t *testing.T) {
	h := NewCommandHarness(t)
	defer h.Close()

	res := h.Execute("env", "get", "--env", "myenv1", "foo.bar")
	h.ErrorContains(res.Err, `cannot specify both`)

	res = h.Execute("env", "get", "-k", "key", "foo.bar")
	h.ErrorContains(res.Err, `cannot specify both`)

	res = h.Execute("env", "get", "--env", "myenv1", "-k", "foo.bar")
	h.ErrorContains(res.Err, `property name may not contain dots`)

	res = h.Execute("env", "set", "--env", "myenv1", "-k", "foo.bar", "-v", "")
	h.ErrorContains(res.Err, `property name may not contain dots`)

	res = h.Execute("env", "set", "--env", "myenv1", "-k", "", "-v", "")
	h.ErrorContains(res.Err, `property name must be specified`)

	res = h.Execute("env", "set", "myenv1")
	h.ErrorContains(res.Err, `property name must be specified`)

	res = h.Execute("env", "set", "myenv1.foo")
	h.ErrorContains(res.Err, `no value provided`)
}
