package temporalcli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/temporalio/cli/temporalcli/internal/printer"
)

func (c *TemporalEnvCommand) envNameAndKey(cctx *CommandContext, args[] string, keyFlag string) (string, string, error) {
	if len(args) > 0 {
		cctx.Logger.Warn("Arguments to env commands are deprecated; please use --env and --key (or -k) instead")

		if (c.Parent.Env != "default" || keyFlag != "") {
			return "", "", fmt.Errorf("cannot specify both an argument and flags; please use flags instead")
		}

		keyPieces := strings.Split(args[0], ".")
		switch len(keyPieces) {
		case 0:
			return "", "", fmt.Errorf("no env or property name specified")
		case 1:
			return keyPieces[0], "", nil
		case 2:
			return keyPieces[0], keyPieces[1], nil
		default:
			return "", "", fmt.Errorf("property name may not contain dots")
		}
	}

	if strings.Contains(keyFlag, ".") {
		return "", "", fmt.Errorf("property name may not contain dots")
	}

	return c.Parent.Env, keyFlag, nil
}

func (c *TemporalEnvDeleteCommand) run(cctx *CommandContext, args []string) error {
	envName, key, err := c.Parent.envNameAndKey(cctx, args, c.Key)
	if err != nil {
		return err
	}

	// Env must be present
	env, ok := cctx.EnvConfigValues[envName]
	if !ok {
		return fmt.Errorf("env %q not found", envName)
	}
	// User can remove single flag or all in env
	if key != "" {
		cctx.Logger.Info("Deleting env property", "env", envName, "property", key)
		delete(env, key)
	} else {
		cctx.Logger.Info("Deleting env", "env", env)
		delete(cctx.EnvConfigValues, envName)
	}
	return cctx.WriteEnvConfigToFile()
}

func (c *TemporalEnvGetCommand) run(cctx *CommandContext, args []string) error {
	envName, key, err := c.Parent.envNameAndKey(cctx, args, c.Key)
	if err != nil {
		return err
	}

	// Env must be present
	env, ok := cctx.EnvConfigValues[envName]
	if !ok {
		return fmt.Errorf("env %q not found", envName)
	}
	type prop struct {
		Property string `json:"property"`
		Value    string `json:"value"`
	}
	var props []prop
	// User can ask for single flag or all in env
	if key != "" {
		props = []prop{{Property: key, Value: env[key]}}
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
	envName, key, err := c.Parent.envNameAndKey(cctx, args, c.Key)
	if err != nil {
		return err
	}

	if key == "" {
		return fmt.Errorf("property name must be specified with -k")
	}

	value := c.Value
	switch len(args) {
	case 0:
		// Use what's in the flag
	case 1:
		// We got an "env.name" argument passed above, but no "value" argument
		return fmt.Errorf("no value provided; see --help")
	case 2:
		// Old-style syntax; pull the value out of the args
		value = args[1]
	default:
		// Cobra should catch this / we should never get here; included for
		// completeness anyway.
		return fmt.Errorf("too many arguments provided; see --help")
	}

	if cctx.EnvConfigValues == nil {
		cctx.EnvConfigValues = map[string]map[string]string{}
	}
	if cctx.EnvConfigValues[envName] == nil {
		cctx.EnvConfigValues[envName] = map[string]string{}
	}
	cctx.Logger.Info("Setting env property", "env", envName, "property", key, "value", value)
	cctx.EnvConfigValues[envName][key] = value
	return cctx.WriteEnvConfigToFile()
}
