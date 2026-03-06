// ABOUTME: Provides tsnet (Tailscale) integration for the dev server.
// Exposes the dev server on a Tailscale network via TCP proxy listeners.
package devserver

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"

	"tailscale.com/tsnet"
)

// TsnetOptions configures the tsnet integration for the dev server.
type TsnetOptions struct {
	Hostname     string
	AuthKey      string
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
}

// proxy forwards traffic bidirectionally between src and a TCP connection to dstAddr.
// It closes both connections when either direction finishes.
func proxy(src net.Conn, dstAddr string) {
	dst, err := net.Dial("tcp", dstAddr)
	if err != nil {
		src.Close()
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(dst, src)
		dst.Close()
	}()

	go func() {
		defer wg.Done()
		io.Copy(src, dst)
		src.Close()
	}()

	wg.Wait()
}

// StartTsnet starts a tsnet node and creates TCP proxy listeners that forward
// connections to the local dev server ports.
func StartTsnet(opts TsnetOptions) (*TsnetServer, error) {
	stateDir := opts.StateDir
	if stateDir == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return nil, fmt.Errorf("failed to determine config directory: %w", err)
		}
		stateDir = configDir + "/tsnet-temporal-dev"
	}

	srv := &tsnet.Server{
		Hostname: opts.Hostname,
		AuthKey:  opts.AuthKey,
		Dir:      stateDir,
	}

	if err := srv.Start(); err != nil {
		return nil, fmt.Errorf("failed to start tsnet: %w", err)
	}

	ts := &TsnetServer{
		Hostname: opts.Hostname,
		server:   srv,
		logger:   opts.Logger,
	}

	// Create gRPC proxy listener.
	frontendLn, err := srv.Listen("tcp", fmt.Sprintf(":%d", opts.FrontendPort))
	if err != nil {
		srv.Close()
		return nil, fmt.Errorf("failed to listen on tsnet gRPC port %d: %w", opts.FrontendPort, err)
	}
	ts.listeners = append(ts.listeners, frontendLn)

	go acceptLoop(frontendLn, opts.FrontendAddr)

	// Create UI proxy listener if not headless.
	if opts.UIAddr != "" && opts.UIPort > 0 {
		uiLn, err := srv.Listen("tcp", fmt.Sprintf(":%d", opts.UIPort))
		if err != nil {
			ts.Stop()
			return nil, fmt.Errorf("failed to listen on tsnet UI port %d: %w", opts.UIPort, err)
		}
		ts.listeners = append(ts.listeners, uiLn)

		go acceptLoop(uiLn, opts.UIAddr)
	}

	if opts.Logger != nil {
		opts.Logger.Info("tsnet node started", "hostname", opts.Hostname)
	}

	return ts, nil
}

// acceptLoop accepts connections on ln and proxies each to targetAddr.
func acceptLoop(ln net.Listener, targetAddr string) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		go proxy(conn, targetAddr)
	}
}

// Stop shuts down all proxy listeners and the tsnet node.
func (ts *TsnetServer) Stop() {
	for _, ln := range ts.listeners {
		ln.Close()
	}
	ts.server.Close()
}
