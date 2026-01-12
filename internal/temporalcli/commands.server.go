package temporalcli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.temporal.io/api/enums/v1"

	"github.com/temporalio/cli/internal/devserver"
)

var defaultDynamicConfigValues = map[string]any{
	// Make search attributes immediately visible on creation, so users don't
	// have to wait for eventual consistency to happen when testing against the
	// dev-server.  Since it's a very rare thing to create search attributes,
	// we're comfortable that this is very unlikely to mask bugs in user code.
	"system.forceSearchAttributesCacheRefreshOnRead": true,

	// Since we disable the SA cache, we need to bump max QPS accordingly.
	// These numbers were chosen to maintain the ratio between the two that's
	// established in the defaults.
	"frontend.persistenceMaxQPS": 10000,
	"history.persistenceMaxQPS":  45000,
}

func (t *TemporalServerStartDevCommand) run(cctx *CommandContext, args []string) error {
	// Have to assume "localhost" is 127.0.0.1 for server to work (it expects IP)
	if t.Ip == "localhost" {
		t.Ip = "127.0.0.1"
	}
	// Prepare options
	opts := devserver.StartOptions{
		FrontendIP:             t.Ip,
		FrontendPort:           t.Port,
		Namespaces:             append([]string{"default"}, t.Namespace...),
		Logger:                 cctx.Logger,
		DatabaseFile:           t.DbFilename,
		MetricsPort:            t.MetricsPort,
		FrontendHTTPPort:       t.HttpPort,
		ClusterID:              uuid.NewString(),
		MasterClusterName:      "active",
		CurrentClusterName:     "active",
		InitialFailoverVersion: 1,
	}
	// Set the log level value of the server to the overall log level given to the
	// CLI. But if it is "never" we have to do a special value, and if it was
	// never changed, we have to use the default of "warn" instead of the CLI
	// default of "info" since server is noisier.
	logLevel := t.Parent.Parent.LogLevel.Value
	if !t.Parent.Parent.LogLevel.ChangedFromDefault {
		logLevel = "warn"
	}
	if logLevel == "never" {
		opts.LogLevel = 100
	} else if err := opts.LogLevel.UnmarshalText([]byte(logLevel)); err != nil {
		return fmt.Errorf("invalid log level %q: %w", logLevel, err)
	}
	if err := devserver.CheckPortFree(opts.FrontendIP, opts.FrontendPort); err != nil {
		return fmt.Errorf("can't set frontend port %d: %w", opts.FrontendPort, err)
	}

	if opts.FrontendHTTPPort > 0 {
		if err := devserver.CheckPortFree(opts.FrontendIP, opts.FrontendHTTPPort); err != nil {
			return fmt.Errorf("can't set frontend HTTP port %d: %w", opts.FrontendHTTPPort, err)
		}
	}
	// Setup UI
	if !t.Headless {
		opts.UIIP, opts.UIPort = t.UiIp, t.UiPort
		if opts.UIIP == "" {
			opts.UIIP = t.Ip
		}
		if opts.UIPort == 0 {
			opts.UIPort = t.Port + 1000
			if opts.UIPort > 65535 {
				opts.UIPort = 65535
			}
			if err := devserver.CheckPortFree(opts.UIIP, opts.UIPort); err != nil {
				return fmt.Errorf("can't use default UI port %d (%d + 1000): %w", opts.UIPort, t.Port, err)
			}
		} else {
			if err := devserver.CheckPortFree(opts.UIIP, opts.UIPort); err != nil {
				return fmt.Errorf("can't set UI port %d: %w", opts.UIPort, err)
			}
		}
		opts.UIAssetPath, opts.UICodecEndpoint, opts.PublicPath = t.UiAssetPath, t.UiCodecEndpoint, t.UiPublicPath
	}
	// Pragmas and dyn config
	var err error
	if opts.SqlitePragmas, err = stringKeysValues(t.SqlitePragma); err != nil {
		return fmt.Errorf("invalid pragma: %w", err)
	} else if opts.DynamicConfigValues, err = stringKeysJSONValues(t.DynamicConfigValue, true); err != nil {
		return fmt.Errorf("invalid dynamic config values: %w", err)
	}
	// We have to convert all dynamic config values that JSON number to int if we
	// can because server dynamic config expecting int won't work with the default
	// float JSON unmarshal uses
	for k, v := range opts.DynamicConfigValues {
		if num, ok := v.(json.Number); ok {
			if newV, err := num.Int64(); err == nil {
				// Dynamic config only accepts int type, not int32 nor int64
				opts.DynamicConfigValues[k] = int(newV)
			} else if newV, err := num.Float64(); err == nil {
				opts.DynamicConfigValues[k] = newV
			} else {
				return fmt.Errorf("invalid JSON value for key %q", k)
			}
		}
	}

	// Apply set of default dynamic config values if not already present
	for k, v := range defaultDynamicConfigValues {
		if _, ok := opts.DynamicConfigValues[k]; !ok {
			if opts.DynamicConfigValues == nil {
				opts.DynamicConfigValues = map[string]any{}
			}
			opts.DynamicConfigValues[k] = v
		}
	}

	// Prepare search attributes for adding before starting server
	searchAttrs, err := t.prepareSearchAttributes()
	if err != nil {
		return err
	}
	opts.SearchAttributes = searchAttrs

	// If not using DB file, set persistent cluster ID
	if t.DbFilename == "" {
		opts.ClusterID = persistentClusterID()
	}
	// Log config if requested
	if t.LogConfig {
		opts.LogConfig = func(b []byte) {
			cctx.Logger.Info("Logging config")
			_, _ = cctx.Options.Stderr.Write(b)
		}
	}
	// Grab a free port for metrics ahead-of-time so we know what port is selected
	if opts.MetricsPort == 0 {
		opts.MetricsPort = devserver.MustGetFreePort(opts.FrontendIP)
	} else {
		if err := devserver.CheckPortFree(opts.FrontendIP, opts.MetricsPort); err != nil {
			return fmt.Errorf("can't set metrics port %d: %w", opts.MetricsPort, err)
		}
	}

	// Start, wait for context complete, then stop
	s, err := devserver.Start(opts)
	if err != nil {
		return fmt.Errorf("failed starting server: %w", err)
	}
	defer s.Stop()

	cctx.Printer.Printlnf("CLI %v\n", VersionString())
	cctx.Printer.Printlnf("%-8s %v:%v", "Server:", toFriendlyIp(opts.FrontendIP), opts.FrontendPort)
	// Only print HTTP port if explicitly provided to avoid promoting the unstable HTTP API.
	if opts.FrontendHTTPPort > 0 {
		cctx.Printer.Printlnf("%-8s %v:%v", "HTTP:", toFriendlyIp(opts.FrontendIP), opts.FrontendHTTPPort)
	}
	if !t.Headless {
		cctx.Printer.Printlnf("%-8s http://%v:%v%v", "UI:", toFriendlyIp(opts.UIIP), opts.UIPort, opts.PublicPath)
	}
	cctx.Printer.Printlnf("%-8s http://%v:%v/metrics", "Metrics:", toFriendlyIp(opts.FrontendIP), opts.MetricsPort)
	<-cctx.Done()
	s.SuppressWarnings()
	return nil
}

