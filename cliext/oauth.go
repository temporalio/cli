package cliext

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/browser"
	"golang.org/x/oauth2"
)

type OAuthClientConfig struct {
	// ClientID is the OAuth client ID.
	ClientID string
	// ClientSecret is the OAuth client secret.
	ClientSecret string
	// TokenURL is the OAuth token endpoint URL.
	TokenURL string
	// AuthURL is the OAuth authorization endpoint URL.
	AuthURL string
	// Scopes are the requested OAuth scopes.
	Scopes []string
	// RequestParams are additional parameters to include in OAuth requests.
	RequestParams map[string]string
}

type OAuthToken struct {
	// AccessToken is the current access token.
	AccessToken string
	// AccessTokenExpiresAt is when the access token expires.
	AccessTokenExpiresAt time.Time
	// AccessTokenRefreshed indicates whether the access token was just refreshed.
	AccessTokenRefreshed bool
	// RefreshToken is the refresh token for obtaining new access tokens.
	RefreshToken string
	// TokenType is the type of token (usually "Bearer").
	TokenType string
}

type OAuthConfig struct {
	OAuthClientConfig
	OAuthToken
}

func oauthTokenFromOAuth2(token *oauth2.Token) OAuthToken {
	return OAuthToken{
		AccessToken:          token.AccessToken,
		AccessTokenExpiresAt: token.Expiry,
		RefreshToken:         token.RefreshToken,
		TokenType:            token.TokenType,
	}
}

// OAuthClient handles OAuth authentication using the Authorization Code Flow.
// See https://tools.ietf.org/html/rfc6749#section-4.1 for details.
type OAuthClient struct {
	Options OAuthClientConfig
	// OnAuthURL is called with the authorization URL. If nil, the URL is opened in the browser.
	// This can be set to override the default behavior, e.g., for testing.
	OnAuthURL func(authURL string)
}

func NewOAuthClient(opts OAuthClientConfig) *OAuthClient {
	return &OAuthClient{Options: opts}
}

func (c *OAuthClient) createOAuth2Config(redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     c.Options.ClientID,
		ClientSecret: c.Options.ClientSecret,
		RedirectURL:  redirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  c.Options.AuthURL,
			TokenURL: c.Options.TokenURL,
		},
		Scopes: c.Options.Scopes,
	}
}

type oauthCallbackResult struct {
	code string
	err  error
}

// Login performs the OAuth Authorization Code Flow with PKCE (Proof Key for Code Exchange).
//
// It starts a local HTTP server to receive the callback, generates an authorization URL,
// and exchanges the authorization code for a token.
func (c *OAuthClient) Login(ctx context.Context) (OAuthToken, error) {
	// Get a free port and start callback listener
	port, err := GetFreePort("127.0.0.1")
	if err != nil {
		return OAuthToken{}, fmt.Errorf("failed to get free port: %w", err)
	}
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return OAuthToken{}, fmt.Errorf("failed to start callback server: %w", err)
	}
	defer listener.Close()

	// Generate PKCE challenge.
	verifier := oauth2.GenerateVerifier()
	authOpts := []oauth2.AuthCodeOption{
		oauth2.S256ChallengeOption(verifier),
	}

	// Add any additional request parameters.
	for key, value := range c.Options.RequestParams {
		authOpts = append(authOpts, oauth2.SetAuthURLParam(key, value))
	}

	// Generate random state for CSRF protection.
	var stateBytes [16]byte
	if _, err := rand.Read(stateBytes[:]); err != nil {
		return OAuthToken{}, fmt.Errorf("failed to generate state: %w", err)
	}
	state := base64.RawURLEncoding.EncodeToString(stateBytes[:])

	// Generate authorization URL.
	redirectURL := fmt.Sprintf("http://127.0.0.1:%d/callback", port)
	cfg := c.createOAuth2Config(redirectURL)
	authURL := cfg.AuthCodeURL(state, authOpts...)

	// Handle auth URL notification.
	if c.OnAuthURL != nil {
		c.OnAuthURL(authURL)
	} else {
		fmt.Printf("Opening browser to authorize. If it doesn't open, visit: %s\n", authURL)
		_ = browser.OpenURL(authURL)
	}

	// Start HTTP server to handle callback.
	var once sync.Once
	resultCh := make(chan oauthCallbackResult, 1)
	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/callback" {
				http.NotFound(w, r)
				return
			}

			// Use sync.Once to only process the first callback.
			once.Do(func() {
				query := r.URL.Query()

				// Check for OAuth error response.
				if errCode := query.Get("error"); errCode != "" {
					errDesc := query.Get("error_description")
					if errDesc == "" {
						errDesc = errCode
					}
					resultCh <- oauthCallbackResult{err: fmt.Errorf("authorization failed: %s", errDesc)}
					http.Error(w, fmt.Sprintf("Authorization failed: %s", errDesc), http.StatusBadRequest)
					return
				}

				// Validate state to prevent CSRF.
				if query.Get("state") != state {
					resultCh <- oauthCallbackResult{err: fmt.Errorf("invalid state parameter")}
					http.Error(w, "Invalid state parameter", http.StatusBadRequest)
					return
				}

				// Check for authorization code.
				code := query.Get("code")
				if code == "" {
					resultCh <- oauthCallbackResult{err: fmt.Errorf("missing authorization code")}
					http.Error(w, "Missing authorization code", http.StatusBadRequest)
					return
				}

				resultCh <- oauthCallbackResult{code: code}
				fmt.Fprint(w, "Authorization successful! You can close this window.")
			})
		}),
	}
	go server.Serve(listener)
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()

	// Wait for callback result or context cancellation.
	var result oauthCallbackResult
	select {
	case result = <-resultCh:
	case <-ctx.Done():
		return OAuthToken{}, ctx.Err()
	}
	if result.err != nil {
		return OAuthToken{}, result.err
	}

	// Exchange code for token with PKCE verifier.
	token, err := cfg.Exchange(ctx, result.code, oauth2.VerifierOption(verifier))
	if err != nil {
		return OAuthToken{}, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	return oauthTokenFromOAuth2(token), nil
}

// Token returns a valid access token, refreshing if necessary.
func (c *OAuthClient) Token(ctx context.Context, config *OAuthConfig) (OAuthToken, error) {
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
		if requiresLogin(err) {
			return c.Login(ctx)
		}
		return OAuthToken{}, fmt.Errorf("failed to refresh token: %w", err)
	}

	token := oauthTokenFromOAuth2(newToken)
	token.AccessTokenRefreshed = true
	return token, nil
}

// requiresLogin checks if the error indicates an invalid or expired refresh token.
func requiresLogin(err error) bool {
	var retrieveErr *oauth2.RetrieveError
	if errors.As(err, &retrieveErr) && retrieveErr.ErrorCode == "invalid_grant" {
		return true
	}
	return false
}
