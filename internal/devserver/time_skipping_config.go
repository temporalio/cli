// The MIT License
//
// Copyright (c) 2022 Temporal Technologies Inc.  All rights reserved.
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Copyright (c) 2021 Datadog, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

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
		value, exists := dynConf[key]
		if !exists {
			dynConf[key] = fallback
			continue
		}
		// Explicit global values win. Constrained values need the default as a
		// fallback for namespaces they do not match.
		constrained, ok := value.([]dynamicconfig.ConstrainedValue)
		if !ok {
			continue
		}
		hasGlobal := false
		for _, entry := range constrained {
			if entry.Constraints == (dynamicconfig.Constraints{}) {
				hasGlobal = true
				break
			}
		}
		if !hasGlobal {
			dynConf[key] = append(append([]dynamicconfig.ConstrainedValue(nil), constrained...), dynamicconfig.ConstrainedValue{Value: fallback})
		}
	}
}
