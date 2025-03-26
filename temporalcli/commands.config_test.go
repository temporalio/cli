package temporalcli_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestConfig_Get(t *testing.T) {
	h := NewCommandHarness(t)
	defer h.Close()
	env := EnvLookupMap{}
	h.Options.EnvLookup = env

	// Put some data in temp file
	f, err := os.CreateTemp("", "")
	h.NoError(err)
	defer os.Remove(f.Name())
	_, err = f.Write([]byte(`
[profile.foo]
address = "my-address"
namespace = "my-namespace"
api_key = "my-api-key"
codec = { endpoint = "my-endpoint", auth = "my-auth" }
grpc_meta = { some-heAder1 = "some-value1", some-header2 = "some-value2", some_heaDer3 = "some-value3" }
some_future_key = "some future value not handled"

[profile.foo.tls]
disabled = true
client_cert_path = "my-client-cert-path"
client_cert_data = "my-client-cert-data"
client_key_path = "my-client-key-path"
client_key_data = "my-client-key-data"
server_ca_cert_path = "my-server-ca-cert-path"
server_ca_cert_data = "my-server-ca-cert-data"
# Intentionally absent
# server_name = "my-server-name"
disable_host_verification = true`))
	f.Close()
	h.NoError(err)
	env["TEMPORAL_CONFIG_FILE"] = f.Name()

	// Bad profile default
	res := h.Execute("config", "get")
	h.ErrorContains(res.Err, `profile "default" not found`)

	// Bad profile env var
	env["TEMPORAL_PROFILE"] = "env-prof"
	res = h.Execute("config", "get")
	h.ErrorContains(res.Err, `profile "env-prof" not found`)
	delete(env, "TEMPORAL_PROFILE")

	// Bad profile CLI arg
	res = h.Execute("config", "get", "--profile", "arg-prof")
	h.ErrorContains(res.Err, `profile "arg-prof" not found`)

	// Unknown prop
	env["TEMPORAL_PROFILE"] = "foo"
	res = h.Execute("config", "get", "--prop", "blah")
	h.ErrorContains(res.Err, `unknown property "blah"`)

	// Unknown meta
	res = h.Execute("config", "get", "--prop", "grpc_meta.wrong")
	h.ErrorContains(res.Err, `unknown property "grpc_meta.wrong"`)
	res = h.Execute("config", "get", "--prop", "grpc_meta.some-heAder1")
	h.ErrorContains(res.Err, `unknown property "grpc_meta.some-heAder1"`)

	// All props
	expectedJSON := map[string]any{
		"address":                       "my-address",
		"namespace":                     "my-namespace",
		"api_key":                       "my-api-key",
		"codec.endpoint":                "my-endpoint",
		"codec.auth":                    "my-auth",
		"grpc_meta.some-header1":        "some-value1",
		"tls":                           true,
		"tls.disabled":                  true,
		"tls.client_cert_path":          "my-client-cert-path",
		"tls.client_cert_data":          []byte("my-client-cert-data"),
		"tls.client_key_path":           "my-client-key-path",
		"tls.client_key_data":           []byte("my-client-key-data"),
		"tls.server_ca_cert_path":       "my-server-ca-cert-path",
		"tls.server_ca_cert_data":       []byte("my-server-ca-cert-data"),
		"tls.server_name":               "",
		"tls.disable_host_verification": true,
	}
	expectedNonJSON := make(map[string]any, len(expectedJSON))
	for prop, expectedVal := range expectedJSON {
		if b, ok := expectedVal.([]byte); ok {
			expectedVal = "bytes(" + base64.StdEncoding.EncodeToString(b) + ")"
		} else {
			expectedNonJSON[prop] = expectedVal
		}
	}

	// JSON individual
	for prop, expectedVal := range expectedJSON {
		res = h.Execute("config", "get", "--prop", prop, "-o", "json")
		b, _ := json.Marshal(expectedVal)
		h.JSONEq(fmt.Sprintf(`{"property": %q, "value": %s}`, prop, b), res.Stdout.String())
	}

	// Non-JSON individual
	for prop, expectedVal := range expectedNonJSON {
		res := h.Execute("config", "get", "--prop", prop)
		h.NoError(res.Err)
		h.ContainsOnSameLine(res.Stdout.String(), prop, fmt.Sprintf("%v", expectedVal))
	}

	// JSON all together
	res = h.Execute("config", "get", "-o", "json")
	h.NoError(res.Err)
	h.JSONEq(`{
		"address": "my-address",
		"api_key": "my-api-key",
		"codec": {
			"auth": "my-auth",
			"endpoint": "my-endpoint"
		},
		"grpc_meta": {
			"some-header1": "some-value1",
			"some-header2": "some-value2",
			"some-header3": "some-value3"
		},
		"namespace": "my-namespace",
		"tls": {
			"client_cert_data": "my-client-cert-data",
			"client_cert_path": "my-client-cert-path",
			"client_key_data": "my-client-key-data",
			"client_key_path": "my-client-key-path",
			"disable_host_verification": true,
			"disabled": true,
			"server_ca_cert_data": "my-server-ca-cert-data",
			"server_ca_cert_path": "my-server-ca-cert-path"
		}
	}`, res.Stdout.String())

	// Non-JSON all together
	res = h.Execute("config", "get")
	h.NoError(res.Err)
	for prop, expectedVal := range expectedNonJSON {
		// Server name is excluded because it's a zero val
		if prop == "tls.server_name" {
			continue
		}
		h.ContainsOnSameLine(res.Stdout.String(), prop, fmt.Sprintf("%v", expectedVal))
	}
}

