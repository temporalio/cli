package main

import (
	goLog "log"
	"os"

	// Load sqlite storage driver
	_ "go.temporal.io/server/common/persistence/sql/sqlplugin/sqlite"

	// Embed time zone database as a fallback if platform database can't be found
	_ "time/tzdata"

	"github.com/temporalio/cli/app"
)

func main() {
	if err := app.BuildApp().Run(os.Args); err != nil {
		goLog.Fatal(err)
	}
}
