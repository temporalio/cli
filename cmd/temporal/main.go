package main

import (
	"context"

	"github.com/temporalio/cli/temporalcli"

	// Prevent the pinned version of sqlite driver from unintentionally changing
	// (see Makefile)
	_ "modernc.org/sqlite"
	// Embed time zone database as a fallback if platform database can't be found
	_ "time/tzdata"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	temporalcli.Execute(ctx, temporalcli.CommandOptions{})
}
