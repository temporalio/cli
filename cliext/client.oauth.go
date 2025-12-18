package cliext

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2"
)

func oauthTokenFromOAuth2(token *oauth2.Token) OAuthToken {
	return OAuthToken{
		AccessToken:          token.AccessToken,
		AccessTokenExpiresAt: token.Expiry,
		RefreshToken:         token.RefreshToken,
		TokenType:            token.TokenType,
	}
}

type oauthClient struct {
	options OAuthClientConfig
}

func newOAuthClient(opts OAuthClientConfig) *oauthClient {
	return &oauthClient{options: opts}
}

func (c *oauthClient) createOAuth2Config(redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     c.options.ClientID,
		ClientSecret: c.options.ClientSecret,
		RedirectURL:  redirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  c.options.AuthURL,
			TokenURL: c.options.TokenURL,
		},
		Scopes: c.options.Scopes,
	}
}

// token returns a valid access token, refreshing if necessary.
func (c *oauthClient) token(ctx context.Context, config *OAuthConfig) (OAuthToken, error) {
	if config == nil || config.RefreshToken == "" {
		return OAuthToken{}, fmt.Errorf("no refresh token available")
	}

	// Check if token is still valid: not expired and not expiring within next minute.
	if !config.AccessTokenExpiresAt.IsZero() && time.Now().Before(config.AccessTokenExpiresAt.Add(-1*time.Minute)) {
		return config.OAuthToken, nil
	}

	// Token is expired or about to expire, refresh it.
	cfg := c.createOAuth2Config("")
	tokenSource := cfg.TokenSource(ctx, &oauth2.Token{RefreshToken: config.RefreshToken})
	newToken, err := tokenSource.Token()
	if err != nil {
		return OAuthToken{}, fmt.Errorf("failed to refresh token: %w", err)
	}

	token := oauthTokenFromOAuth2(newToken)
	token.AccessTokenRefreshed = true
	return token, nil
}

// getOAuthToken returns a valid OAuth access token, refreshing if necessary.
// If the token is still valid (not expired and not expiring within the next minute),
// it returns the existing token. Otherwise, it attempts to refresh using the refresh token.
func getOAuthToken(ctx context.Context, config *OAuthConfig) (OAuthToken, error) {
	if config == nil {
		return OAuthToken{}, fmt.Errorf("OAuth config is nil")
	}
	client := newOAuthClient(config.OAuthClientConfig)
	return client.token(ctx, config)
}

// NewOAuthDynamicTokenProvider creates a function that provides OAuth access tokens dynamically.
// The returned function loads OAuth configuration on-demand from the config file and refreshes
// tokens as needed. Returns empty string if no OAuth is configured.
func NewOAuthDynamicTokenProvider(opts ClientOptionsBuilder) func(context.Context) (string, error) {
	return func(ctx context.Context) (string, error) {
		result, err := LoadClientOAuth(LoadClientOAuthOptions{
			ConfigFilePath: opts.CommonOptions.ConfigFile,
			ProfileName:    opts.CommonOptions.Profile,
			EnvLookup:      opts.EnvLookup,
		})
		if err != nil {
			return "", fmt.Errorf("failed to load OAuth config: %w", err)
		}
		if result.OAuth == nil {
			return "", nil // No OAuth configured, return empty token
		}
		token, err := getOAuthToken(ctx, result.OAuth)
		if err != nil {
			return "", err
		}
		return token.AccessToken, nil
	}
}
