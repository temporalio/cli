package temporalcli

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/temporalio/cli/internal/printer"
	"go.temporal.io/sdk/contrib/envconfig"
)

func (c *TemporalConfigDeleteCommand) run(cctx *CommandContext, _ []string) error {
	if c.Prop == updateCheckConfigProp {
		configFile, _ := resolveConfigFile(cctx.Options)
		return setUpdateCheckEnabled(configFile, false)
	}
	// Load config
	profileName := envConfigProfileName(cctx)
	conf, confProfile, err := loadEnvConfigProfile(cctx, profileName, true)
	if err != nil {
		return err
	}
	if strings.HasPrefix(c.Prop, "grpc_meta.") {
		key := strings.TrimPrefix(c.Prop, "grpc_meta.")
		if _, ok := confProfile.GRPCMeta[key]; !ok {
			return fmt.Errorf("gRPC meta key %q not found", key)
		}
		delete(confProfile.GRPCMeta, key)
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
	if c.Prop == updateCheckConfigProp {
		configFile, _ := resolveConfigFile(cctx.Options)
		_, state, err := loadUpdateCheckState(configFile)
		if err != nil {
			return err
		}
		type prop struct {
			Property string `json:"property"`
			Value    any    `json:"value"`
		}
		return cctx.Printer.PrintStructured(
			prop{Property: c.Prop, Value: state.Enabled},
			printer.StructuredOptions{Table: &printer.TableOptions{}},
		)
	}
	// Load config profile
	profileName := envConfigProfileName(cctx)
	conf, confProfile, err := loadEnvConfigProfile(cctx, profileName, true)
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
		var reflectVal reflect.Value
		// gRPC meta is special
		if strings.HasPrefix(c.Prop, "grpc_meta.") {
			v, ok := confProfile.GRPCMeta[strings.TrimPrefix(c.Prop, "grpc_meta.")]
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
		if b, err := conf.ToTOML(envconfig.ClientConfigToTOMLOptions{}); err != nil {
			return fmt.Errorf("failed converting to TOML: %w", err)
		} else if err := toml.Unmarshal(b, &tomlConf); err != nil {
			return fmt.Errorf("failed converting from TOML: %w", err)
		}
		return cctx.Printer.PrintStructured(tomlConf.Profiles[profileName], printer.StructuredOptions{})
	} else {
		// Capture whether TLS is configured before the loop below. Looking up
		// any "tls.*" property via reflectEnvConfigProp lazily initializes
		// confProfile.TLS to a non-nil empty struct, which would otherwise make
		// TLS appear configured when it is not (#1077).
		tlsConfigured := confProfile.TLS != nil

		// Get every property individually as a property-value pair except zero
		// vals
		var props []prop
		for k := range envConfigPropsToFieldNames {
			// TLS is a special case
			if k == "tls" {
				if tlsConfigured {
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

		// Add "grpc_meta"
		for k, v := range confProfile.GRPCMeta {
			props = append(props, prop{Property: "grpc_meta." + k, Value: v})
		}

		// Sort and display
		sort.Slice(props, func(i, j int) bool { return props[i].Property < props[j].Property })
		return cctx.Printer.PrintStructured(props, printer.StructuredOptions{Table: &printer.TableOptions{}})
	}
}

func (c *TemporalConfigListCommand) run(cctx *CommandContext, _ []string) error {
	clientConfig, err := envconfig.LoadClientConfig(envconfig.LoadClientConfigOptions{
		ConfigFilePath: cctx.RootCommand.ConfigFile,
		EnvLookup:      cctx.Options.EnvLookup,
	})
	if err != nil {
		return err
	}
	type profile struct {
		Name string `json:"name"`
	}
	profiles := make([]profile, 0, len(clientConfig.Profiles))
	for k := range clientConfig.Profiles {
		profiles = append(profiles, profile{Name: k})
	}
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].Name < profiles[j].Name })
	return cctx.Printer.PrintStructured(profiles, printer.StructuredOptions{Table: &printer.TableOptions{}})
}

func (c *TemporalConfigSetCommand) run(cctx *CommandContext, _ []string) error {
	if c.Prop == updateCheckConfigProp {
		if c.Value != "true" && c.Value != "false" {
			return fmt.Errorf("must be 'true' or 'false' to set this property")
		}
		configFile, _ := resolveConfigFile(cctx.Options)
		return setUpdateCheckEnabled(configFile, c.Value == "true")
	}
	// Load config
	conf, confProfile, err := loadEnvConfigProfile(cctx, envConfigProfileName(cctx), false)
	if err != nil {
		return err
	}
	// As a special case, "grpc_meta." values are handled specifically
	if strings.HasPrefix(c.Prop, "grpc_meta.") {
		if confProfile.GRPCMeta == nil {
			confProfile.GRPCMeta = map[string]string{}
		}
		confProfile.GRPCMeta[strings.TrimPrefix(c.Prop, "grpc_meta.")] = c.Value
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
			if reflectVal.Type().Elem().Kind() != reflect.Uint8 {
				return fmt.Errorf("unexpected slice of type %v", reflectVal.Type())
			}
			reflectVal.SetBytes([]byte(c.Value))
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
) (*envconfig.ClientConfig, *envconfig.ClientConfigProfile, error) {
	clientConfig, err := envconfig.LoadClientConfig(envconfig.LoadClientConfigOptions{
		ConfigFilePath: cctx.RootCommand.ConfigFile,
		EnvLookup:      cctx.Options.EnvLookup,
	})
	if err != nil {
		return nil, nil, err
	}

	// Load profile
	clientProfile := clientConfig.Profiles[profile]
	if clientProfile == nil {
		if failIfNotFound {
			return nil, nil, fmt.Errorf("profile %q not found", profile)
		}
		clientProfile = &envconfig.ClientConfigProfile{}
		clientConfig.Profiles[profile] = clientProfile
	}
	return &clientConfig, clientProfile, nil
}

var envConfigPropsToFieldNames = map[string]string{
	"address":                       "Address",
	"namespace":                     "Namespace",
	"api_key":                       "APIKey",
	"authority":                     "Authority",
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
}

func reflectEnvConfigProp(
	prof *envconfig.ClientConfigProfile,
	prop string,
	failIfParentNotFound bool,
) (reflect.Value, error) {
	// Get field name
	field := envConfigPropsToFieldNames[prop]
	if field == "" {
		return reflect.Value{}, fmt.Errorf("unknown property %q", prop)
	}

	// Load reflect val
	parentVal := reflect.ValueOf(prof)
	if strings.HasPrefix(prop, "tls.") {
		if prof.TLS == nil {
			if failIfParentNotFound {
				return reflect.Value{}, fmt.Errorf("no TLS options found")
			}
			prof.TLS = &envconfig.ClientConfigTLS{}
		}
		parentVal = reflect.ValueOf(prof.TLS)
	} else if strings.HasPrefix(prop, "codec.") {
		if prof.Codec == nil {
			if failIfParentNotFound {
				return reflect.Value{}, fmt.Errorf("no codec options found")
			}
			prof.Codec = &envconfig.ClientConfigCodec{}
		}
		parentVal = reflect.ValueOf(prof.Codec)
	}

	// Return reflected field
	if parentVal.Kind() == reflect.Pointer {
		parentVal = parentVal.Elem()
	}
	return parentVal.FieldByName(field), nil
}

func writeEnvConfigFile(cctx *CommandContext, conf *envconfig.ClientConfig) error {
	// Get file
	configFile := cctx.RootCommand.ConfigFile
	if configFile == "" {
		configFile, _ = cctx.Options.EnvLookup.LookupEnv("TEMPORAL_CONFIG_FILE")
		if configFile == "" {
			configFile = envconfig.DefaultConfigFilePath()

		}
	}

	// Convert to TOML
	b, err := conf.ToTOML(envconfig.ClientConfigToTOMLOptions{})
	if err != nil {
		return fmt.Errorf("failed building TOML: %w", err)
	}
	// The SDK config type only knows about connection profiles. Preserve CLI-owned
	// top-level configuration when rewriting those profiles.
	if existing, readErr := os.ReadFile(configFile); readErr == nil {
		var existingRaw, generatedRaw map[string]any
		if _, decodeErr := toml.Decode(string(existing), &existingRaw); decodeErr != nil {
			return fmt.Errorf("failed parsing existing config: %w", decodeErr)
		}
		if _, decodeErr := toml.Decode(string(b), &generatedRaw); decodeErr != nil {
			return fmt.Errorf("failed parsing generated config: %w", decodeErr)
		}
		if cliConfig, ok := existingRaw["cli"]; ok {
			generatedRaw["cli"] = cliConfig
		}
		var buf bytes.Buffer
		if encodeErr := toml.NewEncoder(&buf).Encode(generatedRaw); encodeErr != nil {
			return fmt.Errorf("failed preserving CLI config: %w", encodeErr)
		}
		b = buf.Bytes()
	} else if !errors.Is(readErr, os.ErrNotExist) {
		return fmt.Errorf("failed reading existing config: %w", readErr)
	}

	// Write to file, making dirs as needed
	if err := os.MkdirAll(filepath.Dir(configFile), 0700); err != nil {
		return fmt.Errorf("failed making config file parent dirs: %w", err)
	} else if err := os.WriteFile(configFile, b, 0600); err != nil {
		return fmt.Errorf("failed writing config file: %w", err)
	}
	return nil
}
