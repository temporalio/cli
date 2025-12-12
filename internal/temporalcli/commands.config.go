package temporalcli

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/temporalio/cli/cliext"
	"github.com/temporalio/cli/internal/printer"
	"go.temporal.io/sdk/contrib/envconfig"
)

func (c *TemporalConfigDeleteCommand) run(cctx *CommandContext, _ []string) error {
	// Load config
	profileName := envConfigProfileName(cctx)
	conf, confProfile, err := loadEnvConfigProfile(cctx, profileName, true)
	if err != nil {
		return err
	}
	if strings.HasPrefix(c.Prop, "grpc_meta.") {
		key := strings.TrimPrefix(c.Prop, "grpc_meta.")
		if _, ok := confProfile.GRPCMeta[key]; !ok {
			return fmt.Errorf("property %q not found", c.Prop)
		}
		delete(confProfile.GRPCMeta, key)
	} else if strings.HasPrefix(c.Prop, "oauth.request_params.") {
		key := strings.TrimPrefix(c.Prop, "oauth.request_params.")
		if confProfile.OAuth == nil || confProfile.OAuth.RequestParams == nil {
			return fmt.Errorf("property %q not found", c.Prop)
		}
		if _, ok := confProfile.OAuth.RequestParams[key]; !ok {
			return fmt.Errorf("property %q not found", c.Prop)
		}
		delete(confProfile.OAuth.RequestParams, key)
	} else {
		reflectVal, err := reflectEnvConfigProp(confProfile, c.Prop, true)
		if err != nil {
			return err
		}
		reflectVal.SetZero()
	}

	// Save
	return writeEnvConfigFile(cctx, conf)
}

func (c *TemporalConfigDeleteProfileCommand) run(cctx *CommandContext, _ []string) error {
	// Load config
	profileName := envConfigProfileName(cctx)
	conf, _, err := loadEnvConfigProfile(cctx, profileName, true)
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

	// Save
	return writeEnvConfigFile(cctx, conf)
}

func (c *TemporalConfigGetCommand) run(cctx *CommandContext, _ []string) error {
	// Load config profile
	profileName := envConfigProfileName(cctx)
	_, confProfile, err := loadEnvConfigProfile(cctx, profileName, true)
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
		if c.Prop == "codec" || c.Prop == "grpc_meta" || c.Prop == "oauth" || c.Prop == "oauth.request_params" {
			return fmt.Errorf("must provide exact property, not parent property")
		}
		var reflectVal reflect.Value
		// gRPC meta and OAuth request params are special
		if strings.HasPrefix(c.Prop, "grpc_meta.") {
			v, ok := confProfile.GRPCMeta[strings.TrimPrefix(c.Prop, "grpc_meta.")]
			if !ok {
				return fmt.Errorf("unknown property %q", c.Prop)
			}
			reflectVal = reflect.ValueOf(v)
		} else if strings.HasPrefix(c.Prop, "oauth.request_params.") {
			if confProfile.OAuth == nil || confProfile.OAuth.RequestParams == nil {
				return fmt.Errorf("unknown property %q", c.Prop)
			}
			v, ok := confProfile.OAuth.RequestParams[strings.TrimPrefix(c.Prop, "oauth.request_params.")]
			if !ok {
				return fmt.Errorf("unknown property %q", c.Prop)
			}
			reflectVal = reflect.ValueOf(v)
		} else {
			// Single value goes into property-value structure
			reflectVal, err = reflectEnvConfigProp(confProfile, c.Prop, false)
			if err != nil {
				return err
			}
			// Pointers become true/false
			if reflectVal.Kind() == reflect.Pointer {
				reflectVal = reflect.ValueOf(!reflectVal.IsNil())
			}
		}
		return cctx.Printer.PrintStructured(
			prop{Property: c.Prop, Value: reflectVal.Interface()},
			printer.StructuredOptions{Table: &printer.TableOptions{}},
		)
	} else if cctx.JSONOutput {
		// If it is JSON and not prop specific, we want to dump the TOML
		// structure in JSON form
		var tomlConf struct {
			Profiles map[string]any `toml:"profile"`
		}
		cliextConfig := &cliext.ClientConfig{
			Profiles: map[string]*cliext.Profile{
				profileName: confProfile,
			},
		}
		if b, err := cliext.ConfigToTOML(cliextConfig); err != nil {
			return fmt.Errorf("failed converting to TOML: %w", err)
		} else if err := toml.Unmarshal(b, &tomlConf); err != nil {
			return fmt.Errorf("failed converting from TOML: %w", err)
		}
		return cctx.Printer.PrintStructured(tomlConf.Profiles[profileName], printer.StructuredOptions{})
	} else {
		// Get every property individually as a property-value pair except zero
		// vals
		var props []prop
		for k := range envConfigPropsToFieldNames {
			// TLS is a special case
			if k == "tls" {
				if confProfile.TLS != nil {
					props = append(props, prop{Property: "tls", Value: true})
				}
				continue
			}
			if val, err := reflectEnvConfigProp(confProfile, k, false); err != nil {
				return err
			} else if !val.IsZero() {
				props = append(props, prop{Property: k, Value: val.Interface()})
			}
		}

		// Add grpc_meta
		for k, v := range confProfile.GRPCMeta {
			props = append(props, prop{Property: "grpc_meta." + k, Value: v})
		}

		// Add oauth.request_params
		if confProfile.OAuth != nil {
			for k, v := range confProfile.OAuth.RequestParams {
				props = append(props, prop{Property: "oauth.request_params." + k, Value: v})
			}
		}

		// Sort and display
		sort.Slice(props, func(i, j int) bool { return props[i].Property < props[j].Property })
		return cctx.Printer.PrintStructured(props, printer.StructuredOptions{Table: &printer.TableOptions{}})
	}
}