func TestConfig_TLS_Boolean(t *testing.T) {
	h := NewCommandHarness(t)
	defer h.Close()

	// Put some data in temp file
	f, err := os.CreateTemp("", "")
	h.NoError(err)
	defer os.Remove(f.Name())
	_, err = f.Write([]byte(`
[profile.foo]
address = "my-address"

[profile.foo.tls]`))
	f.Close()
	h.NoError(err)
	h.Options.EnvLookup = EnvLookupMap{"TEMPORAL_CONFIG_FILE": f.Name(), "TEMPORAL_PROFILE": "foo"}

	// Check that it shows TLS as true
	res := h.Execute("config", "get", "--prop", "tls")
	h.NoError(res.Err)
	h.ContainsOnSameLine(res.Stdout.String(), "tls", "true")

	// Now set it as false and confirm deleted
	res = h.Execute("config", "set", "--prop", "tls", "--value", "false")
	h.NoError(res.Err)
	b, err := os.ReadFile(f.Name())
	h.NoError(err)
	h.NotContains(string(b), "tls")
	res = h.Execute("config", "get", "--prop", "tls")
	h.NoError(res.Err)
	h.ContainsOnSameLine(res.Stdout.String(), "tls", "false")

	// Set it as true and confirm it is there again
	res = h.Execute("config", "set", "--prop", "tls", "--value", "true")
	h.NoError(res.Err)
	b, err = os.ReadFile(f.Name())
	h.NoError(err)
	h.Contains(string(b), "tls")
	res = h.Execute("config", "get", "--prop", "tls")
	h.NoError(res.Err)
	h.ContainsOnSameLine(res.Stdout.String(), "tls", "true")

	// Delete and confirm gone
	res = h.Execute("config", "delete", "--prop", "tls")
	h.NoError(res.Err)
	b, err = os.ReadFile(f.Name())
	h.NoError(err)
	h.NotContains(string(b), "tls")
	res = h.Execute("config", "get", "--prop", "tls")
	h.NoError(res.Err)
	h.ContainsOnSameLine(res.Stdout.String(), "tls", "false")
}

func TestConfig_Delete(t *testing.T) {
	h := NewCommandHarness(t)
	defer h.Close()

	// Put some data in temp file
	f, err := os.CreateTemp("", "")
	h.NoError(err)
	defer os.Remove(f.Name())
	_, err = f.Write([]byte(`
[profile.foo]
address = "my-address"
namespace = "my-namespace"

[profile.foo.tls]`))
	f.Close()
	h.NoError(err)
	h.Options.EnvLookup = EnvLookupMap{"TEMPORAL_CONFIG_FILE": f.Name(), "TEMPORAL_PROFILE": "foo"}

	// Confirm address and namespace is there
	res := h.Execute("config", "get")
	h.NoError(res.Err)
	h.Contains(res.Stdout.String(), "my-address")
	h.Contains(res.Stdout.String(), "my-namespace")

	// Delete namespace and confirm gone but address still there
	res = h.Execute("config", "delete", "--prop", "namespace")
	h.NoError(res.Err)
	res = h.Execute("config", "get")
	h.NoError(res.Err)
	h.Contains(res.Stdout.String(), "my-address")
	h.NotContains(res.Stdout.String(), "my-namespace")

	// Delete entire profile
	res = h.Execute("config", "delete-profile", "--profile", "foo")
	h.NoError(res.Err)
	res = h.Execute("config", "get")
	h.ErrorContains(res.Err, `profile "foo" not found`)
}

