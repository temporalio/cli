package cliext

import (
	"fmt"

	"go.temporal.io/sdk/contrib/envconfig"
)

// LoadProfileOptions contains options for loading a profile.
type LoadProfileOptions struct {
	// ConfigFilePath is the path to the configuration file.
	// If empty, TEMPORAL_CONFIG_FILE env var is checked, then the default path is used.
	ConfigFilePath string

	// ProfileName is the name of the profile to load.
	// If empty, TEMPORAL_PROFILE env var is checked, then the default profile is used.
	ProfileName string

	// CreateIfMissing creates an empty profile if it doesn't exist.
	CreateIfMissing bool

	// EnvLookup is used for environment variable lookups.
	// If nil, os.LookupEnv is used.
	EnvLookup EnvLookup
}

// LoadProfileResult contains the result of loading a profile.
type LoadProfileResult struct {
	// Config is the loaded configuration.
	Config *envconfig.ClientConfig

	// ConfigFilePath is the resolved path to the configuration file.
	ConfigFilePath string

	// Profile is the loaded profile.
	Profile *envconfig.ClientConfigProfile

	// ProfileName is the resolved profile name.
	ProfileName string
}

// LoadProfile loads a specific profile from the configuration.
func LoadProfile(opts LoadProfileOptions) (LoadProfileResult, error) {
	envLookup := opts.EnvLookup
	if envLookup == nil {
		envLookup = EnvLookupOS
	}

	configResult, err := LoadConfig(LoadConfigOptions{
		ConfigFilePath: opts.ConfigFilePath,
		EnvLookup:      envLookup,
	})
	if err != nil {
		return LoadProfileResult{}, err
	}

	profileName := opts.ProfileName
	if profileName == "" {
		profileName, _ = envLookup.LookupEnv("TEMPORAL_PROFILE")
	}
	if profileName == "" {
		profileName = envconfig.DefaultConfigFileProfile
	}

	profile := configResult.Config.Profiles[profileName]
	if profile == nil {
		if !opts.CreateIfMissing {
			return LoadProfileResult{}, fmt.Errorf("profile %q not found", profileName)
		}
		profile = &envconfig.ClientConfigProfile{}
		configResult.Config.Profiles[profileName] = profile
	}
	return LoadProfileResult{
		Config:         configResult.Config,
		Profile:        profile,
		ConfigFilePath: configResult.ConfigFilePath,
		ProfileName:    profileName,
	}, nil
}
