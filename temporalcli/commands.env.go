package temporalcli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/temporalio/cli/temporalcli/internal/printer"
)

func (c *TemporalEnvDeleteCommand) run(cctx *CommandContext, args []string) error {
	keyPieces := strings.Split(args[0], ".")
	if len(keyPieces) > 2 {
		return fmt.Errorf("env property key to delete cannot have more than one dot")
	}
	// Env must be present (but flag itself doesn't have to be)
	env, ok := cctx.EnvConfigValues[keyPieces[0]]
	if !ok {
		return fmt.Errorf("env %q not found", keyPieces[0])
	}
	// User can remove single flag or all in env
	if len(keyPieces) > 1 {
		cctx.Logger.Info("Deleting env property", "env", keyPieces[0], "property", keyPieces[1])
		delete(env, keyPieces[1])
	} else {
		cctx.Logger.Info("Deleting env", "env", keyPieces[0])
		delete(cctx.EnvConfigValues, keyPieces[0])
	}
	return cctx.WriteEnvConfigToFile()
}

func (c *TemporalEnvGetCommand) run(cctx *CommandContext, args []string) error {
	keyPieces := strings.Split(args[0], ".")
	if len(keyPieces) > 2 {
		return fmt.Errorf("env key to get cannot have more than one dot")
	}
	// Env must be present (but flag itself doesn't have to me)
	env, ok := cctx.EnvConfigValues[keyPieces[0]]
	if !ok {
		return fmt.Errorf("env %q not found", keyPieces[0])
	}
	type prop struct {
		Property string `json:"property"`
		Value    string `json:"value"`
	}
	var props []prop
	// User can ask for single flag or all in env
	if len(keyPieces) > 1 {
		props = []prop{{Property: keyPieces[1], Value: env[keyPieces[1]]}}
	} else {
		props = make([]prop, 0, len(env))
		for k, v := range env {
			props = append(props, prop{Property: k, Value: v})
		}
		sort.Slice(props, func(i, j int) bool { return props[i].Property < props[j].Property })
	}
	// Print as table
	return cctx.Printer.PrintStructured(props, printer.StructuredOptions{Table: &printer.TableOptions{}})
}

func (c *TemporalEnvListCommand) run(cctx *CommandContext, args []string) error {
	type env struct {
		Name string `json:"name"`
	}
	envs := make([]env, 0, len(cctx.EnvConfigValues))
	for k := range cctx.EnvConfigValues {
		envs = append(envs, env{Name: k})
	}
	// Print as table
	return cctx.Printer.PrintStructured(envs, printer.StructuredOptions{Table: &printer.TableOptions{}})
}

func (c *TemporalEnvSetCommand) run(cctx *CommandContext, args []string) error {
	keyPieces := strings.Split(args[0], ".")
	if len(keyPieces) != 2 {
		return fmt.Errorf("env property key to set must have single dot separating env and value")
	}
	if cctx.EnvConfigValues == nil {
		cctx.EnvConfigValues = map[string]map[string]string{}
	}
	if cctx.EnvConfigValues[keyPieces[0]] == nil {
		cctx.EnvConfigValues[keyPieces[0]] = map[string]string{}
	}
	cctx.Logger.Info("Setting env property", "env", keyPieces[0], "property", keyPieces[1], "value", args[1])
	cctx.EnvConfigValues[keyPieces[0]][keyPieces[1]] = args[1]
	return cctx.WriteEnvConfigToFile()
}
