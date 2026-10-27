package temporalcli

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	commonpb "go.temporal.io/api/common/v1"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestPatchTimeSkippingNilConfig(t *testing.T) {
	for _, tc := range []struct {
		name  string
		patch func() error
	}{
		{
			name: "schedule",
			patch: func() error {
				return patchScheduleTimeSkipping(nil, nil, "YourScheduleId", nil)
			},
		},
		{
			name: "workflow",
			patch: func() error {
				return updateWorkflowOptionsForTimeSkipping(nil, nil, "YourWorkflowId", "", nil)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.EqualError(t, tc.patch(), "time-skipping configuration is required")
		})
	}
}

func TestTimeSkippingConfig(t *testing.T) {
	for _, tc := range []struct {
		name               string
		duration           time.Duration
		hasDuration        bool
		fastForwardID      string
		disablePropagation bool
		maxSkips           int
		wantErr            string
	}{
		{name: "no fast-forward uses defaults"},
		{name: "minimum user skip limit", maxSkips: 1},
		{name: "maximum user skip limit", maxSkips: 10_000},
		{name: "fast-forward with explicit ID", duration: 5 * time.Hour, hasDuration: true, fastForwardID: "YourFastForwardId", disablePropagation: true, maxSkips: 17},
		{name: "fast-forward generates ID", duration: time.Hour, hasDuration: true},
		{name: "fast-forward with maximum user skip limit", duration: time.Hour, hasDuration: true, maxSkips: 10_000},
		{name: "zero duration", hasDuration: true, wantErr: "--fast-forward-duration must be greater than zero"},
		{name: "negative duration", duration: -time.Second, hasDuration: true, wantErr: "--fast-forward-duration must be greater than zero"},
		{name: "ID without duration", fastForwardID: "YourFastForwardId", wantErr: "--fast-forward-id requires --fast-forward-duration"},
		{name: "negative skip limit", maxSkips: -1, wantErr: "--max-skip-per-session must be between 1 and 10000"},
		{name: "skip limit too large", maxSkips: 10_001, wantErr: "--max-skip-per-session must be between 1 and 10000"},
		{name: "fast-forward with skip limit too large", duration: time.Hour, hasDuration: true, maxSkips: 10_001, wantErr: "--max-skip-per-session must be between 1 and 10000"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config, err := timeSkippingConfig(tc.duration, tc.hasDuration, tc.fastForwardID, tc.disablePropagation, tc.maxSkips)
			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
				require.Nil(t, config)
				return
			}
			require.NoError(t, err)
			require.True(t, config.Enabled)
			require.Equal(t, tc.disablePropagation, config.DisablePropagation)
			require.Equal(t, int32(tc.maxSkips), config.MaxSessionSkipCount)
			if !tc.hasDuration {
				require.Nil(t, config.FastForwardConfig)
				return
			}
			require.NotNil(t, config.FastForwardConfig)
			require.Equal(t, tc.duration, config.FastForwardConfig.Duration)
			if tc.fastForwardID == "" {
				require.NotEmpty(t, config.FastForwardConfig.ID)
			} else {
				require.Equal(t, tc.fastForwardID, config.FastForwardConfig.ID)
			}
		})
	}
}

func TestTimeSkippingToPrintable(t *testing.T) {
	targetTime := time.Date(2026, time.September, 27, 12, 0, 0, 0, time.UTC)
	config := &commonpb.TimeSkippingConfig{
		Enabled:             true,
		DisablePropagation:  true,
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
	require.True(t, printable.DisablePropagation)
	require.Equal(t, int32(200), printable.MaxSessionSkipCount)
	require.Equal(t, int32(4), printable.CurrentSessionSkipCount)
	require.Equal(t, "ff-id", printable.FastForwardInfo.Id)
	require.Equal(t, "5h 0m 0s", printable.FastForwardInfo.Duration)
	require.Equal(t, targetTime, printable.FastForwardInfo.TargetTime)
	require.True(t, printable.FastForwardInfo.HasCompleted)

	config.DisablePropagation = false
	require.False(t, timeSkippingToPrintable(info).DisablePropagation)
}

func TestTimeSkippingToPrintableNilValues(t *testing.T) {
	t.Run("nil info", func(t *testing.T) {
		require.Nil(t, timeSkippingToPrintable(nil))
	})

	t.Run("empty info", func(t *testing.T) {
		require.Equal(t, &printableTimeSkippingInfo{}, timeSkippingToPrintable(&commonpb.TimeSkippingInfo{}))
	})

	currentTime := time.Date(2026, time.October, 27, 12, 0, 0, 0, time.UTC)
	targetTime := currentTime.Add(5 * time.Hour)
	// Exercise every combination of the optional config, current time,
	// fast-forward info, duration, and target time. Nested fields cannot be
	// present when their parent fast-forward info is nil.
	for mask := 0; mask < 32; mask++ {
		hasConfig := mask&1 != 0
		hasCurrentTime := mask&2 != 0
		hasFastForward := mask&4 != 0
		hasDuration := mask&8 != 0
		hasTargetTime := mask&16 != 0
		if !hasFastForward && (hasDuration || hasTargetTime) {
			continue
		}
		name := fmt.Sprintf("config=%t/current_time=%t/fast_forward=%t/duration=%t/target_time=%t",
			hasConfig, hasCurrentTime, hasFastForward, hasDuration, hasTargetTime)
		t.Run(name, func(t *testing.T) {
			info := &commonpb.TimeSkippingInfo{CurrentSessionSkipCount: 4}
			want := &printableTimeSkippingInfo{CurrentSessionSkipCount: 4}
			if hasConfig {
				info.EffectiveConfig = &commonpb.TimeSkippingConfig{
					Enabled: true, DisablePropagation: true, MaxSessionSkipCount: 200,
				}
				want.EffectiveEnabled = true
				want.DisablePropagation = true
				want.MaxSessionSkipCount = 200
			}
			if hasCurrentTime {
				info.CurrentTime = timestamppb.New(currentTime)
				want.CurrentTime = currentTime
			}
			if hasFastForward {
				info.FastForwardInfo = &commonpb.TimeSkippingFastForwardInfo{
					FastForwardId: "ff-id", HasCompleted: true,
				}
				want.FastForwardInfo = &printableTimeSkippingFastForwardInfo{
					Id: "ff-id", HasCompleted: true, Duration: "0s",
				}
				if hasDuration {
					info.FastForwardInfo.FastForwardDuration = durationpb.New(5 * time.Hour)
					want.FastForwardInfo.Duration = "5h 0m 0s"
				}
				if hasTargetTime {
					info.FastForwardInfo.TargetTime = timestamppb.New(targetTime)
					want.FastForwardInfo.TargetTime = targetTime
				}
			}
			require.Equal(t, want, timeSkippingToPrintable(info))
		})
	}
}
