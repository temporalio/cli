//go:build headless

package main

import (
	"runtime/debug"
	"testing"
)

// This test ensures that the ui-server module is not a dependency of Temporal CLI when built
// for headless mode.
func TestNoUIServerDependency(t *testing.T) {
	info, _ := debug.ReadBuildInfo()
	for _, dep := range info.Deps {
		if dep.Path == server.UIServerModule {
			t.Errorf("%s should not be a dependency when headless tag is enabled", server.UIServerModule)
		}
	}
}
