package tests

import (
	"os"
	"path/filepath"

	"github.com/temporalio/tctl-kit/pkg/config"
	"github.com/urfave/cli/v2"
)

const (
	testEnvName = "tctl-test-env"
)

func (s *e2eSuite) TestSetEnvValue() {
	testserver, app, _ := s.setUpTestEnvironment()
	defer func() {
		_ = testserver.Stop()
	}()

	cleanup := setupConfig(s, app)
	defer cleanup()

	err := app.Run([]string{"", "env", "set", testEnvName + ".address", "0.0.0.0:00000"})
	s.NoError(err)

	cfg := readConfig(s)
	s.Contains(cfg, "tctl-test-env:")
	s.Contains(cfg, "address: 0.0.0.0:00000")
	s.Contains(cfg, "namespace: tctl-test-namespace")
}

func (s *e2eSuite) TestDeleteEnvProperty() {
	testserver, app, _ := s.setUpTestEnvironment()
	defer func() {
		_ = testserver.Stop()
	}()

	cleanup := setupConfig(s, app)
	defer cleanup()

	err := app.Run([]string{"", "env", "set", testEnvName + ".address", "1.2.3.4:5678"})
	s.NoError(err)

	err = app.Run([]string{"", "env", "delete", testEnvName + ".address"})
	s.NoError(err)

	cfg := readConfig(s)
	s.Contains(cfg, "tctl-test-env:")
	s.Contains(cfg, "namespace: tctl-test-namespace")
	s.NotContains(cfg, "address: 1.2.3.4:5678")
}

func (s *e2eSuite) TestDeleteEnv() {
	testserver, app, _ := s.setUpTestEnvironment()
	defer func() {
		_ = testserver.Stop()
	}()

	cleanup := setupConfig(s, app)
	defer cleanup()

	err := app.Run([]string{"", "env", "set", testEnvName + ".address", "1.2.3.4:5678"})
	s.NoError(err)

	err = app.Run([]string{"", "env", "delete", testEnvName})
	s.NoError(err)

	cfg := readConfig(s)
	s.NotContains(cfg, "tctl-test-env:")
	s.NotContains(cfg, "namespace: tctl-test-namespace")
	s.NotContains(cfg, "address: 1.2.3.4:5678")
}

func (s *e2eSuite) TestDeleteEnv_Default() {
	testserver, app, _ := s.setUpTestEnvironment()
	defer func() {
		_ = testserver.Stop()
	}()

	cleanup := setupConfig(s, app)
	defer cleanup()

	err := app.Run([]string{"", "env", "set", testEnvName + ".address", "1.2.3.4:5678"})
	s.NoError(err)

	err = app.Run([]string{"", "env", "delete", config.DefaultEnv})
	s.NoError(err)

	cfg := readConfig(s)
	s.NotContains(cfg, "default:")

	err = app.Run([]string{"", "workflow", "list"})
	s.NoError(err)
}

func setupConfig(s *e2eSuite, app *cli.App) func() {
	err := app.Run([]string{"", "env", "set", testEnvName + ".namespace", "tctl-test-namespace"})
	s.NoError(err)

	return func() {
		err := app.Run([]string{"", "env", "delete", testEnvName})
		s.NoError(err, "unable to unset test env")
	}
}

func readConfig(s *e2eSuite) string {
	path := getConfigPath(s)
	content, err := os.ReadFile(path)
	s.NoError(err)

	return string(content)
}

func getConfigPath(s *e2eSuite) string {
	dpath, err := os.UserHomeDir()
	s.NoError(err)

	path := filepath.Join(dpath, ".config", "temporalio", "temporal.yaml")

	return path
}