func toFriendlyIp(host string) string {
	if host == "127.0.0.1" || host == "::1" {
		return "localhost"
	}
	return devserver.MaybeEscapeIPv6(host)
}

func persistentClusterID() string {
	// If there is not a database file in use, we want a cluster ID to be the same
	// for every re-run, so we set it as an environment config in a special env
	// file. We do not error if we can neither read nor write the file.
	file := defaultDeprecatedEnvConfigFile("temporalio", "version-info")
	if file == "" {
		// No file, can do nothing here
		return uuid.NewString()
	}
	// Try to get existing first
	env, _ := readDeprecatedEnvConfigFile(file)
	if id := env["default"]["cluster-id"]; id != "" {
		return id
	}
	// Create and try to write
	id := uuid.NewString()
	_ = writeDeprecatedEnvConfigFile(file, map[string]map[string]string{"default": {"cluster-id": id}})
	return id
}

func (t *TemporalServerStartDevCommand) prepareSearchAttributes() (map[string]enums.IndexedValueType, error) {
	opts, err := stringKeysValues(t.SearchAttribute)
	if err != nil {
		return nil, fmt.Errorf("invalid search attributes: %w", err)
	}
	attrs := make(map[string]enums.IndexedValueType, len(opts))
	for k, v := range opts {
		// Case-insensitive index type lookup
		var valType enums.IndexedValueType
		for valTypeName, valTypeOrd := range enums.IndexedValueType_shorthandValue {
			if strings.EqualFold(v, valTypeName) {
				valType = enums.IndexedValueType(valTypeOrd)
				break
			}
		}
		if valType == 0 {
			return nil, fmt.Errorf("invalid search attribute value type %q", v)
		}
		attrs[k] = valType
	}
	return attrs, nil
}
