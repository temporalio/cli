// Unless explicitly stated otherwise all files in this repository are licensed under the MIT License.
//
// This product includes software developed at Datadog (https://www.datadoghq.com/). Copyright 2021 Datadog, Inc.

//go:build headless

package server

import (
	"runtime/debug"
	"testing"
)

// This test ensures that the ui-server module is not a dependency of Temporal CLI when built
// for headless mode.
func TestNoUIServerDependency(t *testing.T) {
	info, _ := debug.ReadBuildInfo()
	for _, dep := range info.Deps {
		if dep.Path == UIServerModule {
			t.Errorf("%s should not be a dependency when headless tag is enabled", UIServerModule)
		}
	}
}
