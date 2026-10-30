package devserver

import "go.temporal.io/server/common/dynamicconfig"

var timeSkippingDynamicConfigDefaults = map[dynamicconfig.Key]any{
	dynamicconfig.WorkflowTimeSkippingEnabled.Key():          true,
	dynamicconfig.CHASMSchedulerTimeSkippingEnabled.Key():    true,
	dynamicconfig.EnableCHASMSchedulerCreation.Key():         true,
	dynamicconfig.CHASMSchedulerCreationRolloutPercent.Key(): 100,
	dynamicconfig.EnableCHASMSchedulerRouting.Key():          true,
}

func applyTimeSkippingDynamicConfig(dynConf dynamicconfig.StaticClient) {
	for key, fallback := range timeSkippingDynamicConfigDefaults {
		// Explicit dynamic config values take precedence over these defaults.
		if _, exists := dynConf[key]; !exists {
			dynConf[key] = fallback
		}
	}
}
