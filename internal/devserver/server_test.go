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
	disabled := map[dynamicconfig.Key]any{
		dynamicconfig.WorkflowTimeSkippingEnabled.Key():          false,
		dynamicconfig.CHASMSchedulerTimeSkippingEnabled.Key():    false,
		dynamicconfig.EnableCHASMSchedulerCreation.Key():         false,
		dynamicconfig.CHASMSchedulerCreationRolloutPercent.Key(): 0,
		dynamicconfig.EnableCHASMSchedulerRouting.Key():          false,
	}
	type wantConfig struct {
		workflow, schedule, creation, routing bool
		rollout, maxSkip                      int
	}
	serverDefaults := wantConfig{routing: true, maxSkip: 200}
	allEnabled := wantConfig{workflow: true, schedule: true, creation: true, routing: true, rollout: 100, maxSkip: 200}
	allDisabled := wantConfig{maxSkip: 200}
	granularDisabled := wantConfig{maxSkip: 50}

	granular := make(map[string]any, len(disabled)+1)
	globalScalar := make(map[string]any, len(disabled))
	globalYAML := make(map[string]any, len(disabled))
	for key, value := range disabled {
		granular[key.String()] = []dynamicconfig.ConstrainedValue{{
			Value: value, Constraints: dynamicconfig.Constraints{Namespace: "timeskipping-test"},
		}}
		globalScalar[key.String()] = value
		globalYAML[key.String()] = []dynamicconfig.ConstrainedValue{{Value: value}}
	}
	granular[dynamicconfig.TimeSkippingMaxSkipPerSession.Key().String()] = []dynamicconfig.ConstrainedValue{{
		Value: 50, Constraints: dynamicconfig.Constraints{Namespace: "timeskipping-test"},
	}}

	for _, tc := range []struct {
		name     string
		enabled  bool
		explicit map[string]any
		wantTest wantConfig
		wantElse wantConfig
		shape    string
	}{
		{name: "flag off", wantTest: serverDefaults, wantElse: serverDefaults, shape: "absent"},
		{name: "no explicit config", enabled: true, wantTest: allEnabled, wantElse: allEnabled, shape: "defaults"},
		{name: "namespace explicit config", enabled: true, explicit: granular, wantTest: granularDisabled, wantElse: allEnabled, shape: "granular"},
		{name: "global explicit scalar config", enabled: true, explicit: globalScalar, wantTest: allDisabled, wantElse: allDisabled, shape: "global"},
		{name: "global explicit YAML config", enabled: true, explicit: globalYAML, wantTest: allDisabled, wantElse: allDisabled, shape: "global"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := timeSkippingConfigForTest(tc.explicit, tc.enabled)
			for key, fallback := range timeSkippingDynamicConfigDefaults {
				switch tc.shape {
				case "absent":
					_, present := raw[key]
					require.False(t, present, "unexpected override for %s", key)
				case "defaults":
					require.Equal(t, fallback, raw[key], "key %s", key)
				case "granular":
					require.Len(t, tc.explicit[key.String()].([]dynamicconfig.ConstrainedValue), 1, "input for %s was mutated", key)
					require.Equal(t, []dynamicconfig.ConstrainedValue{
						{Value: disabled[key], Constraints: dynamicconfig.Constraints{Namespace: "timeskipping-test"}},
						{Value: fallback},
					}, raw[key], "key %s", key)
				case "global":
					require.Equal(t, tc.explicit[key.String()], raw[key], "key %s was replaced or appended to", key)
				}
			}
			config := dynamicconfig.NewCollection(raw, log.NewNoopLogger())
			for namespace, want := range map[string]wantConfig{
				"timeskipping-test": tc.wantTest,
				"another-namespace": tc.wantElse,
			} {
				t.Run(namespace, func(t *testing.T) {
					require.Equal(t, want.workflow, dynamicconfig.WorkflowTimeSkippingEnabled.Get(config)(namespace))
					require.Equal(t, want.schedule, dynamicconfig.CHASMSchedulerTimeSkippingEnabled.Get(config)(namespace))
					require.Equal(t, want.creation, dynamicconfig.EnableCHASMSchedulerCreation.Get(config)(namespace))
					require.Equal(t, want.rollout, dynamicconfig.CHASMSchedulerCreationRolloutPercent.Get(config)(namespace))
					require.Equal(t, want.routing, dynamicconfig.EnableCHASMSchedulerRouting.Get(config)(namespace))
					require.Equal(t, want.maxSkip, dynamicconfig.TimeSkippingMaxSkipPerSession.Get(config)(namespace))
				})
			}
		})
	}
}
