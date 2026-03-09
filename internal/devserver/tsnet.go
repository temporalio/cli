// ABOUTME: Provides tsnet (Tailscale) integration for the dev server.
// Exposes the dev server on a Tailscale network via TCP proxy listeners.
package devserver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"sync"

	"tailscale.com/tsnet"
)

// TsnetOptions configures the tsnet integration for the dev server.
type TsnetOptions struct {
	Hostname string
	// AuthKey is the Tailscale auth key. At the call site this is populated
	// from the generated TsnetAuthkey field (--tsnet-authkey flag). The name
	// difference is intentional: the flag uses the CLI naming convention while
	// this field uses the domain term.
	AuthKey string
	StateDir     string
	FrontendAddr string // local gRPC address, e.g. "127.0.0.1:7233"
	UIAddr       string // local UI address, e.g. "127.0.0.1:8233" (empty if headless)
	FrontendPort int    // port for tsnet gRPC listener
	UIPort       int    // port for tsnet UI listener (0 if headless)
	Logger       *slog.Logger
}

// TsnetServer holds the running tsnet node and its proxy listeners.
type TsnetServer struct {
	Hostname  string
	server    *tsnet.Server
	listeners []net.Listener
	logger    *slog.Logger
	ctx       context.Context
	cancel    context.CancelFunc
	stopOnce  sync.Once
	wg        sync.WaitGroup // tracks in-flight proxy goroutines
}

// halfCloser is implemented by connections that support TCP half-close.
type halfCloser interface {
	CloseWrite() error
}

// isClosedErr reports whether err is a benign connection-closed error.
func isClosedErr(err error) bool {
	return errors.Is(err, net.ErrClosed) ||
		errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrClosedPipe)
}

// proxy forwards traffic bidirectionally between src and a TCP connection to dstAddr.
// It uses TCP half-close when available so that one direction can signal "done writing"
// without killing the other direction.
// If parentWg is non-nil, it calls parentWg.Done() when the proxy completes.
func proxy(src net.Conn, dstAddr string, logger *slog.Logger, parentWg *sync.WaitGroup) {
	if parentWg != nil {
		defer parentWg.Done()
	}
	dst, err := net.Dial("tcp", dstAddr)
	if err != nil {
		if logger != nil {
			logger.Warn("failed to dial proxy destination", "addr", dstAddr, "error", err)
		}
		src.Close()
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// src -> dst: when src is done sending, half-close dst's write side.
	go func() {
		defer wg.Done()
		_, err := io.Copy(dst, src)
		if err != nil && !isClosedErr(err) && logger != nil {
			logger.Debug("proxy copy src->dst ended", "error", err)
		}
		if hc, ok := dst.(halfCloser); ok {
			hc.CloseWrite()
		} else {
			dst.Close()
		}
	}()

	// dst -> src: when dst is done sending, half-close src's write side.
	go func() {
		defer wg.Done()
		_, err := io.Copy(src, dst)
		if err != nil && !isClosedErr(err) && logger != nil {
			logger.Debug("proxy copy dst->src ended", "error", err)
		}
		if hc, ok := src.(halfCloser); ok {
			hc.CloseWrite()
		} else {
			src.Close()
		}
	}()

	wg.Wait()
	// Full cleanup after both directions are done.
	src.Close()
	dst.Close()
}

// StartTsnet starts a tsnet node and creates TCP proxy listeners that forward
// connections to the local dev server ports.
func StartTsnet(ctx context.Context, opts TsnetOptions) (*TsnetServer, error) {
	ctx, cancel := context.WithCancel(ctx)

	stateDir := opts.StateDir
	if stateDir == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to determine config directory: %w", err)
		}
		stateDir = filepath.Join(configDir, "tsnet-temporal-dev")
	}

	srv := &tsnet.Server{
		Hostname: opts.Hostname,
		AuthKey:  opts.AuthKey,
		Dir:      stateDir,
	}

	if err := srv.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start tsnet: %w", err)
	}

	ts := &TsnetServer{
		Hostname: opts.Hostname,
		server:   srv,
		logger:   opts.Logger,
		ctx:      ctx,
		cancel:   cancel,
	}

	// Create gRPC proxy listener.
	frontendLn, err := srv.Listen("tcp", fmt.Sprintf(":%d", opts.FrontendPort))
	if err != nil {
		cancel()
		srv.Close()
		return nil, fmt.Errorf("failed to listen on tsnet gRPC port %d: %w", opts.FrontendPort, err)
	}
	ts.listeners = append(ts.listeners, frontendLn)

	go acceptLoop(frontendLn, opts.FrontendAddr, ts.logger, &ts.wg)

	// Create UI proxy listener if not headless.
	if opts.UIAddr != "" && opts.UIPort > 0 {
		uiLn, err := srv.Listen("tcp", fmt.Sprintf(":%d", opts.UIPort))
		if err != nil {
			ts.Stop()
			return nil, fmt.Errorf("failed to listen on tsnet UI port %d: %w", opts.UIPort, err)
		}
		ts.listeners = append(ts.listeners, uiLn)

		go acceptLoop(uiLn, opts.UIAddr, ts.logger, &ts.wg)
	}

	if opts.Logger != nil {
		opts.Logger.Info("tsnet node started", "hostname", opts.Hostname)
	}

	return ts, nil
}

// acceptLoop accepts connections on ln and proxies each to targetAddr.
// It returns silently on net.ErrClosed (expected shutdown) and retries on transient errors.
// If wg is non-nil, each proxy goroutine is tracked via wg.Add/Done.
func acceptLoop(ln net.Listener, targetAddr string, logger *slog.Logger, wg *sync.WaitGroup) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			if logger != nil {
				logger.Warn("accept error, retrying", "error", err)
			}
			continue
		}
		if wg != nil {
			wg.Add(1)
		}
		go proxy(conn, targetAddr, logger, wg)
	}
}

// Stop shuts down all proxy listeners and the tsnet node.
// It is safe to call multiple times; only the first call has any effect.
func (ts *TsnetServer) Stop() {
	ts.stopOnce.Do(func() {
		if ts.cancel != nil {
			ts.cancel()
		}
		for _, ln := range ts.listeners {
			if err := ln.Close(); err != nil && ts.logger != nil {
				ts.logger.Warn("failed to close tsnet listener", "error", err)
			}
		}
		// Wait for all in-flight proxy goroutines to finish before closing the server.
		ts.wg.Wait()
		if ts.server != nil {
			if err := ts.server.Close(); err != nil && ts.logger != nil {
				ts.logger.Warn("failed to close tsnet server", "error", err)
			}
		}
	})
}
