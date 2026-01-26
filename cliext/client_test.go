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

func TestClientOptionsBuilder_OptionsPassedThrough(t *testing.T) {
	builder := &cliext.ClientOptionsBuilder{
		CommonOptions: cliext.CommonOptions{
			DisableConfigFile:    true,
			ClientConnectTimeout: cliext.FlagDuration(5 * time.Second),
		},
		ClientOptions: cliext.ClientOptions{
			Address:         "my-custom-host:7233",
			Namespace:       "my-namespace",
			ApiKey:          "my-api-key",
			Identity:        "my-identity",
			ClientAuthority: "my-authority",
			CodecEndpoint:   "http://localhost:8080/codec",
			CodecAuth:       "my-codec-auth",
			GrpcMeta:        []string{"key1=value1", "key2=value2"},
		},
	}
	opts, err := builder.Build(t.Context())

	require.NoError(t, err)
	assert.Equal(t, "my-custom-host:7233", opts.HostPort)
	assert.Equal(t, "my-namespace", opts.Namespace)
	assert.NotNil(t, opts.Credentials)
	assert.Equal(t, "my-identity", opts.Identity)
	assert.Equal(t, "my-authority", opts.ConnectionOptions.Authority)
	assert.Equal(t, 5*time.Second, opts.ConnectionOptions.GetSystemInfoTimeout)
	// CodecEndpoint/CodecAuth result in a gRPC interceptor being added
	assert.NotEmpty(t, opts.ConnectionOptions.DialOptions)
	// GrpcMeta results in a HeadersProvider being set
	assert.NotNil(t, opts.HeadersProvider)
}

func TestClientOptionsBuilder_NamespaceReplacementInAddress(t *testing.T) {
	builder := &cliext.ClientOptionsBuilder{
		CommonOptions: cliext.CommonOptions{
			DisableConfigFile: true,
		},
		ClientOptions: cliext.ClientOptions{
			Address:   "${namespace}.api.temporal.io:7233",
			Namespace: "my-namespace",
		},
	}

	opts, err := builder.Build(t.Context())

	require.NoError(t, err)
	assert.Equal(t, "my-namespace.api.temporal.io:7233", opts.HostPort)
}

func TestClientOptionsBuilder_OAuth_ValidToken(t *testing.T) {
	s := newMockOAuthServer(t)
	configFile := createTestOAuthConfig(t, s.tokenURL, time.Now().Add(time.Hour))

	builder := &cliext.ClientOptionsBuilder{
		CommonOptions: cliext.CommonOptions{
			ConfigFile: configFile,
		},
		ClientOptions: cliext.ClientOptions{
			Address:   "localhost:17233",
			Namespace: "my-namespace",
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

	configFile := createTestOAuthConfig(t, s.tokenURL, time.Now().Add(-time.Hour))

	builder := &cliext.ClientOptionsBuilder{
		CommonOptions: cliext.CommonOptions{
			ConfigFile: configFile,
		},
		ClientOptions: cliext.ClientOptions{
			Address:   "localhost:17233",
			Namespace: "my-namespace",
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

func TestClientOptionsBuilder_OAuth_NoConfigFile(t *testing.T) {
	builder := &cliext.ClientOptionsBuilder{
		CommonOptions: cliext.CommonOptions{
			ConfigFile:        "/non/existent/path.toml",
			DisableConfigFile: true,
		},
		ClientOptions: cliext.ClientOptions{
			Address:   "localhost:17233",
			Namespace: "my-namespace",
		},
	}
	opts, err := builder.Build(t.Context())

	require.NoError(t, err)
	assert.Nil(t, opts.Credentials)
}

func TestClientOptionsBuilder_OAuth_NoOAuthConfigured(t *testing.T) {
	configFile := filepath.Join(t.TempDir(), "config.toml")
	err := os.WriteFile(configFile, []byte{}, 0600) // empty
	require.NoError(t, err)

	builder := &cliext.ClientOptionsBuilder{
		CommonOptions: cliext.CommonOptions{
			ConfigFile: configFile,
		},
		ClientOptions: cliext.ClientOptions{
			Address:   "localhost:17233",
			Namespace: "my-namespace",
		},
	}
	opts, err := builder.Build(t.Context())

	require.NoError(t, err)
	assert.Nil(t, opts.Credentials)
}

func TestClientOptionsBuilder_OAuth_DisableConfigFile(t *testing.T) {
	s := newMockOAuthServer(t)
	configFileWithOAuth := createTestOAuthConfig(t, s.tokenURL, time.Now().Add(time.Hour))

	builder := &cliext.ClientOptionsBuilder{
		CommonOptions: cliext.CommonOptions{
			ConfigFile:        configFileWithOAuth,
			DisableConfigFile: true, // disables loading config file
		},
		ClientOptions: cliext.ClientOptions{
			Address:   "localhost:17233",
			Namespace: "my-namespace",
		},
	}
	opts, err := builder.Build(t.Context())

	require.NoError(t, err)
	assert.Nil(t, opts.Credentials)
}

func TestClientOptionsBuilder_OAuth_APIKeyTakesPrecedence(t *testing.T) {
	s := newMockOAuthServer(t)
	configFileWithOAuth := createTestOAuthConfig(t, s.tokenURL, time.Now().Add(time.Hour))

	builder := &cliext.ClientOptionsBuilder{
		CommonOptions: cliext.CommonOptions{
			ConfigFile: configFileWithOAuth,
		},
		ClientOptions: cliext.ClientOptions{
			Address:   "localhost:17233",
			Namespace: "my-namespace",
			ApiKey:    "explicit-api-key", // takes precedence
		},
	}
	opts, err := builder.Build(t.Context())

	require.NoError(t, err)
	assert.NotNil(t, opts.Credentials)
}

func TestClientOptionsBuilder_OAuth_NonDefaultNamespaceRequired(t *testing.T) {
	configFileWithOAuth := createTestOAuthConfig(t, "", time.Now().Add(time.Hour))

	builder := &cliext.ClientOptionsBuilder{
		CommonOptions: cliext.CommonOptions{
			ConfigFile: configFileWithOAuth,
		},
		ClientOptions: cliext.ClientOptions{
			Address:   "${namespace}.api.temporal.io:7233",
			Namespace: "default",
		},
	}

	_, err := builder.Build(t.Context())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "namespace is required")
}

func createTestOAuthConfig(t *testing.T, tokenURL string, expiry time.Time) string {
	t.Helper()
	configFile := filepath.Join(t.TempDir(), "config.toml")
	err := cliext.StoreClientOAuth(cliext.StoreClientOAuthOptions{
		ConfigFilePath: configFile,
		OAuth: &cliext.OAuthConfig{
			ClientConfig: &oauth2.Config{
				ClientID:     "test-client",
				ClientSecret: "test-secret",
				Endpoint: oauth2.Endpoint{
					TokenURL: tokenURL,
				},
			},
			Token: &oauth2.Token{
				AccessToken:  "test-access-token",
				RefreshToken: "test-refresh-token",
				Expiry:       expiry,
			},
		},
	})
	require.NoError(t, err)
	return configFile
}
