package cliext_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/cli/cliext"
	"go.temporal.io/sdk/client"
	"golang.org/x/oauth2"
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

func TestClientOptionsBuilder_OAuth_ValidToken(t *testing.T) {
	s := newMockOAuthServer(t)

	configFile := filepath.Join(t.TempDir(), "config.toml")

	err := cliext.StoreClientOAuth(cliext.StoreClientOAuthOptions{
		ConfigFilePath: configFile,
		OAuth: &cliext.OAuthConfig{
			ClientConfig: &oauth2.Config{
				ClientID:     "test-client",
				ClientSecret: "test-secret",
				Endpoint: oauth2.Endpoint{
					TokenURL: s.tokenURL,
				},
			},
			Token: &oauth2.Token{
				AccessToken:  "test-access-token",
				RefreshToken: "test-refresh-token",
				Expiry:       time.Now().Add(time.Hour), // not expired
			},
		},
	})
	require.NoError(t, err)

	builder := &cliext.ClientOptionsBuilder{
		CommonOptions: cliext.CommonOptions{
			ConfigFile: configFile,
		},
		ClientOptions: cliext.ClientOptions{
			Address:   "localhost:7233",
			Namespace: "default",
		},
	}
	opts, err := builder.Build(t.Context())

	require.NoError(t, err)
	assert.NotNil(t, opts.Credentials)
}

func TestClientOptionsBuilder_OAuth_Refresh(t *testing.T) {
	s := newMockOAuthServer(t)
	s.TokenRefreshHandler = func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"access_token":"refreshed-token","refresh_token":"new-refresh-token","expires_in":3600,"token_type":"Bearer"}`)
	}

	configFile := filepath.Join(t.TempDir(), "config.toml")

	err := cliext.StoreClientOAuth(cliext.StoreClientOAuthOptions{
		ConfigFilePath: configFile,
		OAuth: &cliext.OAuthConfig{
			ClientConfig: &oauth2.Config{
				ClientID:     "test-client",
				ClientSecret: "test-secret",
				Endpoint: oauth2.Endpoint{
					TokenURL: s.tokenURL,
				},
			},
			Token: &oauth2.Token{
				AccessToken:  "test-access-token",
				RefreshToken: "test-refresh-token",
				Expiry:       time.Now().Add(-time.Hour), // expired
			},
		},
	})
	require.NoError(t, err)

	builder := &cliext.ClientOptionsBuilder{
		CommonOptions: cliext.CommonOptions{
			ConfigFile: configFile,
		},
		ClientOptions: cliext.ClientOptions{
			Address:   "localhost:7233",
			Namespace: "default",
		},
	}
	opts, err := builder.Build(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, opts.Credentials)

	// Create a client and attempt to dial to force the token refresh.
	// This will fail to connect since there's no server, but it will trigger
	// the OAuth token refresh and persistence.
	_, _ = client.Dial(opts)

	// Verify that the refreshed token was persisted to the config file
	result, err := cliext.LoadClientOAuth(cliext.LoadClientOAuthOptions{
		ConfigFilePath: configFile,
	})
	require.NoError(t, err)
	require.NotNil(t, result.OAuth)
	assert.Equal(t, "refreshed-token", result.OAuth.Token.AccessToken)
	assert.Equal(t, "new-refresh-token", result.OAuth.Token.RefreshToken)
}

func TestClientOptionsBuilder_OAuth_APIKeyTakesPrecedence(t *testing.T) {
	s := newMockOAuthServer(t)

	configFile := filepath.Join(t.TempDir(), "config.toml")

	err := cliext.StoreClientOAuth(cliext.StoreClientOAuthOptions{
		ConfigFilePath: configFile,
		OAuth: &cliext.OAuthConfig{
			ClientConfig: &oauth2.Config{
				ClientID:     "test-client",
				ClientSecret: "test-secret",
				Endpoint: oauth2.Endpoint{
					TokenURL: s.tokenURL,
				},
			},
			Token: &oauth2.Token{
				AccessToken:  "test-access-token",
				RefreshToken: "test-refresh-token",
				Expiry:       time.Now().Add(time.Hour),
			},
		},
	})
	require.NoError(t, err)

	// When API key is set, OAuth should not be used
	builder := &cliext.ClientOptionsBuilder{
		CommonOptions: cliext.CommonOptions{
			ConfigFile: configFile,
		},
		ClientOptions: cliext.ClientOptions{
			Address:   "localhost:7233",
			Namespace: "default",
			ApiKey:    "explicit-api-key",
		},
	}
	opts, err := builder.Build(t.Context())

	require.NoError(t, err)
	// The API key credentials should be used, not OAuth
	assert.NotNil(t, opts.Credentials)
}

func TestClientOptionsBuilder_OAuth_NoOAuth(t *testing.T) {
	configFile := filepath.Join(t.TempDir(), "config.toml")
	err := os.WriteFile(configFile, []byte("[profile.default]\naddress = \"localhost:7233\"\n"), 0600)
	require.NoError(t, err)

	builder := &cliext.ClientOptionsBuilder{
		CommonOptions: cliext.CommonOptions{
			ConfigFile: configFile,
		},
		ClientOptions: cliext.ClientOptions{
			Address:   "localhost:7233",
			Namespace: "default",
		},
	}
	opts, err := builder.Build(t.Context())

	require.NoError(t, err)
	// When no OAuth is configured, credentials should be nil
	assert.Nil(t, opts.Credentials)
}

func TestClientOptionsBuilder_OAuth_NoConfigFile(t *testing.T) {
	// Test with a non-existent config file
	builder := &cliext.ClientOptionsBuilder{
		CommonOptions: cliext.CommonOptions{
			ConfigFile:        "/non/existent/path.toml",
			DisableConfigFile: true,
		},
		ClientOptions: cliext.ClientOptions{
			Address:   "localhost:7233",
			Namespace: "default",
		},
	}
	opts, err := builder.Build(t.Context())

	require.NoError(t, err)
	// When config file is disabled, credentials should be nil (no OAuth configured)
	assert.Nil(t, opts.Credentials)
}
