package cliext

import (
	"go.temporal.io/sdk/contrib/envconfig"
)

type Profile struct {
	envconfig.ClientConfigProfile
	OAuth *OAuthConfig
}

type LoadProfileOptions struct {
	envconfig.LoadClientConfigProfileOptions
}

type LoadProfileResult struct {
	// Config is the loaded configuration.
	Config ClientConfig
	// ConfigFilePath is the resolved path to the configuration file.
	ConfigFilePath string
	// ProfileName is the resolved profile name.
	ProfileName string
	// Profile points to Config.Profiles[ProfileName].
	Profile *Profile
}

// LoadProfile loads a profile from the environment configuration.
func LoadProfile(opts LoadProfileOptions) (LoadProfileResult, error) {
	envLookup := opts.EnvLookup
	if envLookup == nil {
		envLookup = envconfig.EnvLookupOS
	}

	// Load the full configuration first to get all profiles.
	configResult, err := LoadConfig(LoadConfigOptions{
		ConfigFilePath: opts.ConfigFilePath,
		EnvLookup:      envLookup,
	})
	if err != nil {
		return LoadProfileResult{}, err
	}

	// Determine the profile name.
	profileName := opts.ConfigFileProfile
	if profileName == "" {
		profileName, _ = envLookup.LookupEnv("TEMPORAL_PROFILE")
	}
	if profileName == "" {
		profileName = envconfig.DefaultConfigFileProfile
	}

	// Get or create the profile in the full config.
	profile := configResult.Config.Profiles[profileName]
	if profile == nil {
		profile = &Profile{}
		configResult.Config.Profiles[profileName] = profile
	}

	// Apply environment variable overrides to the profile (unless disabled).
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
