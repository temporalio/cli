package cliext_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/cli/cliext"
)

func TestLoadClientOAuth(t *testing.T) {
	// Create temp config file with OAuth settings
	f, err := os.CreateTemp("", "temporal-config-*.toml")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	_, err = f.WriteString(`
[profile.default.oauth]
client_id = "test-client"
client_secret = "test-secret"
token_url = "https://example.com/oauth/token"
auth_url = "https://example.com/oauth/authorize"
access_token = "test-access-token"
refresh_token = "test-refresh-token"
token_type = "Bearer"
expires_at = "2318-03-23T00:00:00Z"
scopes = ["openid", "profile", "email"]
request_params = { audience = "https://api.example.com" }
`)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	result, err := cliext.LoadClientOAuth(cliext.LoadClientOAuthOptions{
		ConfigFilePath: f.Name(),
	})
	require.NoError(t, err)
	require.NotNil(t, result.OAuth)

	assert.Equal(t, "test-client", result.OAuth.ClientID)
	assert.Equal(t, "test-secret", result.OAuth.ClientSecret)
	assert.Equal(t, "https://example.com/oauth/token", result.OAuth.TokenURL)
	assert.Equal(t, "https://example.com/oauth/authorize", result.OAuth.AuthURL)
	assert.Equal(t, "test-access-token", result.OAuth.AccessToken)
	assert.Equal(t, "test-refresh-token", result.OAuth.RefreshToken)
	assert.Equal(t, "Bearer", result.OAuth.TokenType)
	assert.Equal(t, time.Date(2318, 3, 23, 0, 0, 0, 0, time.UTC), result.OAuth.AccessTokenExpiresAt)
	assert.Equal(t, []string{"openid", "profile", "email"}, result.OAuth.Scopes)
	assert.Equal(t, map[string]string{"audience": "https://api.example.com"}, result.OAuth.RequestParams)
}

func TestLoadClientOAuth_DifferentProfile(t *testing.T) {
	f, err := os.CreateTemp("", "temporal-config-*.toml")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	_, err = f.WriteString(`
[profile.custom.oauth]
client_id = "custom-client"
access_token = "custom-token"
refresh_token = "custom-refresh"
`)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	result, err := cliext.LoadClientOAuth(cliext.LoadClientOAuthOptions{
		ConfigFilePath: f.Name(),
		ProfileName:    "custom",
	})
	require.NoError(t, err)
	require.NotNil(t, result.OAuth)

	assert.Equal(t, "custom-client", result.OAuth.ClientID)
	assert.Equal(t, "custom-token", result.OAuth.AccessToken)
	assert.Equal(t, "custom-refresh", result.OAuth.RefreshToken)
	assert.Equal(t, "custom", result.ProfileName)
}

func TestLoadClientOAuth_NoOAuth(t *testing.T) {
	f, err := os.CreateTemp("", "temporal-config-*.toml")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	_, err = f.WriteString(`
[profile.default]
address = "localhost:7233"
`)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	result, err := cliext.LoadClientOAuth(cliext.LoadClientOAuthOptions{
		ConfigFilePath: f.Name(),
	})
	require.NoError(t, err)
	assert.Nil(t, result.OAuth)
}

func TestLoadClientOAuth_FileNotFound(t *testing.T) {
	result, err := cliext.LoadClientOAuth(cliext.LoadClientOAuthOptions{
		ConfigFilePath: "/nonexistent/path/config.toml",
	})
	require.NoError(t, err)
	assert.Nil(t, result.OAuth)
}

