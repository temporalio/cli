package cliext

import (
	"fmt"
	"os"
	"path/filepath"

	"go.temporal.io/sdk/contrib/envconfig"
)

// LoadConfig loads the client configuration from the specified file or default location.
// If configFilePath is empty, the default path will be used.
func LoadConfig(configFilePath string) (*envconfig.ClientConfig, error) {
	clientConfig, err := envconfig.LoadClientConfig(envconfig.LoadClientConfigOptions{
		ConfigFilePath: configFilePath,
	})
	if err != nil {
		return nil, err
	}
	return &clientConfig, nil
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
