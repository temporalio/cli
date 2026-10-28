package temporalcli

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/client"
)

const maxUserSetSkipPerSession = 10_000

func timeSkippingConfig(duration time.Duration, hasDuration bool, fastForwardID string, disablePropagation bool, maxSkips int) (*client.TimeSkippingConfig, error) {
	if hasDuration && duration <= 0 {
		return nil, errors.New("--fast-forward-duration must be greater than zero")
	}
	if !hasDuration && fastForwardID != "" {
		return nil, errors.New("--fast-forward-id requires --fast-forward-duration")
	}
	if maxSkips < 0 || maxSkips > maxUserSetSkipPerSession {
		return nil, fmt.Errorf("--max-skip-per-session must be between 1 and %d", maxUserSetSkipPerSession)
	}
	config := &client.TimeSkippingConfig{
		Enabled:             true,
		DisablePropagation:  disablePropagation,
		MaxSessionSkipCount: int32(maxSkips),
	}
	if hasDuration {
		if fastForwardID == "" {
			fastForwardID = uuid.NewString()
		}
		config.FastForwardConfig = &client.FastForwardConfig{ID: fastForwardID, Duration: duration}
	}
	return config, nil
}

type printableTimeSkippingInfo struct {
	EffectiveEnabled        bool
	DisablePropagation      bool
	MaxSessionSkipCount     int32
	CurrentSessionSkipCount int32
	CurrentTime             time.Time                             `cli:",cardOmitEmpty"`
	FastForwardInfo         *printableTimeSkippingFastForwardInfo `cli:",cardOmitEmpty"`
}

type printableTimeSkippingFastForwardInfo struct {
	Id           string
	Duration     string
	TargetTime   time.Time `cli:",cardOmitEmpty"`
	HasCompleted bool
}

// timeSkippingToPrintable renders the raw proto form, which is what
// DescribeWorkflowExecution still returns.
func timeSkippingToPrintable(info *commonpb.TimeSkippingInfo) *printableTimeSkippingInfo {
	if info == nil {
		return nil
	}

	out := &printableTimeSkippingInfo{
		EffectiveEnabled:        info.GetEffectiveConfig().GetEnabled(),
		DisablePropagation:      info.GetEffectiveConfig().GetDisablePropagation(),
		MaxSessionSkipCount:     info.GetEffectiveConfig().GetMaxSessionSkipCount(),
		CurrentSessionSkipCount: info.GetCurrentSessionSkipCount(),
	}
	if currentTime := info.GetCurrentTime(); currentTime != nil {
		out.CurrentTime = currentTime.AsTime()
	}
	if ff := info.GetFastForwardInfo(); ff != nil {
		out.FastForwardInfo = &printableTimeSkippingFastForwardInfo{
			Id:           ff.GetFastForwardId(),
			Duration:     formatDuration(ff.GetFastForwardDuration().AsDuration()),
			HasCompleted: ff.GetHasCompleted(),
		}
		if targetTime := ff.GetTargetTime(); targetTime != nil {
			out.FastForwardInfo.TargetTime = targetTime.AsTime()
		}
	}
	return out
}

// clientTimeSkippingToPrintable renders the SDK's typed form, which is what
// ScheduleHandle.Describe returns. The SDK has already done the proto decoding,
// so there are no timestamp nil checks to repeat here.
func clientTimeSkippingToPrintable(info *client.TimeSkippingInfo) *printableTimeSkippingInfo {
	if info == nil {
		return nil
	}

	out := &printableTimeSkippingInfo{
		CurrentTime:             info.CurrentTime,
		CurrentSessionSkipCount: info.CurrentSessionSkipCount,
	}
	if cfg := info.EffectiveConfig; cfg != nil {
		out.EffectiveEnabled = cfg.Enabled
		out.DisablePropagation = cfg.DisablePropagation
		out.MaxSessionSkipCount = cfg.MaxSessionSkipCount
	}
	if ff := info.FastForwardInfo; ff != nil {
		out.FastForwardInfo = &printableTimeSkippingFastForwardInfo{
			Id:           ff.ID,
			Duration:     formatDuration(ff.Duration),
			TargetTime:   ff.TargetTime,
			HasCompleted: ff.HasCompleted,
		}
	}
	return out
}