func (c *TemporalConfigListCommand) run(cctx *CommandContext, _ []string) error {
	loadResult, err := cliext.LoadConfig(cliext.LoadConfigOptions{
		ConfigFilePath: cctx.RootCommand.ConfigFile,
		EnvLookup:      cctx.Options.EnvLookup,
	})
	if err != nil {
		return err
	}
	type profile struct {
		Name string `json:"name"`
	}
	profiles := make([]profile, 0, len(loadResult.Config.Profiles))
	for k := range loadResult.Config.Profiles {
		profiles = append(profiles, profile{Name: k})
	}
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].Name < profiles[j].Name })
	return cctx.Printer.PrintStructured(profiles, printer.StructuredOptions{Table: &printer.TableOptions{}})
}

func (c *TemporalConfigSetCommand) run(cctx *CommandContext, _ []string) error {
	// Load config
	conf, confProfile, err := loadEnvConfigProfile(cctx, envConfigProfileName(cctx), false)
	if err != nil {
		return err
	}
	// gRPC meta and OAuth request params are handled specifically
	if strings.HasPrefix(c.Prop, "grpc_meta.") {
		if confProfile.GRPCMeta == nil {
			confProfile.GRPCMeta = map[string]string{}
		}
		confProfile.GRPCMeta[strings.TrimPrefix(c.Prop, "grpc_meta.")] = c.Value
	} else if strings.HasPrefix(c.Prop, "oauth.request_params.") {
		if confProfile.OAuth == nil {
			confProfile.OAuth = &cliext.OAuthConfig{}
		}
		if confProfile.OAuth.RequestParams == nil {
			confProfile.OAuth.RequestParams = map[string]string{}
		}
		confProfile.OAuth.RequestParams[strings.TrimPrefix(c.Prop, "oauth.request_params.")] = c.Value
	} else {
		// Get reflect value
		reflectVal, err := reflectEnvConfigProp(confProfile, c.Prop, false)
		if err != nil {
			return err
		}
		// Set it from string
		switch reflectVal.Kind() {
		case reflect.String:
			reflectVal.SetString(c.Value)
		case reflect.Pointer:
			// Used for "tls", true makes an empty object, false sets nil
			switch c.Value {
			case "true":
				// Only set if not set
				if reflectVal.IsZero() {
					reflectVal.Set(reflect.New(reflectVal.Type().Elem()))
				}
			case "false":
				reflectVal.SetZero()
			default:
				return fmt.Errorf("must be 'true' or 'false' to set this property")
			}
		case reflect.Slice:
			switch reflectVal.Type().Elem().Kind() {
			case reflect.Uint8:
				// []byte - set as bytes
				reflectVal.SetBytes([]byte(c.Value))
			case reflect.String:
				// []string - split by comma
				if c.Value == "" {
					reflectVal.Set(reflect.MakeSlice(reflectVal.Type(), 0, 0))
				} else {
					parts := strings.Split(c.Value, ",")
					for i := range parts {
						parts[i] = strings.TrimSpace(parts[i])
					}
					reflectVal.Set(reflect.ValueOf(parts))
				}
			default:
				return fmt.Errorf("unexpected slice of type %v", reflectVal.Type())
			}
		case reflect.Bool:
			if c.Value != "true" && c.Value != "false" {
				return fmt.Errorf("must be 'true' or 'false' to set this property")
			}
			reflectVal.SetBool(c.Value == "true")
		case reflect.Map:
			return fmt.Errorf("must set each individual value of a map")
		default:
			return fmt.Errorf("unexpected type %v", reflectVal.Type())
		}
	}

	// Save
	return writeEnvConfigFile(cctx, conf)
}

