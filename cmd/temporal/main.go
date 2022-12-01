package main

import (
	goLog "log"
	"os"

	// Load sqlite storage driver
	_ "go.temporal.io/server/common/persistence/sql/sqlplugin/sqlite"

	"github.com/temporalio/cli/app"
)

// These variables are set by GoReleaser using ldflags
var version string

func main() {
	if err := app.BuildApp(version).Run(os.Args); err != nil {
		goLog.Fatal(err)
	}
}
