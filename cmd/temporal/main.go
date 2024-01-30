package main

import (
	"context"

	"github.com/temporalio/cli/temporalcli"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	temporalcli.Execute(ctx, temporalcli.CommandOptions{})
}
