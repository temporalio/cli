package temporalcli

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/temporalio/cli/cliext"
	"go.temporal.io/sdk/contrib/envconfig"
)

// Profile represents a configuration profile with client settings and OAuth.
type Profile struct {
	envconfig.ClientConfigProfile
	OAuth *cliext.OAuthConfig
}

// FileConfig represents the structure of a Temporal configuration file.
type FileConfig struct {
	Profiles map[string]*Profile
}

type LoadProfileOptions struct {
	envconfig.LoadClientConfigProfileOptions
}

type LoadProfileResult struct {
	Config         FileConfig
	ConfigFilePath string
	ProfileName    string
	Profile        *Profile
}

// LoadProfile loads a profile from the environment configuration.
func LoadProfile(opts LoadProfileOptions) (LoadProfileResult, error) {
	envLookup := opts.EnvLookup
	if envLookup == nil {
		envLookup = envconfig.EnvLookupOS
	}

	configResult, err := LoadConfig(LoadConfigOptions{
		ConfigFilePath: opts.ConfigFilePath,
		EnvLookup:      envLookup,
	})
	if err != nil {
		return LoadProfileResult{}, err
	}

	profileName := opts.ConfigFileProfile
	if profileName == "" {
		profileName, _ = envLookup.LookupEnv("TEMPORAL_PROFILE")
	}
	if profileName == "" {
		profileName = envconfig.DefaultConfigFileProfile
	}

	profile := configResult.Config.Profiles[profileName]
	if profile == nil {
		profile = &Profile{}
		configResult.Config.Profiles[profileName] = profile
	}

	if !opts.DisableEnv {
		if err := profile.ClientConfigProfile.ApplyEnvVars(envLookup); err != nil {
			return LoadProfileResult{}, err
		}
	}

	return LoadProfileResult{
		Config:         configResult.Config,
		ConfigFilePath: configResult.ConfigFilePath,
		ProfileName:    profileName,
		Profile:        profile,
	}, nil
}

type LoadConfigOptions struct {
	ConfigFilePath string
	EnvLookup      envconfig.EnvLookup
}

type LoadConfigResult struct {
	Config         FileConfig
	ConfigFilePath string
}

// LoadConfig loads the client configuration from the specified file or default location.
func LoadConfig(options LoadConfigOptions) (LoadConfigResult, error) {
	envLookup := options.EnvLookup
	if envLookup == nil {
		envLookup = envconfig.EnvLookupOS
	}

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

	oauthByProfile, err := loadOAuthConfigFromFile(resolvedPath)
	if err != nil {
		return LoadConfigResult{}, err
	}

	profiles := make(map[string]*Profile)
	for name, baseProfile := range clientConfig.Profiles {
		profiles[name] = &Profile{
			ClientConfigProfile: *baseProfile,
			OAuth:               oauthByProfile[name],
		}
	}

	return LoadConfigResult{
		Config:         FileConfig{Profiles: profiles},
		ConfigFilePath: resolvedPath,
	}, nil
}

type WriteConfigOptions struct {
	Config         FileConfig
	ConfigFilePath string
	EnvLookup      envconfig.EnvLookup
}

// ConfigToTOML serializes the configuration to TOML bytes.
func ConfigToTOML(config *FileConfig) ([]byte, error) {
	envConfig := &envconfig.ClientConfig{
		Profiles: make(map[string]*envconfig.ClientConfigProfile),
	}
	for name, p := range config.Profiles {
		if p != nil {
			envConfig.Profiles[name] = &p.ClientConfigProfile
		}
	}

	b, err := envConfig.ToTOML(envconfig.ClientConfigToTOMLOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed building TOML: %w", err)
	}

	var buf bytes.Buffer
	buf.Write(b)

	if err := appendOAuthToTOML(&buf, config.Profiles); err != nil {
		return nil, err
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

	if err := os.MkdirAll(filepath.Dir(configFilePath), 0700); err != nil {
		return fmt.Errorf("failed making config file parent dirs: %w", err)
	}
	if err := os.WriteFile(configFilePath, b, 0600); err != nil {
		return fmt.Errorf("failed writing config file: %w", err)
	}
	return nil
}

// OAuth TOML handling

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

type inlineStringMap map[string]string

func (m inlineStringMap) MarshalTOML() ([]byte, error) {
	if len(m) == 0 {
		return nil, nil
	}
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

func loadOAuthConfigFromFile(path string) (map[string]*cliext.OAuthConfig, error) {
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

	oauthByProfile := make(map[string]*cliext.OAuthConfig)
	for profileName, profile := range raw.Profile {
		if profile == nil || profile.OAuth == nil {
			oauthByProfile[profileName] = nil
			continue
		}
		cfg := profile.OAuth
		oauth := &cliext.OAuthConfig{
			OAuthClientConfig: cliext.OAuthClientConfig{
				ClientID:      cfg.ClientID,
				ClientSecret:  cfg.ClientSecret,
				TokenURL:      cfg.TokenURL,
				AuthURL:       cfg.AuthURL,
				RequestParams: cfg.RequestParams,
				Scopes:        cfg.Scopes,
			},
			OAuthToken: cliext.OAuthToken{
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

func oauthConfigToTOML(oauth *cliext.OAuthConfig) *oauthConfigTOML {
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

func appendOAuthToTOML(buf *bytes.Buffer, profiles map[string]*Profile) error {
	for name, profile := range profiles {
		if profile == nil || profile.OAuth == nil {
			continue
		}
		oauthTOML := oauthConfigToTOML(profile.OAuth)
		oauthBytes, err := toml.Marshal(oauthTOML)
		if err != nil {
			return fmt.Errorf("failed marshaling OAuth config: %w", err)
		}
		fmt.Fprintf(buf, "\n[profile.%s.oauth]\n", name)
		buf.Write(oauthBytes)
	}
	return nil
}
