package devserver

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/server/common/dynamicconfig"
	"go.temporal.io/server/common/log"
)

func timeSkippingConfigForTest(values map[string]any, enabled bool) dynamicconfig.StaticClient {
	config := make(dynamicconfig.StaticClient)
	for key, value := range values {
		config[dynamicconfig.MakeKey(key)] = value
	}
	if enabled {
		applyTimeSkippingDynamicConfig(config)
	}
	return config
}

func TestDevServerTimeSkippingDynamicConfigPrecedence(t *testing.T) {
	type wantConfig struct {
		workflow, schedule, creation, routing bool
		rollout, maxSkip                      int
	}
	serverDefaults := wantConfig{routing: true, maxSkip: 200}
	allEnabled := wantConfig{workflow: true, schedule: true, creation: true, routing: true, rollout: 100, maxSkip: 200}

	for _, tc := range []struct {
		name     string
		enabled  bool
		explicit map[string]any
		want     wantConfig
	}{
		{name: "flag off uses server defaults", want: serverDefaults},
		{name: "flag on without explicit config", enabled: true, want: allEnabled},
		{
			name:    "explicit global config overrides all time-skipping defaults",
			enabled: true,
			explicit: map[string]any{
				dynamicconfig.WorkflowTimeSkippingEnabled.Key().String():          false,
				dynamicconfig.CHASMSchedulerTimeSkippingEnabled.Key().String():    false,
				dynamicconfig.EnableCHASMSchedulerCreation.Key().String():         false,
				dynamicconfig.CHASMSchedulerCreationRolloutPercent.Key().String(): 0,
				dynamicconfig.EnableCHASMSchedulerRouting.Key().String():          false,
			},
			want: wantConfig{maxSkip: 200},
		},
		{
			name:    "partial global override preserves other defaults",
			enabled: true,
			explicit: map[string]any{
				dynamicconfig.CHASMSchedulerTimeSkippingEnabled.Key().String(): false,
			},
			want: wantConfig{workflow: true, creation: true, routing: true, rollout: 100, maxSkip: 200},
		},
		{
			name:    "explicit global max skip overrides server default",
			enabled: true,
			explicit: map[string]any{
				dynamicconfig.TimeSkippingMaxSkipPerSession.Key().String(): 50,
			},
			want: wantConfig{workflow: true, schedule: true, creation: true, routing: true, rollout: 100, maxSkip: 50},
		},
		{
			name: "flag off preserves explicit global config",
			explicit: map[string]any{
				dynamicconfig.WorkflowTimeSkippingEnabled.Key().String(): true,
			},
			want: wantConfig{workflow: true, routing: true, maxSkip: 200},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := timeSkippingConfigForTest(tc.explicit, tc.enabled)
			for key, value := range tc.explicit {
				require.Equal(t, value, raw[dynamicconfig.MakeKey(key)], "explicit key %s was replaced", key)
			}
			for key, fallback := range timeSkippingDynamicConfigDefaults {
				if _, explicit := tc.explicit[key.String()]; explicit {
					continue
				}
				if tc.enabled {
					require.Equal(t, fallback, raw[key], "key %s", key)
				} else {
					_, present := raw[key]
					require.False(t, present, "unexpected override for %s", key)
				}
			}
			config := dynamicconfig.NewCollection(raw, log.NewNoopLogger())
			for _, namespace := range []string{"timeskipping-test", "another-namespace"} {
				t.Run(namespace, func(t *testing.T) {
					require.Equal(t, tc.want.workflow, dynamicconfig.WorkflowTimeSkippingEnabled.Get(config)(namespace))
					require.Equal(t, tc.want.schedule, dynamicconfig.CHASMSchedulerTimeSkippingEnabled.Get(config)(namespace))
					require.Equal(t, tc.want.creation, dynamicconfig.EnableCHASMSchedulerCreation.Get(config)(namespace))
					require.Equal(t, tc.want.rollout, dynamicconfig.CHASMSchedulerCreationRolloutPercent.Get(config)(namespace))
					require.Equal(t, tc.want.routing, dynamicconfig.EnableCHASMSchedulerRouting.Get(config)(namespace))
					require.Equal(t, tc.want.maxSkip, dynamicconfig.TimeSkippingMaxSkipPerSession.Get(config)(namespace))
				})
			}
		})
	}
}
