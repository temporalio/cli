package bridge

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/cli/internal/devserver"
	"go.temporal.io/server/service/localexecution"
)

func TestStartRequiresDurableState(t *testing.T) {
	_, err := Start(context.Background(), StartOptions{
		Configuration: validConfiguration("127.0.0.1:7233"),
		Logger:        testLogger(),
	})
	require.EqualError(t, err, "bridge state store is required")
}

func TestStartAndStop(t *testing.T) {
	upstreamPort := devserver.MustGetFreePort(loopbackAddress)
	upstream := startTestUpstream(t, upstreamPort)
	t.Cleanup(upstream.Stop)
	stateDirectory := filepath.Join(t.TempDir(), "bridge")
	require.NoError(t, os.Mkdir(stateDirectory, 0o700))
	require.NoError(t, os.Chmod(stateDirectory, 0o700))
	stateStore, err := localexecution.OpenBridgeStateStore(stateDirectory)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, stateStore.Close()) })

	ctx, cancel := context.WithCancel(context.Background())
	server, err := Start(ctx, StartOptions{
		Configuration: validConfiguration(upstreamAddress(upstreamPort)),
		StateStore:    stateStore,
		Logger:        testLogger(),
	})
	require.NoError(t, err)
	require.NotEmpty(t, server.FrontendAddress())

	cancel()
	select {
	case err := <-server.Done():
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		require.Fail(t, "bridge runtime did not stop")
	}
	server.Stop()
}

func startTestUpstream(t *testing.T, port int) *devserver.Server {
	t.Helper()
	server, err := devserver.Start(devserver.StartOptions{
		FrontendIP:             loopbackAddress,
		FrontendPort:           port,
		Namespaces:             []string{"namespace"},
		ClusterID:              uuid.NewString(),
		MasterClusterName:      "active",
		CurrentClusterName:     "active",
		InitialFailoverVersion: 1,
		Logger:                 testLogger(),
		LogLevel:               slog.LevelError,
		MetricsPort:            devserver.MustGetFreePort(loopbackAddress),
		EnableGlobalNamespace:  true,
		DynamicConfigValues: map[string]any{
			"history.enableLocalExecution": true,
		},
	})
	require.NoError(t, err)
	return server
}

func validConfiguration(address string) localexecution.BridgeConfiguration {
	return localexecution.BridgeConfiguration{
		Namespace: "namespace",
		Upstream: localexecution.UpstreamConnectionProfile{
			Address: address,
		},
		Options: localexecution.BridgeLocalFirstOptions{
			SyncIntervalMilliseconds:    1_000,
			MaximumUnsynchronizedEvents: 10_240,
			MaximumUnsynchronizedBytes:  8 << 20,
		},
		Registrations: localexecution.WorkerRegistrationManifest{
			TaskQueue:     "task-queue",
			WorkflowTypes: []string{"workflow"},
			ActivityTypes: []string{"activity"},
		},
	}
}

func upstreamAddress(port int) string {
	return "127.0.0.1:" + fmt.Sprint(port)
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
