package temporalcli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/cli/internal/devserver"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/server/service/localexecution"
)

// TODO(cretz): To test:
// * Start server with UI
// * Server reuse existing database file

func TestServer_StartBridgeRequiresPersistentStateAndBootstrapToken(t *testing.T) {
	h := NewCommandHarness(t)
	defer h.Close()

	result := h.Execute("server", "start-bridge", "--help")
	require.NoError(t, result.Err)
	assert.Contains(t, result.Stdout.String(), "--state-dir string")
	assert.Contains(t, result.Stdout.String(), "--bootstrap-token-file string")
	assert.NotContains(t, result.Stdout.String(), "--db-filename")
}

func TestServer_StartBridgeRejectsInvalidBootstrapPort(t *testing.T) {
	h := NewCommandHarness(t)
	defer h.Close()

	result := h.Execute(
		"server", "start-bridge",
		"--state-dir", t.TempDir(),
		"--bootstrap-token-file", filepath.Join(t.TempDir(), "token"),
		"--bootstrap-port", "0",
	)
	require.EqualError(t, result.Err, "bootstrap port must be between 1 and 65535")
}

func TestServer_StartBridgeBootstrapLifecycle(t *testing.T) {
	upstreamPort := devserver.MustGetFreePort("127.0.0.1")
	upstream, err := devserver.Start(devserver.StartOptions{
		FrontendIP:             "127.0.0.1",
		FrontendPort:           upstreamPort,
		Namespaces:             []string{"namespace"},
		ClusterID:              uuid.NewString(),
		MasterClusterName:      "active",
		CurrentClusterName:     "active",
		InitialFailoverVersion: 1,
		Logger:                 slog.New(slog.NewTextHandler(io.Discard, nil)),
		LogLevel:               100,
		MetricsPort:            devserver.MustGetFreePort("127.0.0.1"),
		EnableGlobalNamespace:  true,
	})
	require.NoError(t, err)
	t.Cleanup(upstream.Stop)

	stateDirectory := filepath.Join(t.TempDir(), "bridge")
	require.NoError(t, os.Mkdir(stateDirectory, 0o700))
	require.NoError(t, os.Chmod(stateDirectory, 0o700))
	token := "0123456789abcdef0123456789abcdef"
	tokenFile := filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(tokenFile, []byte(token), 0o600))
	bootstrapPort := devserver.MustGetFreePort("127.0.0.1")

	h := NewCommandHarness(t)
	defer h.Close()
	resultChannel := make(chan *CommandResult, 1)
	go func() {
		resultChannel <- h.Execute(
			"server", "start-bridge",
			"--state-dir", stateDirectory,
			"--bootstrap-token-file", tokenFile,
			"--bootstrap-port", strconv.Itoa(bootstrapPort),
			"--log-level", "never",
		)
	}()
	bootstrapAddress := fmt.Sprintf("http://127.0.0.1:%d%s", bootstrapPort, localexecution.BridgeBootstrapPath)
	h.EventuallyWithT(func(t *assert.CollectT) {
		connection, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", bootstrapPort), 100*time.Millisecond)
		assert.NoError(t, err)
		if err == nil {
			assert.NoError(t, connection.Close())
		}
	}, 10*time.Second, 100*time.Millisecond)

	configuration := localexecution.BridgeConfiguration{
		Namespace: "namespace",
		Upstream: localexecution.UpstreamConnectionProfile{
			Address: fmt.Sprintf("127.0.0.1:%d", upstreamPort),
			Headers: map[string]string{"x-bridge-test": "secret-header-value"},
		},
		Options: localexecution.BridgeLocalFirstOptions{
			SyncIntervalMilliseconds:    1_000,
			MaximumUnsynchronizedEvents: 10_240,
			MaximumUnsynchronizedBytes:  8 << 20,
		},
		Registrations: localexecution.WorkerRegistrationManifest{
			TaskQueue: "task-queue",
		},
	}
	response := sendBridgeBootstrap(t, bootstrapAddress, token, configuration)
	require.Equal(t, http.StatusOK, response.StatusCode)
	var bootstrapResult localexecution.BridgeBootstrapResponse
	require.NoError(t, json.NewDecoder(response.Body).Decode(&bootstrapResult))
	require.NoError(t, response.Body.Close())
	require.NotEmpty(t, bootstrapResult.LocalServerID)
	require.NotEmpty(t, bootstrapResult.FrontendAddress)
	_, err = os.Stat(tokenFile)
	require.ErrorIs(t, err, os.ErrNotExist)

	localClient, err := client.Dial(client.Options{
		HostPort:  bootstrapResult.FrontendAddress,
		Namespace: "namespace",
	})
	require.NoError(t, err)
	defer localClient.Close()
	_, err = localClient.WorkflowService().DescribeNamespace(
		t.Context(),
		&workflowservice.DescribeNamespaceRequest{Namespace: "namespace"},
	)
	require.NoError(t, err)

	reused := sendBridgeBootstrap(t, bootstrapAddress, token, configuration)
	require.Equal(t, http.StatusUnauthorized, reused.StatusCode)
	require.NoError(t, reused.Body.Close())
	h.CancelContext()
	select {
	case result := <-resultChannel:
		require.NoError(t, result.Err)
		require.NotContains(t, result.Stdout.String(), token)
		require.NotContains(t, result.Stderr.String(), token)
		require.NotContains(t, result.Stdout.String(), "secret-header-value")
		require.NotContains(t, result.Stderr.String(), "secret-header-value")
	case <-time.After(20 * time.Second):
		require.Fail(t, "bridge command did not stop")
	}
}

