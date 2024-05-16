package temporalcli_test

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/cli/temporalcli/internal/freeport"
	"go.temporal.io/sdk/client"
)

// TODO(cretz): To test:
// * Start server with UI
// * Server reuse existing database file

func TestServer_StartDev_Simple(t *testing.T) {
	h := NewCommandHarness(t)
	defer h.Close()

	// Start in background, then wait for client to be able to connect
	port := strconv.Itoa(freeport.MustGetFreePort("127.0.0.1"))
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
		assert.NoError(t, err)
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
	case <-time.After(20 * time.Second):
		h.Fail("didn't cleanup after 20 seconds")
	case res := <-resCh:
		h.NoError(res.Err)
	}
}

func TestServer_StartDev_ConcurrentStarts(t *testing.T) {
	startOne := func() {
		h := NewCommandHarness(t)
		defer h.Close()

		// Start in background, then wait for client to be able to connect
		port := strconv.Itoa(freeport.MustGetFreePort("127.0.0.1"))
		resCh := make(chan *CommandResult, 1)
		go func() {
			resCh <- h.Execute("server", "start-dev", "-p", port, "--headless", "--log-level", "never")
		}()

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
			cl, err = client.Dial(client.Options{HostPort: "127.0.0.1:" + port, Logger: testLogger{t: h.t}})
			assert.NoError(t, err)
		}, 3*time.Second, 200*time.Millisecond)
		defer cl.Close()

		// Send an interrupt by cancelling context
		h.CancelContext()

		// FIXME: We should technically wait for server cleanup, but this is
		// slowing down the test considerably, presumably due to the issue fixed
		// in https://github.com/temporalio/temporal/pull/5459. Uncomment the
		// following code when the server dependency is updated to 1.24.0.
		//
		// select {
		// case <-time.After(20 * time.Second):
		// 	h.Fail("didn't cleanup after 20 seconds")
		// case res := <-resCh:
		// 	h.NoError(res.Err)
		// }
	}

	// Start 200 dev server instances, with 16 concurrent executions
	instanceCounter := atomic.Int32{}
	instanceCounter.Store(200)
	wg := &sync.WaitGroup{}
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			for instanceCounter.Add(-1) >= 0 {
				startOne()
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

type testLogger struct {
	t *testing.T
}

func (l testLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.t.Logf("DEBUG: "+msg, keysAndValues...)
}

func (l testLogger) Info(msg string, keysAndValues ...interface{}) {
	l.t.Logf("INFO: "+msg, keysAndValues...)
}

func (l testLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.t.Logf("WARN: "+msg, keysAndValues...)
}

func (l testLogger) Error(msg string, keysAndValues ...interface{}) {
	l.t.Logf("ERROR: "+msg, keysAndValues...)
}
