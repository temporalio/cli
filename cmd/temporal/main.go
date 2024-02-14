package main

import (
	"context"

	"github.com/temporalio/cli/temporalcli"

	// Embed time zone database as a fallback if platform database can't be found
	_ "time/tzdata"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	temporalcli.Execute(ctx, temporalcli.CommandOptions{})
}
