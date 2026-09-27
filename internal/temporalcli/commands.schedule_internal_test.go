package temporalcli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	commonpb "go.temporal.io/api/common/v1"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestScheduleTimeSkippingConfig(t *testing.T) {
	config := scheduleTimeSkippingConfig(5 * time.Hour)

	require.True(t, config.Enabled)
	require.NotEmpty(t, config.FastForwardConfig.Id)
	require.Equal(t, 5*time.Hour, config.FastForwardConfig.Duration.AsDuration())
}

func TestTimeSkippingToPrintable(t *testing.T) {
	targetTime := time.Date(2026, time.September, 27, 12, 0, 0, 0, time.UTC)
	config := &commonpb.TimeSkippingConfig{
		Enabled:             true,
		MaxSessionSkipCount: 200,
		FastForwardConfig: &commonpb.FastForwardConfig{
			Id:       "ff-id",
			Duration: durationpb.New(5 * time.Hour),
		},
	}
	info := &commonpb.TimeSkippingInfo{
		CurrentTime:             timestamppb.New(targetTime.Add(-time.Hour)),
		EffectiveConfig:         config,
		CurrentSessionSkipCount: 4,
		FastForwardInfo: &commonpb.TimeSkippingFastForwardInfo{
			FastForwardDuration: durationpb.New(5 * time.Hour),
			FastForwardId:       "ff-id",
			TargetTime:          timestamppb.New(targetTime),
			HasCompleted:        true,
		},
	}

	printable := timeSkippingToPrintable(config, info)
	require.True(t, printable.ConfiguredEnabled)
	require.True(t, printable.EffectiveEnabled)
	require.Equal(t, int32(200), printable.MaxSessionSkipCount)
	require.Equal(t, int32(4), printable.CurrentSessionSkipCount)
	require.Equal(t, "ff-id", printable.FastForward.Id)
	require.Equal(t, "5h 0m 0s", printable.FastForward.Duration)
	require.Equal(t, targetTime, printable.FastForward.TargetTime)
	require.True(t, printable.FastForward.HasCompleted)
}
