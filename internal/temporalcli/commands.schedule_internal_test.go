package temporalcli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestScheduleTimeSkippingConfig(t *testing.T) {
	config := scheduleTimeSkippingConfig(5 * time.Hour)

	require.True(t, config.Enabled)
	require.NotEmpty(t, config.FastForwardConfig.Id)
	require.Equal(t, 5*time.Hour, config.FastForwardConfig.Duration.AsDuration())
}
