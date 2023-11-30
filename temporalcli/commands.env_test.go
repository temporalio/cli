package temporalcli_test

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

// TODO(cretz): To test:
// * Env var actually sets CLI arg
// * Get single and all for env
// * Delete single and all for env
// * List envs

func TestEnv_Simple(t *testing.T) {
	h := NewCommandHarness(t)
	defer h.Close()

	// Non-existent file, no env found for get
	h.Options.EnvFile = "does-not-exist"
	res := h.Execute("env", "get", "myenv")
	h.ErrorContains(res.Err, `env "myenv" not found`)

	// Temp file for env
	tmpFile, err := os.CreateTemp("", "")
	h.NoError(err)
	h.Options.EnvFile = tmpFile.Name()
	defer os.Remove(h.Options.EnvFile)

	// Store a key
	res = h.Execute("env", "set", "myenv.foo", "bar")
	h.NoError(res.Err)
	// Confirm file is YAML with expected values
	b, err := os.ReadFile(h.Options.EnvFile)
	h.NoError(err)
	var yamlVals map[string]map[string]map[string]string
	h.NoError(yaml.Unmarshal(b, &yamlVals))
	h.Equal("bar", yamlVals["env"]["myenv"]["foo"])
}
