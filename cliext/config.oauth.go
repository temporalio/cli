package cliext

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.temporal.io/sdk/contrib/envconfig"
	"golang.org/x/oauth2"
)

// oauthConfigJSON is an intermediate struct for JSON serialization of OAuth config.
type oauthConfigJSON struct {
	ClientID     string   `json:"client_id,omitempty"`
	ClientSecret string   `json:"client_secret,omitempty"`
	TokenURL     string   `json:"token_url,omitempty"`
	AuthURL      string   `json:"auth_url,omitempty"`
	RedirectURL  string   `json:"redirect_url,omitempty"`
	AccessToken  string   `json:"access_token,omitempty"`
	RefreshToken string   `json:"refresh_token,omitempty"`
	TokenType    string   `json:"token_type,omitempty"`
	ExpiresAt    string   `json:"expires_at,omitempty"`
	Scopes       []string `json:"scopes,omitempty"`
}

// OAuthConfig combines OAuth client configuration with token information.
type OAuthConfig struct {
	// ClientConfig is the OAuth 2.0 client configuration.
	ClientConfig *oauth2.Config
	// Token is the OAuth 2.0 token.
	Token *oauth2.Token
}

// newTokenSource creates an oauth2.TokenSource that automatically refreshes
// the access token when it expires.
func (c *OAuthConfig) newTokenSource(ctx context.Context) oauth2.TokenSource {
	if c.ClientConfig == nil || c.Token == nil {
		return nil
	}

	return oauth2.ReuseTokenSourceWithExpiry(c.Token, c.ClientConfig.TokenSource(ctx, c.Token), time.Minute)
}

// LoadClientOAuthOptions are options for LoadClientOAuth.
type LoadClientOAuthOptions struct {
	// ConfigFilePath overrides the config file path. If empty, uses TEMPORAL_CONFIG_FILE
	// env var or the default path.
	ConfigFilePath string
	// ProfileName specifies which profile to load OAuth from. If empty, uses TEMPORAL_PROFILE
	// env var or "default".
	ProfileName string
	// EnvLookup overrides environment variable lookup. If nil, uses os.LookupEnv.
	EnvLookup envconfig.EnvLookup
}

// LoadClientOAuthResult is the result of LoadClientOAuth.
type LoadClientOAuthResult struct {
	// OAuth is the loaded OAuth configuration, or nil if not configured.
	OAuth *OAuthConfig
	// ConfigFilePath is the resolved path to the config file.
	ConfigFilePath string
	// ProfileName is the resolved profile name.
	ProfileName string
}

// LoadClientOAuth loads OAuth configuration from the config file for a specific profile.
func LoadClientOAuth(opts LoadClientOAuthOptions) (LoadClientOAuthResult, error) {
	configFilePath, profileName, err := resolveConfigAndProfile(opts.ConfigFilePath, opts.ProfileName, opts.EnvLookup)
	if err != nil {
		return LoadClientOAuthResult{}, err
	}

	// Load OAuth from file.
	oauthByProfile, err := loadOAuthConfigFromFile(configFilePath)
	if err != nil {
		return LoadClientOAuthResult{}, err
	}

	return LoadClientOAuthResult{
		OAuth:          oauthByProfile[profileName],
		ConfigFilePath: configFilePath,
		ProfileName:    profileName,
	}, nil
}

// loadOAuthConfigFromFile loads OAuth configurations for all profiles from a TOML file.
func loadOAuthConfigFromFile(path string) (map[string]*OAuthConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Use envconfig's FromTOML with AdditionalProfileFields to capture OAuth fields
	var conf envconfig.ClientConfig
	additional := make(map[string]map[string]any)
	if err := conf.FromTOML(data, envconfig.ClientConfigFromTOMLOptions{
		AdditionalProfileFields: additional,
	}); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	oauthByProfile := make(map[string]*OAuthConfig)
	for profileName, fields := range additional {
		oauthRaw, ok := fields["oauth"].(map[string]any)
		if !ok {
			continue
		}
		oauth, err := oauthConfigFromMap(oauthRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse oauth for profile %q: %w", profileName, err)
		}
		oauthByProfile[profileName] = oauth
	}
	return oauthByProfile, nil
}

// oauthConfigFromMap converts a map[string]any to OAuthConfig using JSON as intermediary.
func oauthConfigFromMap(m map[string]any) (*OAuthConfig, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal oauth config: %w", err)
	}

	var cfg oauthConfigJSON
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal oauth config: %w", err)
	}

	// Parse expiry time if present
	var expiry time.Time
	if cfg.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, cfg.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse expires_at: %w", err)
		}
		expiry = t
	}

	return &OAuthConfig{
		ClientConfig: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       cfg.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  cfg.AuthURL,
				TokenURL: cfg.TokenURL,
			},
		},
		Token: &oauth2.Token{
			AccessToken:  cfg.AccessToken,
			RefreshToken: cfg.RefreshToken,
			TokenType:    cfg.TokenType,
			Expiry:       expiry,
		},
	}, nil
}

