package temporalcli

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/temporalio/cli/internal/bridge"
	"go.temporal.io/server/service/localexecution"
)

const bridgeBootstrapShutdownTimeout = 10 * time.Second

func (t *TemporalServerStartBridgeCommand) run(cctx *CommandContext, _ []string) error {
	if t.BootstrapPort <= 0 || t.BootstrapPort > 65535 {
		return errors.New("bootstrap port must be between 1 and 65535")
	}
	if t.Port < 0 || t.Port > 65535 {
		return errors.New("frontend port must be between 0 and 65535")
	}
	stateStore, err := localexecution.OpenBridgeStateStore(t.StateDir)
	if err != nil {
		return fmt.Errorf("open bridge state: %w", err)
	}
	defer func() {
		if err := stateStore.Close(); err != nil {
			cctx.Logger.Error("failed closing bridge state", "error", err)
		}
	}()

	token, err := localexecution.ReadAndRemoveBootstrapToken(t.BootstrapTokenFile)
	if err != nil {
		return err
	}
	defer func() {
		for i := range token {
			token[i] = 0
		}
	}()
	listener, err := net.Listen(
		"tcp",
		net.JoinHostPort("127.0.0.1", strconv.Itoa(t.BootstrapPort)),
	)
	if err != nil {
		return fmt.Errorf("listen for bridge bootstrap: %w", err)
	}

	var bridgeMutex sync.Mutex
	var bridgeServer *bridge.Server
	runtimeErrors := make(chan error, 1)
	bootstrapServer, err := localexecution.NewBridgeBootstrapServer(
		listener,
		token,
		func(_ context.Context, configuration localexecution.BridgeConfiguration) (localexecution.BridgeBootstrapResponse, error) {
			started, err := bridge.Start(cctx, bridge.StartOptions{
				Configuration: configuration,
				StateStore:    stateStore,
				FrontendPort:  t.Port,
				Logger:        cctx.Logger,
			})
			if err != nil {
				return localexecution.BridgeBootstrapResponse{}, err
			}
			bridgeMutex.Lock()
			bridgeServer = started
			bridgeMutex.Unlock()
			go func() {
				runtimeErrors <- <-started.Done()
			}()
			return localexecution.BridgeBootstrapResponse{
				FrontendAddress: started.FrontendAddress(),
				LocalServerID:   stateStore.LocalServerID(),
			}, nil
		},
	)
	if err != nil {
		_ = listener.Close()
		return err
	}
	defer func() {
		shutdownContext, cancelShutdown := context.WithTimeout(
			context.Background(),
			bridgeBootstrapShutdownTimeout,
		)
		defer cancelShutdown()
		if err := bootstrapServer.Shutdown(shutdownContext); err != nil {
			cctx.Logger.Error("failed shutting down bridge bootstrap server", "error", err)
		}
		bridgeMutex.Lock()
		started := bridgeServer
		bridgeMutex.Unlock()
		if started != nil {
			started.Stop()
		}
	}()

	serveErrors := make(chan error, 1)
	go func() {
		serveErrors <- bootstrapServer.Serve()
	}()
	cctx.Printer.Printlnf("%-21s %s", "Bridge Bootstrap:", bootstrapServer.Address())
	select {
	case <-cctx.Done():
		return nil
	case err := <-serveErrors:
		if err != nil {
			return fmt.Errorf("serve bridge bootstrap endpoint: %w", err)
		}
		return nil
	case err := <-runtimeErrors:
		if err != nil {
			return fmt.Errorf("bridge runtime stopped: %w", err)
		}
		return errors.New("bridge runtime stopped unexpectedly")
	}
}
