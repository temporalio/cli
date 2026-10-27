package devserver

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/server/common/dynamicconfig"
	"go.temporal.io/server/common/log"
)

func TestTimeSkippingDynamicConfigDefaults(t *testing.T) {
	config := dynamicconfig.NewCollection(devServerDynamicConfig(nil), log.NewNoopLogger())
	for _, namespace := range []string{"default", "timeskipping-test"} {
		require.True(t, dynamicconfig.WorkflowTimeSkippingEnabled.Get(config)(namespace))
		require.True(t, dynamicconfig.CHASMSchedulerTimeSkippingEnabled.Get(config)(namespace))
		require.True(t, dynamicconfig.EnableCHASMSchedulerCreation.Get(config)(namespace))
		require.Equal(t, 100, dynamicconfig.CHASMSchedulerCreationRolloutPercent.Get(config)(namespace))
		require.True(t, dynamicconfig.EnableCHASMSchedulerRouting.Get(config)(namespace))
	}
}

func TestTimeSkippingDynamicConfigNamespaceOverride(t *testing.T) {
	values := map[string]any{
		dynamicconfig.WorkflowTimeSkippingEnabled.Key().String(): []dynamicconfig.ConstrainedValue{{
			Value:       false,
			Constraints: dynamicconfig.Constraints{Namespace: "timeskipping-test"},
		}},
		dynamicconfig.CHASMSchedulerTimeSkippingEnabled.Key().String(): []dynamicconfig.ConstrainedValue{{
			Value:       false,
			Constraints: dynamicconfig.Constraints{Namespace: "timeskipping-test"},
		}},
	}
	config := dynamicconfig.NewCollection(devServerDynamicConfig(values), log.NewNoopLogger())
	for _, enabled := range []func(string) bool{
		dynamicconfig.WorkflowTimeSkippingEnabled.Get(config),
		dynamicconfig.CHASMSchedulerTimeSkippingEnabled.Get(config),
	} {
		require.False(t, enabled("timeskipping-test"))
		require.True(t, enabled("another-namespace"))
	}
}

func TestTimeSkippingDynamicConfigGlobalOverride(t *testing.T) {
	key := dynamicconfig.WorkflowTimeSkippingEnabled.Key().String()
	config := dynamicconfig.NewCollection(devServerDynamicConfig(map[string]any{key: false}), log.NewNoopLogger())
	require.False(t, dynamicconfig.WorkflowTimeSkippingEnabled.Get(config)("default"))

	yamlValues := map[string]any{
		key: []dynamicconfig.ConstrainedValue{{Value: false}},
	}
	config = dynamicconfig.NewCollection(devServerDynamicConfig(yamlValues), log.NewNoopLogger())
	require.False(t, dynamicconfig.WorkflowTimeSkippingEnabled.Get(config)("another-namespace"))
}
