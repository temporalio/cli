package temporalcli

import (
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
	// Load config
	profileName := envConfigProfileName(cctx)
	conf, confProfile, additionalProfileFields, err := loadEnvConfigProfile(cctx, profileName, true)
	if err != nil {
		return err
	}
	propName := normalizeEnvConfigProp(c.Prop)
	if strings.HasPrefix(c.Prop, "grpc_meta.") {
		key := strings.TrimPrefix(c.Prop, "grpc_meta.")
		if _, ok := confProfile.GRPCMeta[key]; !ok {
			return fmt.Errorf("gRPC meta key %q not found", key)
		}
		delete(confProfile.GRPCMeta, key)
	} else if propName == envConfigPropClientAuthority {
		deleteClientAuthority(additionalProfileFields, profileName)
	} else {
		reflectVal, err := reflectEnvConfigProp(confProfile, propName, true)
		if err != nil {
			return err
		}
		reflectVal.SetZero()
	}

	// Save
	return writeEnvConfigFile(cctx, conf, additionalProfileFields)
}

func (c *TemporalConfigDeleteProfileCommand) run(cctx *CommandContext, _ []string) error {
	// Load config
	profileName := envConfigProfileName(cctx)
	conf, _, additionalProfileFields, err := loadEnvConfigProfile(cctx, profileName, true)
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
	delete(additionalProfileFields, profileName)

	// Save
	return writeEnvConfigFile(cctx, conf, additionalProfileFields)
}

