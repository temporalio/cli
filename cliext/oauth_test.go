package cliext_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/cli/cliext"
)

type mockOAuthServer struct {
	t        *testing.T
	server   *httptest.Server
	authURL  string
	tokenURL string

	AuthorizeHandler     func(http.ResponseWriter, *http.Request)
	TokenExchangeHandler func(http.ResponseWriter, *http.Request) // authorization_code grant
	TokenRefreshHandler  func(http.ResponseWriter, *http.Request) // refresh_token grant

	capturedAuthURL atomic.Value
}

func newMockOAuthServer(t *testing.T) *mockOAuthServer {
	m := &mockOAuthServer{t: t}

	mux := http.NewServeMux()
	mux.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		if m.AuthorizeHandler == nil {
			http.Error(w, "AuthorizeHandler not set", http.StatusInternalServerError)
			return
		}
		m.AuthorizeHandler(w, r)
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		w.Header().Set("Content-Type", "application/json")
		switch r.FormValue("grant_type") {
		case "authorization_code":
			if m.TokenExchangeHandler == nil {
				http.Error(w, "TokenExchangeHandler not set", http.StatusInternalServerError)
				return
			}
			m.TokenExchangeHandler(w, r)
		case "refresh_token":
			if m.TokenRefreshHandler == nil {
				http.Error(w, "TokenRefreshHandler not set", http.StatusInternalServerError)
				return
			}
			m.TokenRefreshHandler(w, r)
		}
	})

	m.server = httptest.NewServer(mux)
	m.authURL = m.server.URL + "/authorize"
	m.tokenURL = m.server.URL + "/token"
	t.Cleanup(m.server.Close)

	return m
}

func (m *mockOAuthServer) newClient(cfg cliext.OAuthClientConfig) *cliext.OAuthClient {
	cfg.AuthURL = m.authURL
	cfg.TokenURL = m.tokenURL
	c := cliext.NewOAuthClient(cfg)
	c.OnAuthURL = func(url string) {
		go func() {
			resp, _ := http.Get(url)
			resp.Body.Close()
		}()
	}
	return c
}

func (m *mockOAuthServer) redirectWithCode(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	http.Redirect(w, r, fmt.Sprintf("%s?code=test-code&state=%s",
		query.Get("redirect_uri"), query.Get("state")), http.StatusFound)
}

func TestOAuthClient_Login(t *testing.T) {
	s := newMockOAuthServer(t)

	s.AuthorizeHandler = s.redirectWithCode
	s.TokenExchangeHandler = func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "test-access-token",
			"refresh_token": "test-refresh-token",
			"token_type":    "Bearer",
			"expires_in":    3600,
		})
	}

	c := s.newClient(cliext.OAuthClientConfig{ClientID: "test-client"})

	token, err := login(t, c)

	require.NoError(t, err)
	assert.Equal(t, "test-access-token", token.AccessToken)
	assert.Equal(t, "test-refresh-token", token.RefreshToken)
	assert.Equal(t, "Bearer", token.TokenType)
}

func TestOAuthClient_Login_RequestParams(t *testing.T) {
	var audience, custom string

	s := newMockOAuthServer(t)
	s.AuthorizeHandler = func(w http.ResponseWriter, r *http.Request) {
		audience = r.URL.Query().Get("audience")
		custom = r.URL.Query().Get("custom")
		s.redirectWithCode(w, r)
	}
	s.TokenExchangeHandler = func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"access_token": "token"})
	}

	c := s.newClient(cliext.OAuthClientConfig{
		ClientID: "test-client",
		RequestParams: map[string]string{
			"audience": "https://api.example.com",
			"custom":   "value",
		},
	})

	_, err := login(t, c)

	require.NoError(t, err)
	assert.Equal(t, "https://api.example.com", audience)
	assert.Equal(t, "value", custom)
}