func sendBridgeBootstrap(
	t *testing.T,
	address string,
	token string,
	configuration localexecution.BridgeConfiguration,
) *http.Response {
	t.Helper()
	body, err := json.Marshal(configuration)
	require.NoError(t, err)
	request, err := http.NewRequestWithContext(t.Context(), http.MethodPost, address, bytes.NewReader(body))
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	return response
}

func TestServer_StartDev_Simple(t *testing.T) {
	port := strconv.Itoa(devserver.MustGetFreePort("127.0.0.1"))
	httpPort := strconv.Itoa(devserver.MustGetFreePort("127.0.0.1"))
	startDevServerAndRunSimpleTest(
		t,
		// TODO(cretz): Remove --headless when
		// https://github.com/temporalio/ui/issues/1773 fixed
		[]string{"server", "start-dev", "-p", port, "--http-port", httpPort, "--headless"},
		"127.0.0.1:"+port,
	)
}

func TestServer_StartDev_IPv4Unspecified(t *testing.T) {
	port := strconv.Itoa(devserver.MustGetFreePort("0.0.0.0"))
	httpPort := strconv.Itoa(devserver.MustGetFreePort("127.0.0.1"))
	startDevServerAndRunSimpleTest(
		t,
		[]string{"server", "start-dev", "--ip", "0.0.0.0", "-p", port, "--http-port", httpPort, "--headless"},
		"0.0.0.0:"+port,
	)
}

