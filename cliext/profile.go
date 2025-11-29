package cliext

import (
	"fmt"
	"reflect"
	"strings"

	"go.temporal.io/sdk/contrib/envconfig"
)

// LoadProfileOptions contains options for loading a profile.
type LoadProfileOptions struct {
	// ConfigFilePath is the path to the configuration file.
	// If empty, the default path will be used.
	ConfigFilePath string

	// ProfileName is the name of the profile to load.
	ProfileName string

	// CreateIfMissing creates an empty profile if it doesn't exist.
	CreateIfMissing bool
}

var envProfilePropsToFieldNames = map[string]string{
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

// LoadProfile loads a specific profile from the configuration.
func LoadProfile(opts LoadProfileOptions) (*envconfig.ClientConfig, *envconfig.ClientConfigProfile, error) {
	config, err := LoadConfig(opts.ConfigFilePath)
	if err != nil {
		return nil, nil, err
	}

	profile := config.Profiles[opts.ProfileName]
	if profile == nil {
		if !opts.CreateIfMissing {
			return nil, nil, fmt.Errorf("profile %q not found", opts.ProfileName)
		}
		profile = &envconfig.ClientConfigProfile{}
		config.Profiles[opts.ProfileName] = profile
	}
	return config, profile, nil
}

// GetPropertyValue gets a property value from a profile by property name.
// For pointer types (like TLS), returns whether the pointer is non-nil.
func GetPropertyValue(profile *envconfig.ClientConfigProfile, prop string) (any, error) {
	// gRPC meta is special
	if strings.HasPrefix(prop, "grpc_meta.") {
		key := strings.TrimPrefix(prop, "grpc_meta.")
		v, ok := profile.GRPCMeta[key]
		if !ok {
			return nil, fmt.Errorf("unknown property %q", prop)
		}
		return v, nil
	}

	// Single value goes into property-value structure
	reflectVal, err := getReflectValue(profile, prop, false)
	if err != nil {
		return nil, err
	}

	// Pointers become true/false
	if reflectVal.Kind() == reflect.Pointer {
		return !reflectVal.IsNil(), nil
	}
	return reflectVal.Interface(), nil
}

// SetPropertyValue sets a property value on a profile by property name.
func SetPropertyValue(profile *envconfig.ClientConfigProfile, prop, value string) error {
	// As a special case, "grpc_meta." values are handled specifically
	if strings.HasPrefix(prop, "grpc_meta.") {
		if profile.GRPCMeta == nil {
			profile.GRPCMeta = map[string]string{}
		}
		profile.GRPCMeta[strings.TrimPrefix(prop, "grpc_meta.")] = value
		return nil
	}

	// Get reflect value
	reflectVal, err := getReflectValue(profile, prop, false)
	if err != nil {
		return err
	}

	// Set it from string
	switch reflectVal.Kind() {
	case reflect.String:
		reflectVal.SetString(value)
	case reflect.Pointer:
		// Used for "tls", true makes an empty object, false sets nil
		switch value {
		case "true":
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
		reflectVal.SetBytes([]byte(value))
	case reflect.Bool:
		if value != "true" && value != "false" {
			return fmt.Errorf("must be 'true' or 'false' to set this property")
		}
		reflectVal.SetBool(value == "true")
	case reflect.Map:
		return fmt.Errorf("must set each individual value of a map")
	default:
		return fmt.Errorf("unexpected type %v", reflectVal.Type())
	}
	return nil
}

// DeleteProperty deletes a property value from a profile.
func DeleteProperty(profile *envconfig.ClientConfigProfile, prop string) error {
	if strings.HasPrefix(prop, "grpc_meta.") {
		key := strings.TrimPrefix(prop, "grpc_meta.")
		if _, ok := profile.GRPCMeta[key]; !ok {
			return fmt.Errorf("gRPC meta key %q not found", key)
		}
		delete(profile.GRPCMeta, key)
		return nil
	}

	reflectVal, err := getReflectValue(profile, prop, true)
	if err != nil {
		return err
	}
	reflectVal.SetZero()
	return nil
}

// ListProperties returns all non-zero properties from a profile.
func ListProperties(profile *envconfig.ClientConfigProfile) (map[string]any, error) {
	// Get every property individually as a property-value pair except zero vals
	props := make(map[string]any)

	for k := range envProfilePropsToFieldNames {
		// TLS is a special case
		if k == "tls" {
			if profile.TLS != nil {
				props["tls"] = true
			}
			continue
		}
		val, err := getReflectValue(profile, k, false)
		if err != nil {
			return nil, err
		}
		if !val.IsZero() {
			props[k] = val.Interface()
		}
	}

	// Add "grpc_meta"
	for k, v := range profile.GRPCMeta {
		props["grpc_meta."+k] = v
	}

	return props, nil
}

func getReflectValue(
	prof *envconfig.ClientConfigProfile,
	prop string,
	failIfParentNotFound bool,
) (reflect.Value, error) {
	// Get field name
	field := envProfilePropsToFieldNames[prop]
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