func TestOAuthClient_Login_Scopes(t *testing.T) {
	var scopes string

	s := newMockOAuthServer(t)
	s.AuthorizeHandler = func(w http.ResponseWriter, r *http.Request) {
		scopes = r.URL.Query().Get("scope")
		s.redirectWithCode(w, r)
	}
	s.TokenExchangeHandler = func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"access_token":"ok"}`)
	}

	c := s.newClient(cliext.OAuthClientConfig{
		ClientID: "test-client",
		Scopes:   []string{"openid", "profile", "email"},
	})

	_, err := login(t, c)

	require.NoError(t, err)
	assert.Contains(t, scopes, "openid")
	assert.Contains(t, scopes, "profile")
	assert.Contains(t, scopes, "email")
}

func TestOAuthClient_Login_InvalidState(t *testing.T) {
	s := newMockOAuthServer(t)
	s.AuthorizeHandler = func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		http.Redirect(w, r, fmt.Sprintf("%s?code=test-code&state=wrong-state",
			query.Get("redirect_uri")), http.StatusFound)
	}

	c := s.newClient(cliext.OAuthClientConfig{ClientID: "test-client"})

	_, err := login(t, c)

	assert.ErrorContains(t, err, "invalid state")
}

func TestOAuthClient_Login_ErrorCallback(t *testing.T) {
	s := newMockOAuthServer(t)
	s.AuthorizeHandler = func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		http.Redirect(w, r, fmt.Sprintf("%s?error=access_denied&error_description=User+denied+access&state=%s",
			query.Get("redirect_uri"), query.Get("state")), http.StatusFound)
	}

	c := s.newClient(cliext.OAuthClientConfig{ClientID: "test-client"})

	_, err := login(t, c)
	assert.ErrorContains(t, err, "User denied access")
}

func TestOAuthClient_Login_PKCE(t *testing.T) {
	var codeChallenge, codeChallengeMethod, codeVerifier string

	s := newMockOAuthServer(t)
	s.AuthorizeHandler = func(w http.ResponseWriter, r *http.Request) {
		codeChallenge = r.URL.Query().Get("code_challenge")
		codeChallengeMethod = r.URL.Query().Get("code_challenge_method")
		s.redirectWithCode(w, r)
	}
	s.TokenExchangeHandler = func(w http.ResponseWriter, r *http.Request) {
		codeVerifier = r.FormValue("code_verifier")
		fmt.Fprint(w, `{"access_token":"ok"}`)
	}

	c := s.newClient(cliext.OAuthClientConfig{ClientID: "test-client"})

	_, err := login(t, c)

	require.NoError(t, err)
	assert.NotEmpty(t, codeChallenge)
	assert.Equal(t, "S256", codeChallengeMethod)
	assert.NotEmpty(t, codeVerifier)
}

func TestOAuthClient_Token_Refresh(t *testing.T) {
	s := newMockOAuthServer(t)
	s.TokenRefreshHandler = func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"access_token":"refreshed-token"}`)
	}

	c := s.newClient(cliext.OAuthClientConfig{})

	token, err := c.Token(t.Context(), &cliext.OAuthConfig{
		OAuthToken: cliext.OAuthToken{
			AccessTokenExpiresAt: time.Now(),
			RefreshToken:         "refresh-token",
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "refreshed-token", token.AccessToken)
	assert.True(t, token.AccessTokenRefreshed)
}

func TestOAuthClient_Token_ReLogin(t *testing.T) {
	s := newMockOAuthServer(t)
	s.AuthorizeHandler = s.redirectWithCode
	s.TokenRefreshHandler = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":"invalid_grant"}`)
	}
	s.TokenExchangeHandler = func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"access_token":"relogin-token"}`)
	}

	c := s.newClient(cliext.OAuthClientConfig{ClientID: "test-client"})

	token, err := c.Token(t.Context(), &cliext.OAuthConfig{
		OAuthClientConfig: c.Options,
		OAuthToken: cliext.OAuthToken{
			AccessTokenExpiresAt: time.Now().Add(-time.Hour),
			RefreshToken:         "invalid-refresh-token",
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "relogin-token", token.AccessToken)
}

func TestOAuthClient_Token_ValidToken(t *testing.T) {
	c := cliext.NewOAuthClient(cliext.OAuthClientConfig{})

	token, err := c.Token(t.Context(), &cliext.OAuthConfig{
		OAuthToken: cliext.OAuthToken{
			AccessToken:          "valid-token",
			AccessTokenExpiresAt: time.Now().Add(time.Hour),
			RefreshToken:         "refresh-token",
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "valid-token", token.AccessToken)
	assert.False(t, token.AccessTokenRefreshed)
}

func TestOAuthClient_Token_NoRefreshToken(t *testing.T) {
	c := cliext.NewOAuthClient(cliext.OAuthClientConfig{})

	_, err := c.Token(t.Context(), nil)
	assert.ErrorContains(t, err, "no refresh token")

	_, err = c.Token(t.Context(), &cliext.OAuthConfig{})
	assert.ErrorContains(t, err, "no refresh token")
}

func TestOAuthClient_Token_RefreshError(t *testing.T) {
	s := newMockOAuthServer(t)
	s.TokenRefreshHandler = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"error":"server_error"}`)
	}

	c := s.newClient(cliext.OAuthClientConfig{})

	_, err := c.Token(t.Context(), &cliext.OAuthConfig{
		OAuthToken: cliext.OAuthToken{
			AccessTokenExpiresAt: time.Now().Add(-time.Hour),
			RefreshToken:         "refresh-token",
		},
	})

	assert.ErrorContains(t, err, "failed to refresh token")
}

func login(t *testing.T, c *cliext.OAuthClient) (cliext.OAuthToken, error) {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	type result struct {
		token cliext.OAuthToken
		err   error
	}
	ch := make(chan result, 1)
	go func() {
		token, err := c.Login(ctx)
		ch <- result{token, err}
	}()

	r := <-ch
	return r.token, r.err
}
