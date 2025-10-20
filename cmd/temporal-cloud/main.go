package main

import (
	"context"

	"github.com/temporalio/cli/temporalcli"
	"github.com/temporalio/cli/temporalcloudcli"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	temporalcloudcli.Execute(ctx, temporalcli.CommandOptions{})
}