func (c *TemporalConfigGetCommand) run(cctx *CommandContext, _ []string) error {
	// Load config profile
	profileName := envConfigProfileName(cctx)
	conf, confProfile, additionalProfileFields, err := loadEnvConfigProfile(cctx, profileName, true)
	if err != nil {
		return err
	}
	type prop struct {
		Property string `json:"property"`
		Value    any    `json:"value"`
	}
	// If there is a specific key requested, show it, otherwise show all
	if c.Prop != "" {
		propName := normalizeEnvConfigProp(c.Prop)
		// We do not support asking for structures with children at this time,
		// but "tls" is a special case because it's also a bool.
		if propName == "codec" || propName == "grpc_meta" {
			return fmt.Errorf("must provide exact property, not parent property")
		}
		// gRPC meta is special
		if strings.HasPrefix(c.Prop, "grpc_meta.") {
			v, ok := confProfile.GRPCMeta[strings.TrimPrefix(c.Prop, "grpc_meta.")]
			if !ok {
				return fmt.Errorf("unknown property %q", c.Prop)
			}
			return cctx.Printer.PrintStructured(
				prop{Property: c.Prop, Value: v},
				printer.StructuredOptions{Table: &printer.TableOptions{}},
			)
		} else if propName == envConfigPropClientAuthority {
			v, _, err := clientAuthorityFromAdditionalProfileFields(additionalProfileFields, profileName)
			if err != nil {
				return err
			}
			return cctx.Printer.PrintStructured(
				prop{Property: envConfigPropClientAuthority, Value: v},
				printer.StructuredOptions{Table: &printer.TableOptions{}},
			)
		} else {
			// Single value goes into property-value structure
			reflectVal, err := reflectEnvConfigProp(confProfile, propName, false)
			if err != nil {
				return err
			}
			// Pointers become true/false
			if reflectVal.Kind() == reflect.Pointer {
				reflectVal = reflect.ValueOf(!reflectVal.IsNil())
			}
			return cctx.Printer.PrintStructured(
				prop{Property: propName, Value: reflectVal.Interface()},
				printer.StructuredOptions{Table: &printer.TableOptions{}},
			)
		}
	} else if cctx.JSONOutput {
		// If it is JSON and not prop specific, we want to dump the TOML
		// structure in JSON form
		var tomlConf struct {
			Profiles map[string]any `toml:"profile"`
		}
		additionalFields, err := knownAdditionalProfileFields(additionalProfileFields, profileName)
		if err != nil {
			return err
		}
		if b, err := conf.ToTOML(envconfig.ClientConfigToTOMLOptions{AdditionalProfileFields: additionalFields}); err != nil {
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
		if v, ok, err := clientAuthorityFromAdditionalProfileFields(additionalProfileFields, profileName); err != nil {
			return err
		} else if ok {
			props = append(props, prop{Property: envConfigPropClientAuthority, Value: v})
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
	// Load config
	profileName := envConfigProfileName(cctx)
	conf, confProfile, additionalProfileFields, err := loadEnvConfigProfile(cctx, profileName, false)
	if err != nil {
		return err
	}
	propName := normalizeEnvConfigProp(c.Prop)
	// As a special case, "grpc_meta." values are handled specifically
	if strings.HasPrefix(c.Prop, "grpc_meta.") {
		if confProfile.GRPCMeta == nil {
			confProfile.GRPCMeta = map[string]string{}
		}
		confProfile.GRPCMeta[strings.TrimPrefix(c.Prop, "grpc_meta.")] = c.Value
	} else if propName == envConfigPropClientAuthority {
		setClientAuthority(additionalProfileFields, profileName, c.Value)
	} else {
		// Get reflect value
		reflectVal, err := reflectEnvConfigProp(confProfile, propName, false)
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
	return writeEnvConfigFile(cctx, conf, additionalProfileFields)
}

func envConfigProfileName(cctx *CommandContext) string {
	if cctx.RootCommand.Profile != "" {
		return cctx.RootCommand.Profile
	} else if cctx.Options.EnvLookup != nil {
		if p, _ := cctx.Options.EnvLookup.LookupEnv("TEMPORAL_PROFILE"); p != "" {
			return p
		}
	}
	return envconfig.DefaultConfigFileProfile
}

func loadEnvConfigProfile(
	cctx *CommandContext,
	profile string,
	failIfNotFound bool,
) (*envconfig.ClientConfig, *envconfig.ClientConfigProfile, map[string]map[string]any, error) {
	clientConfig, additionalProfileFields, err := loadEnvConfigFile(cctx)
	if err != nil {
		return nil, nil, nil, err
	}

	// Load profile
	clientProfile := clientConfig.Profiles[profile]
	if clientProfile == nil {
		if failIfNotFound {
			return nil, nil, nil, fmt.Errorf("profile %q not found", profile)
		}
		clientProfile = &envconfig.ClientConfigProfile{}
		clientConfig.Profiles[profile] = clientProfile
	}
	return clientConfig, clientProfile, additionalProfileFields, nil
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
}

const (
	envConfigPropClientAuthority      = "client_authority"
	envConfigPropClientAuthorityAlias = "client-authority"
)

func normalizeEnvConfigProp(prop string) string {
	if prop == envConfigPropClientAuthorityAlias {
		return envConfigPropClientAuthority
	}
	return prop
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

func loadEnvConfigFile(cctx *CommandContext) (*envconfig.ClientConfig, map[string]map[string]any, error) {
	configFile, err := envConfigFilePath(cctx)
	if err != nil {
		return nil, nil, err
	}

	var data []byte
	if b, err := os.ReadFile(configFile); err == nil {
		data = b
	} else if !os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("failed reading file at %v: %w", configFile, err)
	}

	clientConfig := &envconfig.ClientConfig{}
	additionalProfileFields := map[string]map[string]any{}
	if err := clientConfig.FromTOML(data, envconfig.ClientConfigFromTOMLOptions{
		AdditionalProfileFields: additionalProfileFields,
	}); err != nil {
		return nil, nil, fmt.Errorf("failed parsing config: %w", err)
	}
	if clientConfig.Profiles == nil {
		clientConfig.Profiles = map[string]*envconfig.ClientConfigProfile{}
	}
	return clientConfig, additionalProfileFields, nil
}

func envConfigFilePath(cctx *CommandContext) (string, error) {
	configFile := cctx.RootCommand.ConfigFile
	if configFile == "" {
		env := cctx.Options.EnvLookup
		if env == nil {
			env = envconfig.EnvLookupOS
		}
		configFile, _ = env.LookupEnv("TEMPORAL_CONFIG_FILE")
		if configFile == "" {
			var err error
			if configFile, err = envconfig.DefaultConfigFilePath(); err != nil {
				return "", err
			}
		}
	}
	return configFile, nil
}

func writeEnvConfigFile(
	cctx *CommandContext,
	conf *envconfig.ClientConfig,
	additionalProfileFields map[string]map[string]any,
) error {
	configFile, err := envConfigFilePath(cctx)
	if err != nil {
		return err
	}
	if err := normalizeAdditionalProfileFields(additionalProfileFields); err != nil {
		return err
	}

	// Convert to TOML
	b, err := conf.ToTOML(envconfig.ClientConfigToTOMLOptions{AdditionalProfileFields: additionalProfileFields})
	if err != nil {
		return fmt.Errorf("failed building TOML: %w", err)
	}

	// Write to file, making dirs as needed
	if err := os.MkdirAll(filepath.Dir(configFile), 0700); err != nil {
		return fmt.Errorf("failed making config file parent dirs: %w", err)
	} else if err := os.WriteFile(configFile, b, 0600); err != nil {
		return fmt.Errorf("failed writing config file: %w", err)
	}
	return nil
}

func clientAuthorityFromAdditionalProfileFields(
	additionalProfileFields map[string]map[string]any,
	profileName string,
) (string, bool, error) {
	profileFields := additionalProfileFields[profileName]
	if profileFields == nil {
		return "", false, nil
	}
	if raw, ok := profileFields[envConfigPropClientAuthority]; ok {
		v, ok := raw.(string)
		if !ok {
			return "", false, fmt.Errorf("property %q must be a string", envConfigPropClientAuthority)
		}
		return v, true, nil
	}
	if raw, ok := profileFields[envConfigPropClientAuthorityAlias]; ok {
		v, ok := raw.(string)
		if !ok {
			return "", false, fmt.Errorf("property %q must be a string", envConfigPropClientAuthority)
		}
		return v, true, nil
	}
	return "", false, nil
}

func setClientAuthority(additionalProfileFields map[string]map[string]any, profileName, value string) {
	if additionalProfileFields[profileName] == nil {
		additionalProfileFields[profileName] = map[string]any{}
	}
	additionalProfileFields[profileName][envConfigPropClientAuthority] = value
	delete(additionalProfileFields[profileName], envConfigPropClientAuthorityAlias)
}

func deleteClientAuthority(additionalProfileFields map[string]map[string]any, profileName string) {
	if additionalProfileFields[profileName] == nil {
		return
	}
	delete(additionalProfileFields[profileName], envConfigPropClientAuthority)
	delete(additionalProfileFields[profileName], envConfigPropClientAuthorityAlias)
}

func normalizeAdditionalProfileFields(additionalProfileFields map[string]map[string]any) error {
	for profileName, profileFields := range additionalProfileFields {
		value, ok, err := clientAuthorityFromAdditionalProfileFields(additionalProfileFields, profileName)
		if err != nil {
			return err
		}
		if ok {
			profileFields[envConfigPropClientAuthority] = value
		}
		delete(profileFields, envConfigPropClientAuthorityAlias)
	}
	return nil
}

func knownAdditionalProfileFields(
	additionalProfileFields map[string]map[string]any,
	profileName string,
) (map[string]map[string]any, error) {
	value, ok, err := clientAuthorityFromAdditionalProfileFields(additionalProfileFields, profileName)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return map[string]map[string]any{
		profileName: {
			envConfigPropClientAuthority: value,
		},
	}, nil
}
