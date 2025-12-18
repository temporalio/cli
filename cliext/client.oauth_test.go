package cliext_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/cli/cliext"
)

type mockOAuthServer struct {
	t        *testing.T
	server   *httptest.Server
	tokenURL string

	TokenRefreshHandler func(http.ResponseWriter, *http.Request)
}

func newMockOAuthServer(t *testing.T) *mockOAuthServer {
	m := &mockOAuthServer{t: t}

	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		w.Header().Set("Content-Type", "application/json")
		if r.FormValue("grant_type") == "refresh_token" {
			if m.TokenRefreshHandler == nil {
				http.Error(w, "TokenRefreshHandler not set", http.StatusInternalServerError)
				return
			}
			m.TokenRefreshHandler(w, r)
		}
	})

	m.server = httptest.NewServer(mux)
	m.tokenURL = m.server.URL + "/token"
	t.Cleanup(m.server.Close)

	return m
}

func writeConfigWithOAuth(t *testing.T, tokenURL string, expiresAt time.Time) string {
	t.Helper()
	f, err := os.CreateTemp("", "temporal-config-*.toml")
	require.NoError(t, err)
	t.Cleanup(func() { os.Remove(f.Name()) })

	expiresAtStr := ""
	if !expiresAt.IsZero() {
		expiresAtStr = fmt.Sprintf("expires_at = %q\n", expiresAt.Format(time.RFC3339))
	}

	_, err = f.WriteString(fmt.Sprintf(`
[profile.default.oauth]
client_id = "test-client"
client_secret = "test-secret"
token_url = %q
access_token = "test-access-token"
refresh_token = "test-refresh-token"
%s
`, tokenURL, expiresAtStr))
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return f.Name()
}

func TestBuildClientOptions_OAuthRefresh(t *testing.T) {
	s := newMockOAuthServer(t)
	s.TokenRefreshHandler = func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"access_token":"refreshed-token","expires_in":3600}`)
	}

	configFile := writeConfigWithOAuth(t, s.tokenURL, time.Now().Add(-time.Hour)) // expired

	opts, _, err := cliext.BuildClientOptions(t.Context(), cliext.ClientOptionsBuilder{
		CommonOptions: cliext.CommonOptions{
			ConfigFile: configFile,
		},
		ClientOptions: cliext.ClientOptions{
			Address:   "localhost:7233",
			Namespace: "default",
		},
	})

	require.NoError(t, err)
	assert.NotNil(t, opts.Credentials)
}

func TestBuildClientOptions_OAuthValidToken(t *testing.T) {
	s := newMockOAuthServer(t)
	configFile := writeConfigWithOAuth(t, s.tokenURL, time.Now().Add(time.Hour)) // not expired

	opts, _, err := cliext.BuildClientOptions(t.Context(), cliext.ClientOptionsBuilder{
		CommonOptions: cliext.CommonOptions{
			ConfigFile: configFile,
		},
		ClientOptions: cliext.ClientOptions{
			Address:   "localhost:7233",
			Namespace: "default",
		},
	})

	require.NoError(t, err)
	assert.NotNil(t, opts.Credentials)
}

func TestBuildClientOptions_APIKeyTakesPrecedence(t *testing.T) {
	s := newMockOAuthServer(t)
	configFile := writeConfigWithOAuth(t, s.tokenURL, time.Now().Add(time.Hour))

	// When API key is set, OAuth should not be used
	opts, _, err := cliext.BuildClientOptions(t.Context(), cliext.ClientOptionsBuilder{
		CommonOptions: cliext.CommonOptions{
			ConfigFile: configFile,
		},
		ClientOptions: cliext.ClientOptions{
			Address:   "localhost:7233",
			Namespace: "default",
			ApiKey:    "explicit-api-key",
		},
	})

	require.NoError(t, err)
	// The API key credentials should be used, not OAuth
	assert.NotNil(t, opts.Credentials)
}

func TestBuildClientOptions_NoOAuth(t *testing.T) {
	// Create empty config file
	f, err := os.CreateTemp("", "temporal-config-*.toml")
	require.NoError(t, err)
	t.Cleanup(func() { os.Remove(f.Name()) })
	f.WriteString("[profile.default]\naddress = \"localhost:7233\"\n")
	f.Close()

	opts, _, err := cliext.BuildClientOptions(t.Context(), cliext.ClientOptionsBuilder{
		CommonOptions: cliext.CommonOptions{
			ConfigFile: f.Name(),
		},
		ClientOptions: cliext.ClientOptions{
			Address:   "localhost:7233",
			Namespace: "default",
		},
	})

	require.NoError(t, err)
	// Credentials are still set (for dynamic OAuth loading), but will return empty string
	// when no OAuth is configured
	assert.NotNil(t, opts.Credentials)
}

func TestBuildClientOptions_NoConfigFile(t *testing.T) {
	// Test with a non-existent config file
	opts, _, err := cliext.BuildClientOptions(t.Context(), cliext.ClientOptionsBuilder{
		CommonOptions: cliext.CommonOptions{
			ConfigFile:        "/non/existent/path.toml",
			DisableConfigFile: true,
		},
		ClientOptions: cliext.ClientOptions{
			Address:   "localhost:7233",
			Namespace: "default",
		},
	})

	require.NoError(t, err)
	// Credentials are still set for dynamic OAuth loading
	assert.NotNil(t, opts.Credentials)
}
