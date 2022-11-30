package env

import (
	"github.com/temporalio/tctl-kit/pkg/config"
)

func NewClientConfig() (*config.Config, error) {
	return config.NewConfig("temporalio", "temporal")
}
