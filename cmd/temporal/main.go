package main

import (
	"context"
	"os"

	"github.com/temporalio/cli/internal/temporalcli"

	// Prevent the pinned version of sqlite driver from unintentionally changing
	// until https://gitlab.com/cznic/sqlite/-/issues/196 is resolved.
	_ "modernc.org/sqlite"
	// Embed time zone database as a fallback if platform database can't be found
	_ "time/tzdata"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := temporalcli.Execute(ctx, temporalcli.CommandOptions{})
	os.Exit(result.ExitStatus)
}