func envConfigProfileName(cctx *CommandContext) string {
	if cctx.RootCommand.Profile != "" {
		return cctx.RootCommand.Profile
	} else if p, _ := cctx.Options.EnvLookup.LookupEnv("TEMPORAL_PROFILE"); p != "" {
		return p
	}
	return envconfig.DefaultConfigFileProfile
}

func loadEnvConfigProfile(
	cctx *CommandContext,
	profile string,
	failIfNotFound bool,
) (cliext.ClientConfig, *cliext.Profile, error) {
	loadResult, err := cliext.LoadConfig(cliext.LoadConfigOptions{
		ConfigFilePath: cctx.RootCommand.ConfigFile,
		EnvLookup:      cctx.Options.EnvLookup,
	})
	if err != nil {
		return cliext.ClientConfig{}, nil, err
	}

	// Load profile
	clientProfile := loadResult.Config.Profiles[profile]
	if clientProfile == nil {
		if failIfNotFound {
			return cliext.ClientConfig{}, nil, fmt.Errorf("profile %q not found", profile)
		}
		clientProfile = &cliext.Profile{}
		loadResult.Config.Profiles[profile] = clientProfile
	}
	return loadResult.Config, clientProfile, nil
}

var envConfigPropsToFieldNames = map[string]string{
	"address":                       "Address",
	"namespace":                     "Namespace",
	"api_key":                       "APIKey",
	"tls":                           "TLS",
	"tls.disabled":                  "Disabled",
	"tls.client_cert_path":          "ClientCertPath",
	"tls.client_cert_data":          "ClientCertData",
	"tls.client_key_path":           "ClientKeyPath",
	"tls.client_key_data":           "ClientKeyData",
	"tls.server_ca_cert_path":       "ServerCACertPath",
	"tls.server_ca_cert_data":       "ServerCACertData",
	"tls.server_name":               "ServerName",
	"tls.disable_host_verification": "DisableHostVerification",
	"codec.endpoint":                "Endpoint",
	"codec.auth":                    "Auth",
	"oauth.client_id":        "ClientID",
	"oauth.client_secret":    "ClientSecret",
	"oauth.auth_url":         "AuthURL",
	"oauth.token_url":        "TokenURL",
	"oauth.scopes":           "Scopes",
	"oauth.access_token":     "AccessToken",
	"oauth.refresh_token":    "RefreshToken",
	"oauth.token_type":       "TokenType",
	"oauth.expires_at":       "AccessTokenExpiresAt",
}

func reflectEnvConfigProp(
	prof *cliext.Profile,
	prop string,
	failIfParentNotFound bool,
) (reflect.Value, error) {
	// Get field name
	field := envConfigPropsToFieldNames[prop]
	if field == "" {
		return reflect.Value{}, fmt.Errorf("unknown property %q", prop)
	}

	// Load reflect val
	parentVal := reflect.ValueOf(&prof.ClientConfigProfile).Elem()
	if strings.HasPrefix(prop, "tls.") {
		if prof.TLS == nil {
			if failIfParentNotFound {
				return reflect.Value{}, fmt.Errorf("no TLS options found")
			}
			prof.TLS = &envconfig.ClientConfigTLS{}
		}
		parentVal = reflect.ValueOf(prof.TLS).Elem()
	} else if strings.HasPrefix(prop, "codec.") {
		if prof.Codec == nil {
			if failIfParentNotFound {
				return reflect.Value{}, fmt.Errorf("no codec options found")
			}
			prof.Codec = &envconfig.ClientConfigCodec{}
		}
		parentVal = reflect.ValueOf(prof.Codec).Elem()
	} else if strings.HasPrefix(prop, "oauth.") {
		if prof.OAuth == nil {
			if failIfParentNotFound {
				return reflect.Value{}, fmt.Errorf("no OAuth options found")
			}
			prof.OAuth = &cliext.OAuthConfig{}
		}
		parentVal = reflect.ValueOf(prof.OAuth).Elem()
	}
	return parentVal.FieldByName(field), nil
}

func writeEnvConfigFile(cctx *CommandContext, conf cliext.ClientConfig) error {
	configFile := cctx.RootCommand.ConfigFile
	cctx.Logger.Info("Writing config file", "file", configFile)
	return cliext.WriteConfig(cliext.WriteConfigOptions{
		Config:         conf,
		ConfigFilePath: configFile,
		EnvLookup:      cctx.Options.EnvLookup,
	})
}
