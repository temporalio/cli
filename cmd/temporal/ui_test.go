//go:build !headless

package main

import (
	"runtime/debug"
	"testing"

	"github.com/temporalio/temporal-cli/server"
)

// This test ensures that ui-server is a dependency of Temporal CLI built in non-headless mode.
func TestHasUIServerDependency(t *testing.T) {
	info, _ := debug.ReadBuildInfo()
	for _, dep := range info.Deps {
		if dep.Path == server.UIServerModule {
			return
		}
	}
	t.Errorf("%s should be a dependency when headless tag is not enabled", server.UIServerModule)
	// If the ui-server module name is ever changed, this test should fail and indicate that the
	// module name should be updated for this and the equivalent test case in ui_disabled_test.go
	// to continue working.
	t.Logf("Temporal CLI's %s dependency is missing. Was this module renamed recently?", server.UIServerModule)
}