func TestServer_StartDev_SQLitePragma(t *testing.T) {
	port := strconv.Itoa(devserver.MustGetFreePort("0.0.0.0"))
	httpPort := strconv.Itoa(devserver.MustGetFreePort("127.0.0.1"))
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

	port := strconv.Itoa(devserver.MustGetFreePort("::"))
	httpPort := strconv.Itoa(devserver.MustGetFreePort("::"))
	startDevServerAndRunSimpleTest(
		t,
		[]string{
			"server", "start-dev",
			"--ip", "::", "--ui-ip", "::1",
			"-p", port,
			"--http-port", httpPort,
			"--ui-port", strconv.Itoa(devserver.MustGetFreePort("::")),
			"--http-port", strconv.Itoa(devserver.MustGetFreePort("::")),
			"--metrics-port", strconv.Itoa(devserver.MustGetFreePort("::"))},
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
	h := NewCommandHarness(t)
	defer h.Close()

	startOne := func() error {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		// Start in background, then wait for client to be able to connect
		port := strconv.Itoa(devserver.MustGetFreePort("127.0.0.1"))
		httpPort := strconv.Itoa(devserver.MustGetFreePort("127.0.0.1"))
		resCh := make(chan *CommandResult, 1)
		go func() {
			resCh <- h.ExecuteWithContext(ctx, "server", "start-dev", "-p", port, "--http-port", httpPort, "--headless", "--log-level", "never")
		}()

		// Try to connect for a bit while checking for error
		var cl client.Client
		var lastDialErr error
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		timeout := time.NewTimer(3 * time.Second)
		defer timeout.Stop()

	waitForServer:
		for {
			select {
			case res := <-resCh:
				return concurrentStartCommandResultError("got early server result", res)
			case <-ticker.C:
				var err error
				cl, err = client.Dial(client.Options{HostPort: "127.0.0.1:" + port, Logger: testLogger{t: t}})
				if err == nil {
					break waitForServer
				}
				lastDialErr = err
			case <-timeout.C:
				break waitForServer
			}
		}
		if cl == nil {
			cancel()
			select {
			case <-time.After(20 * time.Second):
				return fmt.Errorf("server was not reachable after 3 seconds: %w; also did not clean up within 20 seconds", lastDialErr)
			case res := <-resCh:
				if res.Err != nil {
					return fmt.Errorf(
						"server was not reachable after 3 seconds: %w; cleanup failed: %w",
						lastDialErr,
						res.Err,
					)
				}
				return fmt.Errorf("server was not reachable after 3 seconds: %w", lastDialErr)
			}
		}
		defer cl.Close()

		// Send an interrupt by cancelling context
		cancel()

		select {
		case <-time.After(20 * time.Second):
			return fmt.Errorf("didn't cleanup after 20 seconds")
		case res := <-resCh:
			if res.Err != nil {
				return concurrentStartCommandResultError("server returned error", res)
			}
		}
		return nil
	}

	// Start 40 dev server instances, with 6 concurrent executions.
	instanceCounter := atomic.Int32{}
	instanceCounter.Store(40)
	errCh := make(chan error, 40)
	wg := &sync.WaitGroup{}
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for instanceCounter.Add(-1) >= 0 {
				if err := startOne(); err != nil {
					errCh <- err
				}
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}
}

func concurrentStartCommandResultError(msg string, res *CommandResult) error {
	if res.Err != nil {
		return fmt.Errorf("%s: %w (stdout: %q, stderr: %q)", msg, res.Err, res.Stdout.String(), res.Stderr.String())
	}
	return fmt.Errorf("%s (stdout: %q, stderr: %q)", msg, res.Stdout.String(), res.Stderr.String())
}

func TestServer_StartDev_WithSearchAttributes(t *testing.T) {
	h := NewCommandHarness(t)
	defer h.Close()

	// Start in background, then wait for client to be able to connect
	port := strconv.Itoa(devserver.MustGetFreePort("127.0.0.1"))
	httpPort := strconv.Itoa(devserver.MustGetFreePort("127.0.0.1"))
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

func TestServer_StartDev_BannerPersistenceInMemory(t *testing.T) {
	h := NewCommandHarness(t)
	defer h.Close()

	port := strconv.Itoa(devserver.MustGetFreePort("127.0.0.1"))
	httpPort := strconv.Itoa(devserver.MustGetFreePort("127.0.0.1"))
	resCh := make(chan *CommandResult, 1)
	go func() {
		resCh <- h.Execute("server", "start-dev", "-p", port, "--http-port", httpPort, "--headless")
	}()

	// Wait until the server is dial-able, then cancel
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

	h.CancelContext()
	var res *CommandResult
	select {
	case <-time.After(20 * time.Second):
		h.Fail("didn't cleanup after 20 seconds")
	case res = <-resCh:
		h.NoError(res.Err)
	}
	out := res.Stdout.String()
	h.Contains(out, "Temporal Persistence:")
	h.Contains(out, "in-memory")
}

func TestServer_StartDev_BannerPersistenceFile(t *testing.T) {
	h := NewCommandHarness(t)
	defer h.Close()

	port := strconv.Itoa(devserver.MustGetFreePort("127.0.0.1"))
	httpPort := strconv.Itoa(devserver.MustGetFreePort("127.0.0.1"))
	// Use a unique file in os.TempDir to isolate repeated invocations from prior
	// SQLite state. Explicit cleanup avoids making t.TempDir's os.RemoveAll race
	// a still-open SQLite handle on Windows.
	dbFilename := filepath.Join(os.TempDir(), "devserver-banner-"+uuid.NewString()+".sqlite")
	t.Cleanup(func() {
		_ = os.Remove(dbFilename)
		_ = os.Remove(dbFilename + "-shm")
		_ = os.Remove(dbFilename + "-wal")
	})
	resCh := make(chan *CommandResult, 1)
	go func() {
		resCh <- h.Execute("server", "start-dev", "-p", port, "--http-port", httpPort,
			"--headless", "--db-filename", dbFilename)
	}()

	var cl client.Client
	// File-backed server takes longer to start due to SQLite initialization,
	// especially on slower Windows CI runners.
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
	}, 30*time.Second, 200*time.Millisecond)
	defer cl.Close()

	h.CancelContext()
	var res *CommandResult
	select {
	case <-time.After(20 * time.Second):
		h.Fail("didn't cleanup after 20 seconds")
	case res = <-resCh:
		h.NoError(res.Err)
	}
	out := res.Stdout.String()
	h.Contains(out, "Temporal Persistence:")
	h.Contains(out, dbFilename)
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
