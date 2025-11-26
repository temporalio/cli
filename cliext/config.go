package cliext

import (
	"go.temporal.io/sdk/contrib/envconfig"
)

// LoadOptions contains options for loading configuration.
type LoadOptions struct {
	// ConfigFilePath is the path to the configuration file.
	// If empty, the default path will be used.
	ConfigFilePath string

	// EnvLookup is used to look up environment variables.
	// If nil, EnvLookupOS will be used.
	EnvLookup EnvLookup
}

// LoadConfig loads the client configuration from the specified file or default location.
func LoadConfig(opts LoadOptions) (*envconfig.ClientConfig, error) {
	envLookup := opts.EnvLookup
	if envLookup == nil {
		envLookup = EnvLookupOS
	}

	clientConfig, err := envconfig.LoadClientConfig(envconfig.LoadClientConfigOptions{
		ConfigFilePath: opts.ConfigFilePath,
		EnvLookup:      envLookup,
	})
	if err != nil {
		return nil, err
	}
	return &clientConfig, nil
}
