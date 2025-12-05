package cliext

import "os"

// EnvLookupOS is the default EnvLookup implementation.
var EnvLookupOS EnvLookup = envLookupOS{}

type EnvLookup interface {
	LookupEnv(key string) (string, bool)
	Environ() []string
}

type envLookupOS struct{}

func (envLookupOS) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

func (envLookupOS) Environ() []string {
	return os.Environ()
}
