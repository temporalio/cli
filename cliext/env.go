package cliext

import "os"

// EnvLookup is an interface for looking up environment variables.
// This matches the envconfig.EnvLookup interface.
type EnvLookup interface {
	LookupEnv(key string) (string, bool)
	Environ() []string
}

// EnvLookupOS is the default EnvLookup implementation that uses os.LookupEnv
// and os.Environ.
var EnvLookupOS EnvLookup = envLookupOS{}

// envLookupOS implements EnvLookup using os.LookupEnv and os.Environ.
type envLookupOS struct{}

func (envLookupOS) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

func (envLookupOS) Environ() []string {
	return os.Environ()
}
