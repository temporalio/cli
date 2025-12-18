package cliext

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/BurntSushi/toml"
	"go.temporal.io/sdk/contrib/envconfig"
)

// OAuthClientConfig contains OAuth client credentials and endpoints.
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

// OAuthToken contains OAuth token information.
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

// OAuthConfig combines OAuth client configuration with token information.
type OAuthConfig struct {
	OAuthClientConfig
	OAuthToken
}

// oauthConfigTOML is the TOML representation of OAuthConfig.
type oauthConfigTOML struct {
	ClientID      string          `toml:"client_id,omitempty"`
	ClientSecret  string          `toml:"client_secret,omitempty"`
	TokenURL      string          `toml:"token_url,omitempty"`
	AuthURL       string          `toml:"auth_url,omitempty"`
	AccessToken   string          `toml:"access_token,omitempty"`
	RefreshToken  string          `toml:"refresh_token,omitempty"`
	TokenType     string          `toml:"token_type,omitempty"`
	ExpiresAt     string          `toml:"expires_at,omitempty"`
	Scopes        []string        `toml:"scopes,omitempty"`
	RequestParams inlineStringMap `toml:"request_params,omitempty"`
}

type rawProfileWithOAuth struct {
	OAuth *oauthConfigTOML `toml:"oauth"`
}

type rawConfigWithOAuth struct {
	Profile map[string]*rawProfileWithOAuth `toml:"profile"`
}

// inlineStringMap wraps a map to marshal as an inline TOML table.
type inlineStringMap map[string]string

func (m inlineStringMap) MarshalTOML() ([]byte, error) {
	if len(m) == 0 {
		return nil, nil
	}

	// Sort keys for deterministic output.
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	buf.WriteString("{ ")
	for i, k := range keys {
		if i > 0 {
			buf.WriteString(", ")
		}
		fmt.Fprintf(&buf, "%s = %q", k, m[k])
	}
	buf.WriteString(" }")
	return buf.Bytes(), nil
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
	envLookup := opts.EnvLookup
	if envLookup == nil {
		envLookup = envconfig.EnvLookupOS
	}

	// Resolve config file path.
	configFilePath := opts.ConfigFilePath
	if configFilePath == "" {
		configFilePath, _ = envLookup.LookupEnv("TEMPORAL_CONFIG_FILE")
	}
	if configFilePath == "" {
		var err error
		configFilePath, err = envconfig.DefaultConfigFilePath()
		if err != nil {
			return LoadClientOAuthResult{}, fmt.Errorf("failed to get default config path: %w", err)
		}
	}

	// Resolve profile name.
	profileName := opts.ProfileName
	if profileName == "" {
		profileName, _ = envLookup.LookupEnv("TEMPORAL_PROFILE")
	}
	if profileName == "" {
		profileName = envconfig.DefaultConfigFileProfile
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
	envLookup := opts.EnvLookup
	if envLookup == nil {
		envLookup = envconfig.EnvLookupOS
	}

	// Resolve config file path.
	configFilePath := opts.ConfigFilePath
	if configFilePath == "" {
		configFilePath, _ = envLookup.LookupEnv("TEMPORAL_CONFIG_FILE")
	}
	if configFilePath == "" {
		var err error
		configFilePath, err = envconfig.DefaultConfigFilePath()
		if err != nil {
			return fmt.Errorf("failed to get default config path: %w", err)
		}
	}

	// Resolve profile name.
	profileName := opts.ProfileName
	if profileName == "" {
		profileName, _ = envLookup.LookupEnv("TEMPORAL_PROFILE")
	}
	if profileName == "" {
		profileName = envconfig.DefaultConfigFileProfile
	}

	// Load existing OAuth configs from file.
	oauthByProfile, err := loadOAuthConfigFromFile(configFilePath)
	if err != nil {
		return fmt.Errorf("failed to load existing config: %w", err)
	}
	if oauthByProfile == nil {
		oauthByProfile = make(map[string]*OAuthConfig)
	}

	// Update the OAuth config for this profile.
	oauthByProfile[profileName] = opts.OAuth

	// Read existing file content to preserve non-OAuth sections.
	existingContent, err := os.ReadFile(configFilePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Write the updated config.
	newContent, err := mergeOAuthIntoConfig(existingContent, oauthByProfile)
	if err != nil {
		return fmt.Errorf("failed to merge OAuth config: %w", err)
	}

	// Ensure directory exists.
	if err := os.MkdirAll(filepath.Dir(configFilePath), 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configFilePath, newContent, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
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
		oauth := &OAuthConfig{
			OAuthClientConfig: OAuthClientConfig{
				ClientID:      cfg.ClientID,
				ClientSecret:  cfg.ClientSecret,
				TokenURL:      cfg.TokenURL,
				AuthURL:       cfg.AuthURL,
				RequestParams: cfg.RequestParams,
				Scopes:        cfg.Scopes,
			},
			OAuthToken: OAuthToken{
				AccessToken:  cfg.AccessToken,
				RefreshToken: cfg.RefreshToken,
				TokenType:    cfg.TokenType,
			},
		}
		if cfg.ExpiresAt != "" {
			t, err := time.Parse(time.RFC3339, cfg.ExpiresAt)
			if err != nil {
				return nil, fmt.Errorf("failed to parse expires_at for profile %q: %w", profileName, err)
			}
			oauth.AccessTokenExpiresAt = t
		}
		oauthByProfile[profileName] = oauth
	}
	return oauthByProfile, nil
}

// mergeOAuthIntoConfig merges OAuth configurations into existing TOML content.
// It preserves all non-OAuth content and updates/adds OAuth sections per profile.
func mergeOAuthIntoConfig(existingContent []byte, oauthByProfile map[string]*OAuthConfig) ([]byte, error) {
	// Parse existing content to get the structure.
	var existingRaw map[string]any
	if len(existingContent) > 0 {
		if _, err := toml.Decode(string(existingContent), &existingRaw); err != nil {
			return nil, fmt.Errorf("failed to parse existing config: %w", err)
		}
	}
	if existingRaw == nil {
		existingRaw = make(map[string]any)
	}

	// Get or create the profile section.
	profileSection, ok := existingRaw["profile"].(map[string]any)
	if !ok {
		profileSection = make(map[string]any)
		existingRaw["profile"] = profileSection
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

	// Marshal back to TOML.
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(existingRaw); err != nil {
		return nil, fmt.Errorf("failed to encode config: %w", err)
	}

	return buf.Bytes(), nil
}

// oauthConfigToTOML converts OAuthConfig to its TOML representation.
func oauthConfigToTOML(oauth *OAuthConfig) *oauthConfigTOML {
	if oauth == nil {
		return nil
	}
	result := &oauthConfigTOML{
		ClientID:      oauth.ClientID,
		ClientSecret:  oauth.ClientSecret,
		TokenURL:      oauth.TokenURL,
		AuthURL:       oauth.AuthURL,
		AccessToken:   oauth.AccessToken,
		RefreshToken:  oauth.RefreshToken,
		TokenType:     oauth.TokenType,
		Scopes:        oauth.Scopes,
		RequestParams: oauth.RequestParams,
	}
	if !oauth.AccessTokenExpiresAt.IsZero() {
		result.ExpiresAt = oauth.AccessTokenExpiresAt.Format(time.RFC3339)
	}
	return result
}
