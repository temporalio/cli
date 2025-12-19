package cliext

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"go.temporal.io/sdk/contrib/envconfig"
	"golang.org/x/oauth2"
)

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

	var raw rawConfigWithOAuth
	if _, err := toml.Decode(string(data), &raw); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	oauthByProfile := make(map[string]*OAuthConfig)
	for profileName, profile := range raw.Profile {
		if profile == nil || profile.OAuth == nil {
			oauthByProfile[profileName] = nil
			continue
		}
		cfg := profile.OAuth

		// Parse expiry time if present
		var expiry time.Time
		if cfg.ExpiresAt != "" {
			t, err := time.Parse(time.RFC3339, cfg.ExpiresAt)
			if err != nil {
				return nil, fmt.Errorf("failed to parse expires_at for profile %q: %w", profileName, err)
			}
			expiry = t
		}

		oauth := &OAuthConfig{
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
		}
		oauthByProfile[profileName] = oauth
	}
	return oauthByProfile, nil
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

	// Read and parse existing file content.
	existingContent, err := os.ReadFile(configFilePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var existingRaw map[string]any
	if len(existingContent) > 0 {
		if _, err := toml.Decode(string(existingContent), &existingRaw); err != nil {
			return fmt.Errorf("failed to parse existing config: %w", err)
		}
	}
	if existingRaw == nil {
		existingRaw = make(map[string]any)
	}

	// Load existing OAuth configs from the parsed content.
	oauthByProfile, err := parseOAuthFromRaw(existingRaw)
	if err != nil {
		return fmt.Errorf("failed to parse existing OAuth config: %w", err)
	}
	if oauthByProfile == nil {
		oauthByProfile = make(map[string]*OAuthConfig)
	}

	// Update the OAuth config for this profile.
	oauthByProfile[profileName] = opts.OAuth

	// Merge OAuth configs back into the raw structure.
	if err := mergeOAuthIntoRaw(existingRaw, oauthByProfile); err != nil {
		return err
	}

	// Marshal back to TOML.
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(existingRaw); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	// Ensure directory exists.
	if err := os.MkdirAll(filepath.Dir(configFilePath), 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configFilePath, buf.Bytes(), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func parseOAuthFromRaw(raw map[string]any) (map[string]*OAuthConfig, error) {
	var parsed rawConfigWithOAuth

	// Re-encode and decode to convert map[string]any to our struct.
	// This is simpler than manual type assertions for nested structures.
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(raw); err != nil {
		return nil, err
	}
	if _, err := toml.Decode(buf.String(), &parsed); err != nil {
		return nil, err
	}

	oauthByProfile := make(map[string]*OAuthConfig)
	for profileName, profile := range parsed.Profile {
		if profile == nil || profile.OAuth == nil {
			continue
		}
		cfg := profile.OAuth

		// Parse expiry time if present
		var expiry time.Time
		if cfg.ExpiresAt != "" {
			t, err := time.Parse(time.RFC3339, cfg.ExpiresAt)
			if err != nil {
				return nil, fmt.Errorf("failed to parse expires_at for profile %q: %w", profileName, err)
			}
			expiry = t
		}

		oauth := &OAuthConfig{
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
		}
		oauthByProfile[profileName] = oauth
	}
	return oauthByProfile, nil
}

// mergeOAuthIntoRaw merges OAuth configurations into a raw TOML structure.
func mergeOAuthIntoRaw(raw map[string]any, oauthByProfile map[string]*OAuthConfig) error {
	// Get or create the profile section.
	profileSection, ok := raw["profile"].(map[string]any)
	if !ok {
		profileSection = make(map[string]any)
		raw["profile"] = profileSection
	}

	// Update OAuth for each profile.
	for profileName, oauth := range oauthByProfile {
		profile, ok := profileSection[profileName].(map[string]any)
		if !ok {
			profile = make(map[string]any)
			profileSection[profileName] = profile
		}

		if oauth == nil {
			delete(profile, "oauth")
		} else {
			profile["oauth"] = oauthConfigToTOML(oauth)
		}
	}

	return nil
}

// oauthConfigTOML is the TOML representation of OAuthConfig.
type oauthConfigTOML struct {
	ClientID     string   `toml:"client_id,omitempty"`
	ClientSecret string   `toml:"client_secret,omitempty"`
	TokenURL     string   `toml:"token_url,omitempty"`
	AuthURL      string   `toml:"auth_url,omitempty"`
	RedirectURL  string   `toml:"redirect_url,omitempty"`
	AccessToken  string   `toml:"access_token,omitempty"`
	RefreshToken string   `toml:"refresh_token,omitempty"`
	TokenType    string   `toml:"token_type,omitempty"`
	ExpiresAt    string   `toml:"expires_at,omitempty"`
	Scopes       []string `toml:"scopes,omitempty"`
}

type rawProfileWithOAuth struct {
	OAuth *oauthConfigTOML `toml:"oauth"`
}

type rawConfigWithOAuth struct {
	Profile map[string]*rawProfileWithOAuth `toml:"profile"`
}

// oauthConfigToTOML converts OAuthConfig to its TOML representation.
func oauthConfigToTOML(oauth *OAuthConfig) *oauthConfigTOML {
	if oauth == nil || oauth.ClientConfig == nil || oauth.Token == nil {
		return nil
	}
	result := &oauthConfigTOML{
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
		result.ExpiresAt = oauth.Token.Expiry.Format(time.RFC3339)
	}
	return result
}
