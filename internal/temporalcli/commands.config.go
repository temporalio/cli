package temporalcli

import (
	"fmt"
	"sort"

	"github.com/BurntSushi/toml"
	"github.com/temporalio/cli/cliext"
	"github.com/temporalio/cli/internal/printer"
	"go.temporal.io/sdk/contrib/envconfig"
)

func (c *TemporalConfigDeleteCommand) run(cctx *CommandContext, _ []string) error {
	opts := loadOptsFromContext(cctx)
	profileName := envConfigProfileName(cctx)
	conf, confProfile, err := cliext.LoadProfile(opts, profileName, false)
	if err != nil {
		return err
	}

	if err := cliext.DeleteProperty(confProfile, c.Prop); err != nil {
		return err
	}

	cctx.Logger.Info("Writing config file", "file", opts.ConfigFilePath)
	return cliext.WriteConfig(conf, opts)
}

func (c *TemporalConfigDeleteProfileCommand) run(cctx *CommandContext, _ []string) error {
	opts := loadOptsFromContext(cctx)
	profileName := envConfigProfileName(cctx)
	conf, _, err := cliext.LoadProfile(opts, profileName, false)
	if err != nil {
		return err
	}
	// To make extra sure they meant to do this, we require the profile name
	// as an explicit CLI arg. This prevents accidentally deleting the
	// "default" profile.
	if cctx.RootCommand.Profile == "" {
		return fmt.Errorf("to delete an entire profile, --profile must be provided explicitly")
	}
	delete(conf.Profiles, profileName)

	cctx.Logger.Info("Writing config file", "file", opts.ConfigFilePath)
	return cliext.WriteConfig(conf, opts)
}

func (c *TemporalConfigGetCommand) run(cctx *CommandContext, _ []string) error {
	opts := loadOptsFromContext(cctx)
	profileName := envConfigProfileName(cctx)
	conf, confProfile, err := cliext.LoadProfile(opts, profileName, false)
	if err != nil {
		return err
	}
	type prop struct {
		Property string `json:"property"`
		Value    any    `json:"value"`
	}
	// If there is a specific key requested, show it, otherwise show all
	if c.Prop != "" {
		// We do not support asking for structures with children at this time,
		// but "tls" is a special case because it's also a bool.
		if c.Prop == "codec" || c.Prop == "grpc_meta" {
			return fmt.Errorf("must provide exact property, not parent property")
		}
		val, err := cliext.GetPropertyValue(confProfile, c.Prop)
		if err != nil {
			return err
		}
		return cctx.Printer.PrintStructured(
			prop{Property: c.Prop, Value: val},
			printer.StructuredOptions{Table: &printer.TableOptions{}},
		)
	} else if cctx.JSONOutput {
		// If it is JSON and not prop specific, we want to dump the TOML
		// structure in JSON form
		var tomlConf struct {
			Profiles map[string]any `toml:"profile"`
		}
		if b, err := conf.ToTOML(envconfig.ClientConfigToTOMLOptions{}); err != nil {
			return fmt.Errorf("failed converting to TOML: %w", err)
		} else if err := toml.Unmarshal(b, &tomlConf); err != nil {
			return fmt.Errorf("failed converting from TOML: %w", err)
		}
		return cctx.Printer.PrintStructured(tomlConf.Profiles[profileName], printer.StructuredOptions{})
	} else {
		// Get every property individually as a property-value pair except zero vals
		propsMap, err := cliext.ListProperties(confProfile)
		if err != nil {
			return err
		}
		var props []prop
		for k, v := range propsMap {
			props = append(props, prop{Property: k, Value: v})
		}

		// Sort and display
		sort.Slice(props, func(i, j int) bool { return props[i].Property < props[j].Property })
		return cctx.Printer.PrintStructured(props, printer.StructuredOptions{Table: &printer.TableOptions{}})
	}
}

func (c *TemporalConfigListCommand) run(cctx *CommandContext, _ []string) error {
	opts := loadOptsFromContext(cctx)
	config, err := cliext.LoadConfig(opts)
	if err != nil {
		return err
	}
	type profile struct {
		Name string `json:"name"`
	}
	profiles := make([]profile, 0, len(config.Profiles))
	for k := range config.Profiles {
		profiles = append(profiles, profile{Name: k})
	}
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].Name < profiles[j].Name })
	return cctx.Printer.PrintStructured(profiles, printer.StructuredOptions{Table: &printer.TableOptions{}})
}

func (c *TemporalConfigSetCommand) run(cctx *CommandContext, _ []string) error {
	opts := loadOptsFromContext(cctx)
	profileName := envConfigProfileName(cctx)
	conf, confProfile, err := cliext.LoadProfile(opts, profileName, true)
	if err != nil {
		return err
	}

	if err := cliext.SetPropertyValue(confProfile, c.Prop, c.Value); err != nil {
		return err
	}

	cctx.Logger.Info("Writing config file", "file", opts.ConfigFilePath)
	return cliext.WriteConfig(conf, opts)
}

func envConfigProfileName(cctx *CommandContext) string {
	if cctx.RootCommand.Profile != "" {
		return cctx.RootCommand.Profile
	} else if p, _ := cctx.Options.EnvLookup.LookupEnv("TEMPORAL_PROFILE"); p != "" {
		return p
	}
	return envconfig.DefaultConfigFileProfile
}

func loadOptsFromContext(cctx *CommandContext) cliext.LoadOptions {
	return cliext.LoadOptions{
		ConfigFilePath: cctx.RootCommand.ConfigFile,
		EnvLookup:      cctx.Options.EnvLookup,
	}
}
