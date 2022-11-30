package app_test

import (
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

const (
	testEnvName = "tctl-test-env"
)

func (s *cliAppSuite) TestSetEnvValue() {
	defer setupConfig(s.app)()

	err := s.app.Run([]string{"", "env", "set", testEnvName + ".address", "0.0.0.0:00000"})
	s.NoError(err)

	config := readConfig()
	s.Contains(config, "tctl-test-env:")
	s.Contains(config, "address: 0.0.0.0:00000")
	s.Contains(config, "namespace: tctl-test-namespace")
}

func (s *cliAppSuite) TestDeleteEnvProperty() {
	defer setupConfig(s.app)()

	err := s.app.Run([]string{"", "env", "set", testEnvName + ".address", "1.2.3.4:5678"})
	s.NoError(err)

	err = s.app.Run([]string{"", "env", "delete", testEnvName + ".address"})
	s.NoError(err)

	config := readConfig()
	s.Contains(config, "tctl-test-env:")
	s.Contains(config, "namespace: tctl-test-namespace")
	s.NotContains(config, "address: 1.2.3.4:5678")
}

func (s *cliAppSuite) TestDeleteEnv() {
	defer setupConfig(s.app)()

	err := s.app.Run([]string{"", "env", "set", testEnvName + ".address", "1.2.3.4:5678"})
	s.NoError(err)

	err = s.app.Run([]string{"", "env", "delete", testEnvName})
	s.NoError(err)

	config := readConfig()
	s.NotContains(config, "tctl-test-env:")
	s.NotContains(config, "namespace: tctl-test-namespace")
	s.NotContains(config, "address: 1.2.3.4:5678")
}

func setupConfig(app *cli.App) func() {
	err := app.Run([]string{"", "env", "set", testEnvName + ".namespace", "tctl-test-namespace"})
	if err != nil {
		log.Fatal(err)
	}

	return func() {
		err := app.Run([]string{"", "env", "delete", testEnvName})
		if err != nil {
			log.Printf("unable to unset test env: %s", err)
		}
	}
}

func readConfig() string {
	path := getConfigPath()
	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return string(content)
}

func getConfigPath() string {
	dpath, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	path := filepath.Join(dpath, ".config", "temporalio", "temporal.yaml")

	return path
}
