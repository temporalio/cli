package temporalcli_test

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/cli/cliext"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

// TODO(cretz): To test:
// * Start server with UI
// * Server reuse existing database file

func TestServer_StartDev_Simple(t *testing.T) {
	port := strconv.Itoa(cliext.MustGetFreePort("127.0.0.1"))
	httpPort := strconv.Itoa(cliext.MustGetFreePort("127.0.0.1"))
	startDevServerAndRunSimpleTest(
		t,
		// TODO(cretz): Remove --headless when
		// https://github.com/temporalio/ui/issues/1773 fixed
		[]string{"server", "start-dev", "-p", port, "--http-port", httpPort, "--headless"},
		"127.0.0.1:"+port,
	)
}

func TestServer_StartDev_IPv4Unspecified(t *testing.T) {
	port := strconv.Itoa(cliext.MustGetFreePort("0.0.0.0"))
	httpPort := strconv.Itoa(cliext.MustGetFreePort("127.0.0.1"))
	startDevServerAndRunSimpleTest(
		t,
		[]string{"server", "start-dev", "--ip", "0.0.0.0", "-p", port, "--http-port", httpPort, "--headless"},
		"0.0.0.0:"+port,
	)
}

func TestServer_StartDev_SQLitePragma(t *testing.T) {
	port := strconv.Itoa(cliext.MustGetFreePort("0.0.0.0"))
	httpPort := strconv.Itoa(cliext.MustGetFreePort("127.0.0.1"))
	dbFilename := filepath.Join(os.TempDir(), "devserver-sqlite-pragma.sqlite")
	defer func() {
		_ = os.Remove(dbFilename)
		_ = os.Remove(dbFilename + "-shm")
		_ = os.Remove(dbFilename + "-wal")
	}()
	startDevServerAndRunSimpleTest(
		t,
		[]string{
			"server", "start-dev",
			"-p", port,
			"--http-port", httpPort,
			"--headless",
			"--db-filename", dbFilename,
			"--sqlite-pragma", "journal_mode=WAL",
			"--sqlite-pragma", "synchronous=NORMAL",
			"--sqlite-pragma", "busy_timeout=5000",
		},
		"0.0.0.0:"+port,
	)
	assert.FileExists(t, dbFilename, "sqlite database file not created")
	assert.FileExists(t, dbFilename+"-shm", "sqlite shared memory file not created")
	assert.FileExists(t, dbFilename+"-wal", "sqlite write-ahead log file not created")
}

func TestServer_StartDev_IPv6Unspecified(t *testing.T) {
	_, err := net.InterfaceByName("::1")
	if err != nil {
		t.Skip("Machine has no IPv6 support")
		return
	}

	port := strconv.Itoa(cliext.MustGetFreePort("::"))
	httpPort := strconv.Itoa(cliext.MustGetFreePort("::"))
	startDevServerAndRunSimpleTest(
		t,
		[]string{
			"server", "start-dev",
			"--ip", "::", "--ui-ip", "::1",
			"-p", port,
			"--http-port", httpPort,
			"--ui-port", strconv.Itoa(cliext.MustGetFreePort("::")),
			"--http-port", strconv.Itoa(cliext.MustGetFreePort("::")),
			"--metrics-port", strconv.Itoa(cliext.MustGetFreePort("::"))},
		"[::]:"+port,
	)
}

func startDevServerAndRunSimpleTest(t *testing.T, args []string, dialAddress string) {
	h := NewCommandHarness(t)
	defer h.Close()

	// Start in background, then wait for client to be able to connect
	resCh := make(chan *CommandResult, 1)
	go func() { resCh <- h.Execute(args...) }()

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
		cl, err = client.Dial(client.Options{HostPort: dialAddress})
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
		port := strconv.Itoa(cliext.MustGetFreePort("127.0.0.1"))
		httpPort := strconv.Itoa(cliext.MustGetFreePort("127.0.0.1"))
		resCh := make(chan *CommandResult, 1)
		go func() {
			resCh <- h.Execute("server", "start-dev", "-p", port, "--http-port", httpPort, "--headless", "--log-level", "never")
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

		select {
		case <-time.After(20 * time.Second):
			h.Fail("didn't cleanup after 20 seconds")
		case res := <-resCh:
			h.NoError(res.Err)
		}
	}

	// Start 40 dev server instances, with 8 concurrent executions
	instanceCounter := atomic.Int32{}
	instanceCounter.Store(40)
	wg := &sync.WaitGroup{}
	for i := 0; i < 6; i++ {
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

func TestServer_StartDev_WithSearchAttributes(t *testing.T) {
	h := NewCommandHarness(t)
	defer h.Close()

	// Start in background, then wait for client to be able to connect
	port := strconv.Itoa(cliext.MustGetFreePort("127.0.0.1"))
	httpPort := strconv.Itoa(cliext.MustGetFreePort("127.0.0.1"))
	resCh := make(chan *CommandResult, 1)
	go func() {
		resCh <- h.Execute(
			"server", "start-dev",
			"-p", port,
			"--http-port", httpPort,
			"--headless",
			"--search-attribute", "search-attr-1=Int",
			"--search-attribute", "search-attr-2=kEyWoRdLiSt",
		)
	}()

	// Try to connect for a bit while checking for error
	var cl client.Client
	h.EventuallyWithT(func(t *assert.CollectT) {
		select {
		case res := <-resCh:
			if res.Err != nil {
				panic(res.Err)
			}
		default:
		}
		var err error
		cl, err = client.Dial(client.Options{HostPort: "127.0.0.1:" + port})
		if !assert.NoError(t, err) {
			return
		}
		// Confirm search attributes are present
		resp, err := cl.OperatorService().ListSearchAttributes(
			context.Background(), &operatorservice.ListSearchAttributesRequest{Namespace: "default"})
		if !assert.NoError(t, err) {
			return
		}
		assert.Contains(t, resp.CustomAttributes, "search-attr-1")
		assert.Contains(t, resp.CustomAttributes, "search-attr-2")

	}, 3*time.Second, 200*time.Millisecond)
	defer cl.Close()

	// Do a workflow start with the search attributes
	run, err := cl.ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			TaskQueue: "my-task-queue",
			TypedSearchAttributes: temporal.NewSearchAttributes(
				temporal.NewSearchAttributeKeyInt64("search-attr-1").ValueSet(123),
				temporal.NewSearchAttributeKeyKeywordList("search-attr-2").ValueSet([]string{"foo", "bar"}),
			),
		},
		"MyWorkflow",
	)
	h.NoError(err)
	h.NotEmpty(run.GetRunID())

	// Check that they are there
	desc, err := cl.DescribeWorkflowExecution(context.Background(), run.GetID(), "")
	h.NoError(err)
	sa := desc.WorkflowExecutionInfo.SearchAttributes.IndexedFields
	h.JSONEq("123", string(sa["search-attr-1"].Data))
	h.JSONEq(`["foo","bar"]`, string(sa["search-attr-2"].Data))

	// Send an interrupt by cancelling context
	h.CancelContext()
	select {
	case <-time.After(20 * time.Second):
		h.Fail("didn't cleanup after 20 seconds")
	case res := <-resCh:
		h.NoError(res.Err)
	}
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