func TestConfig_List(t *testing.T) {
	h := NewCommandHarness(t)
	defer h.Close()

	// Put some data in temp file
	f, err := os.CreateTemp("", "")
	h.NoError(err)
	defer os.Remove(f.Name())
	_, err = f.Write([]byte(`
[profile.foo]
address = "my-address-foo"
[profile.bar]
address = "my-address-bar"`))
	f.Close()
	h.NoError(err)
	h.Options.EnvLookup = EnvLookupMap{"TEMPORAL_CONFIG_FILE": f.Name()}

	// Confirm both profiles are there
	res := h.Execute("config", "list")
	h.NoError(res.Err)
	h.Contains(res.Stdout.String(), "foo")
	h.Contains(res.Stdout.String(), "bar")

	// Same in JSON
	res = h.Execute("config", "list", "-o", "json")
	h.NoError(res.Err)
	h.Contains(res.Stdout.String(), `"foo"`)
	h.Contains(res.Stdout.String(), `"bar"`)

	// Now delete and try again
	res = h.Execute("config", "delete-profile", "--profile", "foo")
	h.NoError(res.Err)
	res = h.Execute("config", "list")
	h.NoError(res.Err)
	h.NotContains(res.Stdout.String(), "foo")
	h.Contains(res.Stdout.String(), "bar")

	// Same in JSON
	res = h.Execute("config", "list", "-o", "json")
	h.NoError(res.Err)
	h.NotContains(res.Stdout.String(), `"foo"`)
	h.Contains(res.Stdout.String(), `"bar"`)
}

func TestConfig_Set(t *testing.T) {
	h := NewCommandHarness(t)
	defer h.Close()

	// Create a temp file then delete it immediately to confirm set lazily
	// creates as needed
	f, err := os.CreateTemp("", "")
	h.NoError(err)
	h.NoError(f.Close())
	h.NoError(os.Remove(f.Name()))
	_, err = os.Stat(f.Name())
	h.True(os.IsNotExist(err))
	// Also remove again at the end
	defer os.Remove(f.Name())
	h.Options.EnvLookup = EnvLookupMap{"TEMPORAL_CONFIG_FILE": f.Name()}

	// Now set an address which will be on default profile and confirm in file
	// and "get"
	res := h.Execute("config", "set", "--prop", "address", "--value", "some-address")
	h.NoError(res.Err)
	b, err := os.ReadFile(f.Name())
	h.NoError(err)
	h.Contains(string(b), "[profile.default]")
	h.Contains(string(b), `"some-address"`)
	res = h.Execute("config", "get", "--prop", "address")
	h.NoError(res.Err)
	h.ContainsOnSameLine(res.Stdout.String(), "address", "some-address")

	// Set a bunch of other things
	toSet := map[string]string{
		"address":                       "my-address",
		"namespace":                     "my-namespace",
		"api_key":                       "my-api-key",
		"codec.endpoint":                "my-endpoint",
		"codec.auth":                    "my-auth",
		"grpc_meta.sOme_header1":        "some-value1",
		"tls":                           "true",
		"tls.disabled":                  "true",
		"tls.client_cert_path":          "my-client-cert-path",
		"tls.client_cert_data":          "my-client-cert-data",
		"tls.client_key_path":           "my-client-key-path",
		"tls.client_key_data":           "my-client-key-data",
		"tls.server_ca_cert_path":       "my-server-ca-cert-path",
		"tls.server_ca_cert_data":       "my-server-ca-cert-data",
		"tls.disable_host_verification": "true",
	}
	for k, v := range toSet {
		res = h.Execute("config", "set", "--prop", k, "--value", v)
		h.NoError(res.Err)
	}

	// TOML parse that whole thing and confirm equals
	b, err = os.ReadFile(f.Name())
	h.NoError(err)
	var all any
	h.NoError(toml.Unmarshal(b, &all))
	h.Equal(
		map[string]any{
			"profile": map[string]any{
				"default": map[string]any{
					"address":   "my-address",
					"namespace": "my-namespace",
					"api_key":   "my-api-key",
					"tls": map[string]any{
						"disabled":                  true,
						"client_cert_path":          "my-client-cert-path",
						"client_cert_data":          "my-client-cert-data",
						"client_key_path":           "my-client-key-path",
						"client_key_data":           "my-client-key-data",
						"server_ca_cert_path":       "my-server-ca-cert-path",
						"server_ca_cert_data":       "my-server-ca-cert-data",
						"disable_host_verification": true,
					},
					"codec": map[string]any{
						"endpoint": "my-endpoint",
						"auth":     "my-auth",
					},
					"grpc_meta": map[string]any{
						"some-header1": "some-value1",
					},
				},
			},
		},
		all)
}
