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
	config, err := scheduleTimeSkippingConfig(5*time.Hour, "", true, 17)
	require.NoError(t, err)

	require.True(t, config.Enabled)
	require.True(t, config.DisablePropagation)
	require.Equal(t, int32(17), config.MaxSessionSkipCount)
	require.NotEmpty(t, config.FastForwardConfig.ID)
	require.Equal(t, 5*time.Hour, config.FastForwardConfig.Duration)

	config, err = scheduleTimeSkippingConfig(5*time.Hour, "YourFastForwardId", false, 0)
	require.NoError(t, err)
	require.Equal(t, "YourFastForwardId", config.FastForwardConfig.ID)
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

	printable := timeSkippingToPrintable(info)
	require.True(t, printable.EffectiveEnabled)
	require.Equal(t, int32(200), printable.MaxSessionSkipCount)
	require.Equal(t, int32(4), printable.CurrentSessionSkipCount)
	require.Equal(t, "ff-id", printable.FastForward.Id)
	require.Equal(t, "5h 0m 0s", printable.FastForward.Duration)
	require.Equal(t, targetTime, printable.FastForward.TargetTime)
	require.True(t, printable.FastForward.HasCompleted)
}

func TestScheduleTimeSkippingPollStatus(t *testing.T) {
	const id = "YourFastForwardId"
	for _, tc := range []struct {
		name   string
		info   *commonpb.TimeSkippingInfo
		status string
		reason string
	}{
		{name: "removed", status: "failed", reason: "fast-forward was removed or was not found in Schedule status"},
		{name: "replaced", info: &commonpb.TimeSkippingInfo{FastForwardInfo: &commonpb.TimeSkippingFastForwardInfo{FastForwardId: "other"}}, status: "failed", reason: `fast-forward was replaced by "other"`},
		{name: "completed", info: &commonpb.TimeSkippingInfo{FastForwardInfo: &commonpb.TimeSkippingFastForwardInfo{FastForwardId: id, HasCompleted: true}}, status: "completed"},
		{name: "disabled", info: &commonpb.TimeSkippingInfo{FastForwardInfo: &commonpb.TimeSkippingFastForwardInfo{FastForwardId: id}}, status: "failed", reason: "time skipping was disabled before fast-forward completed"},
		{name: "pending", info: &commonpb.TimeSkippingInfo{EffectiveConfig: &commonpb.TimeSkippingConfig{Enabled: true}, FastForwardInfo: &commonpb.TimeSkippingFastForwardInfo{FastForwardId: id}}, status: "pending"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			status, reason := scheduleTimeSkippingPollStatus(tc.info, id)
			require.Equal(t, tc.status, status)
			require.Equal(t, tc.reason, reason)
		})
	}
}
