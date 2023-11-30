package temporalcli_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/cli/temporalcli/devserver"
	"go.temporal.io/sdk/client"
)

// TODO(cretz): To test:
// * Start server with UI
// * Server reuse existing database file

func TestServerStartDev_Simple(t *testing.T) {
	h := NewCommandHarness(t)
	defer h.Close()

	// Start in background, then wait for client to be able to connect
	port := strconv.Itoa(devserver.MustGetFreePort())
	resCh := make(chan *CommandResult, 1)
	// TODO(cretz): Remove --headless when
	// https://github.com/temporalio/ui/issues/1773 fixed
	go func() { resCh <- h.Execute("server", "start-dev", "-p", port, "--headless") }()

	// Try to connect for a bit while checking for error
	var cl client.Client
	h.EventuallyWithT(func(t *assert.CollectT) {
		select {
		case res := <-resCh:
			require.NoError(t, res.Err)
			require.Fail(t, "got early server result")
		default:
		}
		var err error
		cl, err = client.Dial(client.Options{HostPort: "127.0.0.1:" + port})
		require.NoError(t, err)
	}, 3*time.Second, 200*time.Millisecond)
	defer cl.Close()

	// Just a simple workflow start will suffice for now
	run, err := cl.ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{TaskQueue: "my-task-queue"},
		"MyWorkflow",
	)
	h.NoError(err)
	h.NotEmpty(run.GetRunID())

	// Send an interrupt by cancelling context
	h.CancelContext()
	select {
	case <-time.After(10 * time.Second):
		h.Fail("didn't cleanup after 10 seconds")
	case res := <-resCh:
		h.NoError(res.Err)
	}
}
