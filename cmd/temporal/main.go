package main

import (
	goLog "log"
	"os"

	// Load sqlite storage driver
	_ "go.temporal.io/server/common/persistence/sql/sqlplugin/sqlite"

	"github.com/temporalio/cli/app"
)

func main() {
	if err := app.BuildApp().Run(os.Args); err != nil {
		goLog.Fatal(err)
	}
}
