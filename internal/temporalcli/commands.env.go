package temporalcli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/temporalio/cli/internal/printer"
	"gopkg.in/yaml.v3"
)

func (c *TemporalEnvCommand) envNameAndKey(cctx *CommandContext, args []string, keyFlag string) (string, string, error) {
	if len(args) > 0 {
		fmt.Fprintln(cctx.Options.Stderr, "Warning: Arguments to env commands are deprecated; please use --env and --key (or -k) instead")

		if c.Parent.Env != "default" || keyFlag != "" {
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

	// Env is guaranteed to already be present
	env, _ := cctx.DeprecatedEnvConfigValues[envName]
	// User can remove single flag or all in env
	if key != "" {
		cctx.Logger.Info("Deleting env property", "env", envName, "property", key)
		delete(env, key)
	} else {
		cctx.Logger.Info("Deleting env", "env", env)
		delete(cctx.DeprecatedEnvConfigValues, envName)
	}
	return writeDeprecatedEnvConfigToFile(cctx)
}

func (c *TemporalEnvGetCommand) run(cctx *CommandContext, args []string) error {
	envName, key, err := c.Parent.envNameAndKey(cctx, args, c.Key)
	if err != nil {
		return err
	}

	// Env is guaranteed to already be present
	env, _ := cctx.DeprecatedEnvConfigValues[envName]
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
	envs := make([]env, 0, len(cctx.DeprecatedEnvConfigValues))
	for k := range cctx.DeprecatedEnvConfigValues {
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

	if cctx.DeprecatedEnvConfigValues == nil {
		cctx.DeprecatedEnvConfigValues = map[string]map[string]string{}
	}
	if cctx.DeprecatedEnvConfigValues[envName] == nil {
		cctx.DeprecatedEnvConfigValues[envName] = map[string]string{}
	}
	cctx.Logger.Info("Setting env property", "env", envName, "property", key, "value", value)
	cctx.DeprecatedEnvConfigValues[envName][key] = value
	return writeDeprecatedEnvConfigToFile(cctx)
}

func writeDeprecatedEnvConfigToFile(cctx *CommandContext) error {
	if cctx.Options.DeprecatedEnvConfig.EnvConfigFile == "" {
		return fmt.Errorf("unable to find place for env file (unknown HOME dir)")
	}
	cctx.Logger.Info("Writing env file", "file", cctx.Options.DeprecatedEnvConfig.EnvConfigFile)
	return writeDeprecatedEnvConfigFile(cctx.Options.DeprecatedEnvConfig.EnvConfigFile, cctx.DeprecatedEnvConfigValues)
}

// May be empty result if can't get user home dir
func defaultDeprecatedEnvConfigFile(appName, configName string) string {
	// No env file if no $HOME
	if dir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(dir, ".config", appName, configName+".yaml")
	}
	return ""
}

func readDeprecatedEnvConfigFile(file string) (env map[string]map[string]string, err error) {
	b, err := os.ReadFile(file)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed reading env file: %w", err)
	}
	var m map[string]map[string]map[string]string
	if err := yaml.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("failed unmarshalling env YAML: %w", err)
	}
	return m["env"], nil
}

func writeDeprecatedEnvConfigFile(file string, env map[string]map[string]string) error {
	b, err := yaml.Marshal(map[string]any{"env": env})
	if err != nil {
		return fmt.Errorf("failed marshaling YAML: %w", err)
	}
	// Make parent directories as needed
	if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
		return fmt.Errorf("failed making env file parent dirs: %w", err)
	} else if err := os.WriteFile(file, b, 0600); err != nil {
		return fmt.Errorf("failed writing env file: %w", err)
	}
	return nil
}
