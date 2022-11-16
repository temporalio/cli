// The MIT License
//
// Copyright (c) 2022 Temporal Technologies Inc.  All rights reserved.
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package server

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/temporalio/temporal-cli/common"
	sconfig "github.com/temporalio/temporal-cli/server/config"
	"github.com/urfave/cli/v2"
	"go.temporal.io/server/common/config"
	"go.temporal.io/server/common/dynamicconfig"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/temporal"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewServerCommands(defaultCfg *sconfig.Config) []*cli.Command {
	return []*cli.Command{
		{
			Name:      "start-dev",
			Usage:     "Start Temporal development server",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    common.FlagDBPath,
					Aliases: []string{"f"},
					Value:   defaultCfg.DatabaseFilePath,
					Usage:   "File in which to persist Temporal state",
				},
				&cli.StringSliceFlag{
					Name:    common.FlagNamespace,
					Aliases: common.FlagNamespaceAlias,
					Usage:   "Specify namespaces that should be pre-created",
					EnvVars: nil,
					Value:   nil,
				},
				&cli.IntFlag{
					Name:    common.FlagPort,
					Aliases: []string{"p"},
					Usage:   "Port for the temporal-frontend GRPC service",
					Value:   sconfig.DefaultFrontendPort,
				},
				&cli.IntFlag{
					Name:  common.FlagMetricsPort,
					Usage: "Port for the metrics listener",
					Value: sconfig.DefaultMetricsPort,
				},
				&cli.IntFlag{
					Name:        common.FlagUIPort,
					Usage:       "port for the temporal web UI",
					DefaultText: fmt.Sprintf("--port + 1000, eg. %d", sconfig.DefaultFrontendPort+1000),
				},
				&cli.BoolFlag{
					Name:  common.FlagHeadless,
					Usage: "disable the temporal web UI",
				},
				&cli.StringFlag{
					Name:    common.FlagIP,
					Usage:   `IPv4 address to bind the frontend service to instead of localhost`,
					EnvVars: nil,
					Value:   "127.0.0.1",
				},
				&cli.StringFlag{
					Name:        common.FlagUIIP,
					Usage:       `IPv4 address to bind the web UI to instead of localhost`,
					DefaultText: "same as --ip (eg. 127.0.0.1)",
				},
				&cli.StringFlag{
					Name:    common.FlagLogFormat,
					Usage:   `customize the log formatting (allowed: ["json" "pretty"])`,
					EnvVars: nil,
					Value:   "json",
				},
				&cli.StringFlag{
					Name:    common.FlagLogLevel,
					Usage:   `customize the log level (allowed: ["debug" "info" "warn" "error" "fatal"])`,
					EnvVars: nil,
					Value:   "info",
				},
				&cli.StringSliceFlag{
					Name:    common.FlagPragma,
					Usage:   fmt.Sprintf("specify sqlite pragma statements in pragma=value format (allowed: %q)", sconfig.GetAllowedPragmas()),
					EnvVars: nil,
					Value:   nil,
				},
				&cli.StringFlag{
					Name:    common.FlagConfig,
					Aliases: []string{"c"},
					Usage:   `config dir path`,
					EnvVars: []string{config.EnvKeyConfigDir},
					Value:   "",
				},
				&cli.StringSliceFlag{
					Name:  common.FlagDynamicConfigValue,
					Usage: `dynamic config value, as KEY=JSON_VALUE (meaning strings need quotes)`,
				},
			},
			Before: func(c *cli.Context) error {
				if c.Args().Len() > 0 {
					return cli.Exit("ERROR: start-dev command doesn't support arguments.", 1)
				}

				// Make sure the default db path exists (user does not specify path explicitly)
				if !c.IsSet(common.FlagDBPath) {
					if err := os.MkdirAll(filepath.Dir(c.String(common.FlagDBPath)), os.ModePerm); err != nil {
						return cli.Exit(err.Error(), 1)
					}
				}

				switch c.String(common.FlagLogFormat) {
				case "json", "pretty", "noop":
				default:
					return cli.Exit(fmt.Sprintf("bad value %q passed for flag %q", c.String(common.FlagLogFormat), common.FlagLogFormat), 1)
				}

				switch c.String(common.FlagLogLevel) {
				case "debug", "info", "warn", "error", "fatal":
				default:
					return cli.Exit(fmt.Sprintf("bad value %q passed for flag %q", c.String(common.FlagLogLevel), common.FlagLogLevel), 1)
				}

				// Check that ip address is valid
				if c.IsSet(common.FlagIP) && net.ParseIP(c.String(common.FlagIP)) == nil {
					return cli.Exit(fmt.Sprintf("bad value %q passed for flag %q", c.String(common.FlagIP), common.FlagIP), 1)
				}

				if c.IsSet(common.FlagConfig) {
					cfgPath := c.String(common.FlagConfig)
					if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
						return cli.Exit(fmt.Sprintf("bad value %q passed for flag %q: file not found", c.String(common.FlagConfig), common.FlagConfig), 1)
					}
				}

				return nil
			},
			Action: func(c *cli.Context) error {
				var (
					ip          = c.String(common.FlagIP)
					serverPort  = c.Int(common.FlagPort)
					metricsPort = c.Int(common.FlagMetricsPort)
					uiPort      = serverPort + 1000
					uiIP        = ip
				)

				if c.IsSet(common.FlagUIPort) {
					uiPort = c.Int(common.FlagUIPort)
				}

				if c.IsSet(common.FlagUIIP) {
					uiIP = c.String(common.FlagUIIP)
				}

				pragmas, err := getPragmaMap(c.StringSlice(common.FlagPragma))
				if err != nil {
					return err
				}

				baseConfig := &config.Config{}
				if c.IsSet(common.FlagConfig) {
					// Temporal server requires a couple of persistence config values to
					// be explicitly set or the config loading fails. While these are the
					// same values used internally, they are overridden later anyways,
					// they are just here to pass validation.
					baseConfig.Persistence.DefaultStore = sconfig.PersistenceStoreName
					baseConfig.Persistence.NumHistoryShards = 1
					if err := config.Load("temporal", c.String(common.FlagConfig), "", &baseConfig); err != nil {
						return err
					}
				}

				interruptChan := make(chan interface{}, 1)
				go func() {
					if doneChan := c.Done(); doneChan != nil {
						s := <-doneChan
						interruptChan <- s
					} else {
						s := <-temporal.InterruptCh()
						interruptChan <- s
					}
				}()

				opts := []ServerOption{
					WithDynamicPorts(),
					WithFrontendPort(serverPort),
					WithMetricsPort(metricsPort),
					WithFrontendIP(ip),
					WithDatabaseFilePath(c.String(common.FlagDBPath)),
					WithNamespaces(c.StringSlice(common.FlagNamespace)...),
					WithSQLitePragmas(pragmas),
					WithUpstreamOptions(
						temporal.InterruptOn(interruptChan),
					),
					WithBaseConfig(baseConfig),
				}
				if !c.Bool(common.FlagHeadless) {
					frontendAddr := fmt.Sprintf("%s:%d", ip, serverPort)
					opt, err := newUIOption(frontendAddr, uiIP, uiPort, c.String(common.FlagConfig))
					if err != nil {
						return err
					}
					if opt != nil {
						opts = append(opts, opt)
					}
				}
				if c.String(common.FlagDBPath) == "" {
					opts = append(opts, WithPersistenceDisabled())
				}

				var logger log.Logger
				switch c.String(common.FlagLogFormat) {
				case "pretty":
					lcfg := zap.NewDevelopmentConfig()
					switch c.String(common.FlagLogLevel) {
					case "debug":
						lcfg.Level.SetLevel(zap.DebugLevel)
					case "info":
						lcfg.Level.SetLevel(zap.InfoLevel)
					case "warn":
						lcfg.Level.SetLevel(zap.WarnLevel)
					case "error":
						lcfg.Level.SetLevel(zap.ErrorLevel)
					case "fatal":
						lcfg.Level.SetLevel(zap.FatalLevel)
					}
					lcfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
					l, err := lcfg.Build(
						zap.WithCaller(false),
						zap.AddStacktrace(zapcore.ErrorLevel),
					)
					if err != nil {
						return err
					}
					logger = log.NewZapLogger(l)
				case "noop":
					logger = log.NewNoopLogger()
				default:
					logger = log.NewZapLogger(log.BuildZapLogger(log.Config{
						Stdout:     true,
						Level:      c.String(common.FlagLogLevel),
						OutputFile: "",
					}))
				}
				opts = append(opts, WithLogger(logger))

				configVals, err := getDynamicConfigValues(c.StringSlice(common.FlagDynamicConfigValue))
				if err != nil {
					return err
				}

				if _, ok := configVals[dynamicconfig.ForceSearchAttributesCacheRefreshOnRead]; !ok {
					opts = append(opts, WithSearchAttributeCacheDisabled())
				}

				for k, v := range configVals {
					opts = append(opts, WithDynamicConfigValue(k, v))
				}

				s, err := NewServer(opts...)
				if err != nil {
					return err
				}

				if err := s.Start(); err != nil {
					return cli.Exit(fmt.Sprintf("Unable to start server. Error: %v", err), 1)
				}
				return cli.Exit("All services are stopped.", 0)
			},
		},
	}
}

func getPragmaMap(input []string) (map[string]string, error) {
	result := make(map[string]string)

	for _, pragma := range input {
		vals := strings.Split(pragma, "=")
		if len(vals) != 2 {
			return nil, fmt.Errorf("ERROR: pragma statements must be in KEY=VALUE format, got %q", pragma)
		}
		result[vals[0]] = vals[1]
	}
	return result, nil
}

func getDynamicConfigValues(input []string) (map[dynamicconfig.Key][]dynamicconfig.ConstrainedValue, error) {
	ret := make(map[dynamicconfig.Key][]dynamicconfig.ConstrainedValue, len(input))
	for _, keyValStr := range input {
		keyVal := strings.SplitN(keyValStr, "=", 2)
		if len(keyVal) != 2 {
			return nil, fmt.Errorf("dynamic config value not in KEY=JSON_VAL format")
		}
		key := dynamicconfig.Key(keyVal[0])
		// We don't support constraints currently
		var val dynamicconfig.ConstrainedValue
		if err := json.Unmarshal([]byte(keyVal[1]), &val.Value); err != nil {
			return nil, fmt.Errorf("invalid JSON value for key %q: %w", key, err)
		}
		ret[key] = append(ret[key], val)
	}
	return ret, nil
}
