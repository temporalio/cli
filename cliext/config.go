package cliext

import (
	"fmt"
	"os"
	"path/filepath"

	"go.temporal.io/sdk/contrib/envconfig"
)

// LoadConfigOptions contains options for loading configuration.
type LoadConfigOptions struct {
	// ConfigFilePath is the path to the configuration file.
	// If empty, TEMPORAL_CONFIG_FILE env var is checked, then the default path is used.
	ConfigFilePath string

	// EnvLookup is used for environment variable lookups.
	// If nil, os.LookupEnv is used.
	EnvLookup EnvLookup
}

// LoadConfigResult contains the result of loading configuration.
type LoadConfigResult struct {
	// Config is the loaded configuration.
	Config *envconfig.ClientConfig

	// ConfigFilePath is the resolved path to the configuration file that was loaded.
	// This may differ from the input if TEMPORAL_CONFIG_FILE env var was used.
	ConfigFilePath string
}

// LoadConfig loads the client configuration from the specified file or default location.
// If ConfigFilePath is empty, the TEMPORAL_CONFIG_FILE environment variable is checked.
func LoadConfig(options LoadConfigOptions) (LoadConfigResult, error) {
	envLookup := options.EnvLookup
	if envLookup == nil {
		envLookup = EnvLookupOS
	}
	configFilePath := options.ConfigFilePath
	if configFilePath == "" {
		configFilePath, _ = envLookup.LookupEnv("TEMPORAL_CONFIG_FILE")
	}
	clientConfig, err := envconfig.LoadClientConfig(envconfig.LoadClientConfigOptions{
		ConfigFilePath: configFilePath,
		EnvLookup:      envLookup,
	})
	if err != nil {
		return LoadConfigResult{}, err
	}
	return LoadConfigResult{
		Config:         &clientConfig,
		ConfigFilePath: configFilePath,
	}, nil
}

// WriteConfig writes the configuration to the specified file or default location.
// If configFilePath is empty, the default path will be used.
func WriteConfig(config *envconfig.ClientConfig, configFilePath string) error {
	// Get file
	if configFilePath == "" {
		var err error
		if configFilePath, err = envconfig.DefaultConfigFilePath(); err != nil {
			return err
		}
	}

	// Convert to TOML
	b, err := config.ToTOML(envconfig.ClientConfigToTOMLOptions{})
	if err != nil {
		return fmt.Errorf("failed building TOML: %w", err)
	}

	// Write to file, making dirs as needed
	if err := os.MkdirAll(filepath.Dir(configFilePath), 0700); err != nil {
		return fmt.Errorf("failed making config file parent dirs: %w", err)
	}
	if err := os.WriteFile(configFilePath, b, 0600); err != nil {
		return fmt.Errorf("failed writing config file: %w", err)
	}
	return nil
}
