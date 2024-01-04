package temporalcli

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/temporalio/cli/temporalcli/devserver"
)

func (t *TemporalServerStartDevCommand) run(cctx *CommandContext, args []string) error {
	// Prepare options
	opts := devserver.StartOptions{
		FrontendIP:       t.Ip,
		FrontendPort:     t.Port,
		Namespaces:       append([]string{"default"}, t.Namespace...),
		Logger:           cctx.Logger,
		DatabaseFile:     t.DbFilename,
		MetricsPort:      t.MetricsPort,
		FrontendHTTPPort: t.HttpPort,
	}
	// Setup UI
	if !t.Headless {
		opts.UIIP, opts.UIPort = t.Ip, t.UiPort
		if opts.UIIP == "" {
			opts.UIIP = t.Ip
		}
		if opts.UIPort == 0 {
			opts.UIPort = t.Port + 1000
		}
		opts.UIAssetPath, opts.UICodecEndpoint = t.UiAssetPath, t.UiCodecEndpoint
	}
	// Pragmas and dyn config
	var err error
	if opts.SqlitePragmas, err = stringKeysValues(t.SqlitePragma); err != nil {
		return fmt.Errorf("invalid pragma: %w", err)
	} else if opts.DynamicConfigValues, err = stringKeysJSONValues(t.DynamicConfigValue); err != nil {
		return fmt.Errorf("invalid dynamic config values: %w", err)
	}
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

	// Start, wait for context complete, then stop
	cctx.Logger.Info("Starting dev server...")
	s, err := devserver.Start(opts)
	if err != nil {
		return fmt.Errorf("failed starting server: %w", err)
	}
	defer s.Stop()
	<-cctx.Done()
	cctx.Logger.Info("Stopping dev server...")
	return nil
}

func persistentClusterID() string {
	// If there is not a database file in use, we want a cluster ID to be the same
	// for every re-run, so we set it as an environment config in a special env
	// file. We do not error if we can neither read nor write the file.
	file := defaultEnvConfigFile("temporalio", "version-info")
	if file == "" {
		// No file, can do nothing here
		return uuid.NewString()
	}
	// Try to get existing first
	env, _ := readEnvConfigFile(file)
	if id := env["default"]["cluster-id"]; id != "" {
		return id
	}
	// Create and try to write
	id := uuid.NewString()
	_ = writeEnvConfigFile(file, map[string]map[string]string{"default": {"cluster-id": id}})
	return id
}