func TestStoreClientOAuth(t *testing.T) {
	f, err := os.CreateTemp("", "temporal-config-*.toml")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	// Start with existing content
	_, err = f.WriteString(`
[profile.default]
address = "localhost:7233"
`)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	// Store OAuth config
	err = cliext.StoreClientOAuth(cliext.StoreClientOAuthOptions{
		ConfigFilePath: f.Name(),
		OAuth: &cliext.OAuthConfig{
			OAuthClientConfig: cliext.OAuthClientConfig{
				ClientID:     "new-client",
				ClientSecret: "new-secret",
				TokenURL:     "https://new.example.com/token",
				AuthURL:      "https://new.example.com/auth",
				Scopes:       []string{"read", "write"},
				RequestParams: map[string]string{
					"audience": "https://new-api.example.com",
				},
			},
			OAuthToken: cliext.OAuthToken{
				AccessToken:          "new-access-token",
				RefreshToken:         "new-refresh-token",
				TokenType:            "Bearer",
				AccessTokenExpiresAt: time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
			},
		},
	})
	require.NoError(t, err)

	// Verify by loading
	result, err := cliext.LoadClientOAuth(cliext.LoadClientOAuthOptions{
		ConfigFilePath: f.Name(),
	})
	require.NoError(t, err)
	require.NotNil(t, result.OAuth)

	assert.Equal(t, "new-client", result.OAuth.ClientID)
	assert.Equal(t, "new-secret", result.OAuth.ClientSecret)
	assert.Equal(t, "new-access-token", result.OAuth.AccessToken)
	assert.Equal(t, "new-refresh-token", result.OAuth.RefreshToken)
	assert.Equal(t, []string{"read", "write"}, result.OAuth.Scopes)
	assert.Equal(t, map[string]string{"audience": "https://new-api.example.com"}, result.OAuth.RequestParams)
}

func TestStoreClientOAuth_Remove(t *testing.T) {
	f, err := os.CreateTemp("", "temporal-config-*.toml")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	// Start with OAuth config
	_, err = f.WriteString(`
[profile.default]
address = "localhost:7233"

[profile.default.oauth]
client_id = "test-client"
access_token = "test-token"
refresh_token = "test-refresh"
`)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	// Remove OAuth by storing nil
	err = cliext.StoreClientOAuth(cliext.StoreClientOAuthOptions{
		ConfigFilePath: f.Name(),
		OAuth:          nil,
	})
	require.NoError(t, err)

	// Verify OAuth is removed
	result, err := cliext.LoadClientOAuth(cliext.LoadClientOAuthOptions{
		ConfigFilePath: f.Name(),
	})
	require.NoError(t, err)
	assert.Nil(t, result.OAuth)
}

func TestStoreClientOAuth_CreateNewFile(t *testing.T) {
	f, err := os.CreateTemp("", "temporal-config-*.toml")
	require.NoError(t, err)
	require.NoError(t, f.Close())
	require.NoError(t, os.Remove(f.Name()))
	defer os.Remove(f.Name())

	// Store to non-existent file
	err = cliext.StoreClientOAuth(cliext.StoreClientOAuthOptions{
		ConfigFilePath: f.Name(),
		OAuth: &cliext.OAuthConfig{
			OAuthClientConfig: cliext.OAuthClientConfig{
				ClientID: "new-client",
			},
			OAuthToken: cliext.OAuthToken{
				AccessToken:  "new-token",
				RefreshToken: "new-refresh",
			},
		},
	})
	require.NoError(t, err)

	// Verify file was created with OAuth
	result, err := cliext.LoadClientOAuth(cliext.LoadClientOAuthOptions{
		ConfigFilePath: f.Name(),
	})
	require.NoError(t, err)
	require.NotNil(t, result.OAuth)
	assert.Equal(t, "new-client", result.OAuth.ClientID)
	assert.Equal(t, "new-token", result.OAuth.AccessToken)
}

func TestStoreClientOAuth_PreservesOtherContent(t *testing.T) {
	f, err := os.CreateTemp("", "temporal-config-*.toml")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	// Start with existing content including non-OAuth settings
	_, err = f.WriteString(`
[profile.default]
address = "localhost:7233"
namespace = "my-namespace"

[profile.other]
address = "other:7233"
`)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	// Store OAuth config
	err = cliext.StoreClientOAuth(cliext.StoreClientOAuthOptions{
		ConfigFilePath: f.Name(),
		OAuth: &cliext.OAuthConfig{
			OAuthClientConfig: cliext.OAuthClientConfig{
				ClientID: "test-client",
			},
			OAuthToken: cliext.OAuthToken{
				AccessToken:  "test-token",
				RefreshToken: "test-refresh",
			},
		},
	})
	require.NoError(t, err)

	// Read file content and verify other settings are preserved
	content, err := os.ReadFile(f.Name())
	require.NoError(t, err)
	assert.Contains(t, string(content), "localhost:7233")
	assert.Contains(t, string(content), "my-namespace")
	assert.Contains(t, string(content), "other:7233")
	assert.Contains(t, string(content), "test-client")
}