// resolveConfigAndProfile resolves the config file path and profile name.
func resolveConfigAndProfile(configFilePath, profileName string, envLookup envconfig.EnvLookup) (string, string, error) {
	if envLookup == nil {
		envLookup = envconfig.EnvLookupOS
	}

	// Resolve config file path.
	if configFilePath == "" {
		configFilePath, _ = envLookup.LookupEnv("TEMPORAL_CONFIG_FILE")
	}
	if configFilePath == "" {
		var err error
		configFilePath, err = envconfig.DefaultConfigFilePath()
		if err != nil {
			return "", "", fmt.Errorf("failed to get default config path: %w", err)
		}
	}

	// Resolve profile name.
	if profileName == "" {
		profileName, _ = envLookup.LookupEnv("TEMPORAL_PROFILE")
	}
	if profileName == "" {
		profileName = envconfig.DefaultConfigFileProfile
	}

	return configFilePath, profileName, nil
}

// StoreClientOAuthOptions are options for StoreClientOAuth.
type StoreClientOAuthOptions struct {
	// ConfigFilePath overrides the config file path. If empty, uses TEMPORAL_CONFIG_FILE
	// env var or the default path.
	ConfigFilePath string
	// ProfileName specifies which profile to store OAuth for. If empty, uses TEMPORAL_PROFILE
	// env var or "default".
	ProfileName string
	// OAuth is the OAuth configuration to store. If nil, removes OAuth from the profile.
	OAuth *OAuthConfig
	// EnvLookup overrides environment variable lookup. If nil, uses os.LookupEnv.
	EnvLookup envconfig.EnvLookup
}

// StoreClientOAuth stores OAuth configuration in the config file for a specific profile.
// If OAuth is nil, it removes the OAuth configuration from the profile.
// This function preserves all other content in the config file.
func StoreClientOAuth(opts StoreClientOAuthOptions) error {
	configFilePath, profileName, err := resolveConfigAndProfile(opts.ConfigFilePath, opts.ProfileName, opts.EnvLookup)
	if err != nil {
		return err
	}

	// Read and parse existing file content using envconfig.
	existingContent, err := os.ReadFile(configFilePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var conf envconfig.ClientConfig
	additional := make(map[string]map[string]any)
	if len(existingContent) > 0 {
		if err := conf.FromTOML(existingContent, envconfig.ClientConfigFromTOMLOptions{
			AdditionalProfileFields: additional,
		}); err != nil {
			return fmt.Errorf("failed to parse existing config: %w", err)
		}
	}

	// Ensure the profile exists in the config.
	if conf.Profiles == nil {
		conf.Profiles = make(map[string]*envconfig.ClientConfigProfile)
	}
	if conf.Profiles[profileName] == nil {
		conf.Profiles[profileName] = &envconfig.ClientConfigProfile{}
	}

	// Update the OAuth config for this profile in additional fields.
	if additional[profileName] == nil {
		additional[profileName] = make(map[string]any)
	}
	if opts.OAuth == nil {
		delete(additional[profileName], "oauth")
	} else {
		additional[profileName]["oauth"] = oauthConfigToMap(opts.OAuth)
	}

	// Marshal back to TOML using envconfig.
	data, err := conf.ToTOML(envconfig.ClientConfigToTOMLOptions{
		AdditionalProfileFields: additional,
	})
	if err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	// Ensure directory exists.
	if err := os.MkdirAll(filepath.Dir(configFilePath), 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configFilePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// oauthConfigToMap converts OAuthConfig to map[string]any using JSON as intermediary.
func oauthConfigToMap(oauth *OAuthConfig) map[string]any {
	if oauth == nil || oauth.ClientConfig == nil || oauth.Token == nil {
		return nil
	}

	cfg := oauthConfigJSON{
		ClientID:     oauth.ClientConfig.ClientID,
		ClientSecret: oauth.ClientConfig.ClientSecret,
		TokenURL:     oauth.ClientConfig.Endpoint.TokenURL,
		AuthURL:      oauth.ClientConfig.Endpoint.AuthURL,
		RedirectURL:  oauth.ClientConfig.RedirectURL,
		AccessToken:  oauth.Token.AccessToken,
		RefreshToken: oauth.Token.RefreshToken,
		TokenType:    oauth.Token.TokenType,
		Scopes:       oauth.ClientConfig.Scopes,
	}
	if !oauth.Token.Expiry.IsZero() {
		cfg.ExpiresAt = oauth.Token.Expiry.Format(time.RFC3339)
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return nil
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil
	}
	return result
}

