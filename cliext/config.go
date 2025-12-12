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

type ClientConfig struct {
	Profiles map[string]*Profile
}

type LoadConfigOptions struct {
	// Override the file path to use to load the TOML file for config. Defaults to TEMPORAL_CONFIG_FILE environment
	// variable or if that is unset/empty, defaults to [os.UserConfigDir]/temporal/temporal.toml. If ConfigFileData is
	// set, this cannot be set and no file loading from disk occurs. Ignored if DisableFile is true.
	ConfigFilePath string
	// Override the environment variable lookup. If nil, defaults to [EnvLookupOS].
	EnvLookup envconfig.EnvLookup
}

type LoadConfigResult struct {
	// Config is the loaded configuration with its profiles.
	Config ClientConfig
	// ConfigFilePath is the resolved path to the configuration file that was loaded.
	// This may differ from the input if TEMPORAL_CONFIG_FILE env var was used.
	ConfigFilePath string
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

// LoadConfig loads the client configuration from the specified file or default location.
// If ConfigFilePath is empty, the TEMPORAL_CONFIG_FILE environment variable is checked,
// then the default path is used.
func LoadConfig(options LoadConfigOptions) (LoadConfigResult, error) {
	envLookup := options.EnvLookup
	if envLookup == nil {
		envLookup = envconfig.EnvLookupOS
	}

	// Resolve the config file path.
	resolvedPath := options.ConfigFilePath
	if resolvedPath == "" {
		resolvedPath, _ = envLookup.LookupEnv("TEMPORAL_CONFIG_FILE")
	}
	if resolvedPath == "" {
		var err error
		resolvedPath, err = envconfig.DefaultConfigFilePath()
		if err != nil {
			return LoadConfigResult{}, fmt.Errorf("failed to get default config path: %w", err)
		}
	}

	clientConfig, err := envconfig.LoadClientConfig(envconfig.LoadClientConfigOptions{
		ConfigFilePath: resolvedPath,
		EnvLookup:      envLookup,
	})
	if err != nil {
		return LoadConfigResult{}, err
	}

	// Load OAuth for all profiles by parsing the config file directly.
	oauthByProfile, err := loadOAuthConfigFromFile(resolvedPath)
	if err != nil {
		return LoadConfigResult{}, err
	}

	// Merge profiles and their OAuth configurations.
	profiles := make(map[string]*Profile)
	for name, baseProfile := range clientConfig.Profiles {
		profiles[name] = &Profile{
			ClientConfigProfile: *baseProfile,
			OAuth:               oauthByProfile[name],
		}
	}

	return LoadConfigResult{
		Config:         ClientConfig{Profiles: profiles},
		ConfigFilePath: resolvedPath,
	}, nil
}

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

type WriteConfigOptions struct {
	// Config is the configuration to write.
	Config ClientConfig
	// ConfigFilePath is the path to write the configuration file to.
	// If empty, TEMPORAL_CONFIG_FILE env var is checked, then the default path is used.
	ConfigFilePath string
	// Override the environment variable lookup. If nil, defaults to [EnvLookupOS].
	EnvLookup envconfig.EnvLookup
}

// ConfigToTOML serializes the configuration to TOML bytes.
func ConfigToTOML(config *ClientConfig) ([]byte, error) {
	// Build envconfig.ClientConfig from profiles.
	envConfig := &envconfig.ClientConfig{
		Profiles: make(map[string]*envconfig.ClientConfigProfile),
	}
	for name, p := range config.Profiles {
		if p != nil {
			envConfig.Profiles[name] = &p.ClientConfigProfile
		}
	}

	// Convert base config to TOML.
	b, err := envConfig.ToTOML(envconfig.ClientConfigToTOMLOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed building TOML: %w", err)
	}

	// Append OAuth sections per profile.
	var buf bytes.Buffer
	buf.Write(b)

	for name, profile := range config.Profiles {
		if profile == nil || profile.OAuth == nil {
			continue
		}
		oauthTOML := &oauthConfigTOML{
			ClientID:      profile.OAuth.ClientID,
			ClientSecret:  profile.OAuth.ClientSecret,
			TokenURL:      profile.OAuth.TokenURL,
			AuthURL:       profile.OAuth.AuthURL,
			AccessToken:   profile.OAuth.AccessToken,
			RefreshToken:  profile.OAuth.RefreshToken,
			TokenType:     profile.OAuth.TokenType,
			Scopes:        profile.OAuth.Scopes,
			RequestParams: profile.OAuth.RequestParams,
		}
		if !profile.OAuth.AccessTokenExpiresAt.IsZero() {
			oauthTOML.ExpiresAt = profile.OAuth.AccessTokenExpiresAt.Format(time.RFC3339)
		}

		oauthBytes, err := toml.Marshal(oauthTOML)
		if err != nil {
			return nil, fmt.Errorf("failed marshaling OAuth config: %w", err)
		}
		fmt.Fprintf(&buf, "\n[profile.%s.oauth]\n", name)
		buf.Write(oauthBytes)
	}
	return buf.Bytes(), nil
}

// WriteConfig writes the environment configuration to disk.
func WriteConfig(opts WriteConfigOptions) error {
	envLookup := opts.EnvLookup
	if envLookup == nil {
		envLookup = envconfig.EnvLookupOS
	}

	configFilePath := opts.ConfigFilePath
	if configFilePath == "" {
		configFilePath, _ = envLookup.LookupEnv("TEMPORAL_CONFIG_FILE")
		if configFilePath == "" {
			var err error
			if configFilePath, err = envconfig.DefaultConfigFilePath(); err != nil {
				return err
			}
		}
	}

	b, err := ConfigToTOML(&opts.Config)
	if err != nil {
		return err
	}

	// Write to file, making dirs as needed.
	if err := os.MkdirAll(filepath.Dir(configFilePath), 0700); err != nil {
		return fmt.Errorf("failed making config file parent dirs: %w", err)
	}
	if err := os.WriteFile(configFilePath, b, 0600); err != nil {
		return fmt.Errorf("failed writing config file: %w", err)
	}
	return nil
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
