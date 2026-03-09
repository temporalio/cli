// ABOUTME: Tests for the TCP proxy helper used by the tsnet integration.
// Validates bidirectional forwarding, connection close handling, and unreachable destinations.
package devserver

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestProxyBidirectional(t *testing.T) {
	// Start a TCP echo server.
	echoLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer echoLn.Close()

	go func() {
		for {
			conn, err := echoLn.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				io.Copy(conn, conn)
			}()
		}
	}()

	// Create a pair of connected pipes to simulate client <-> proxy.
	clientConn, proxyConn := net.Pipe()
	defer clientConn.Close()

	// Run proxy in background: proxyConn -> echo server -> proxyConn.
	go proxy(proxyConn, echoLn.Addr().String(), nil, nil)

	// Write data through the client side and read the echo back.
	msg := []byte("hello tsnet proxy")
	if _, err := clientConn.Write(msg); err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, len(msg))
	clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, err := io.ReadFull(clientConn, buf); err != nil {
		t.Fatal(err)
	}

	if string(buf) != string(msg) {
		t.Fatalf("expected %q, got %q", msg, buf)
	}
}

func TestProxyConnectionClose(t *testing.T) {
	// Start a TCP server that immediately closes connections.
	closeLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer closeLn.Close()

	go func() {
		for {
			conn, err := closeLn.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	// Create a pipe pair.
	clientConn, proxyConn := net.Pipe()
	defer clientConn.Close()

	// Run proxy — should not panic when the destination closes immediately.
	done := make(chan struct{})
	go func() {
		defer close(done)
		proxy(proxyConn, closeLn.Addr().String(), nil, nil)
	}()

	// The proxy should finish once the destination closes.
	select {
	case <-done:
		// success
	case <-time.After(5 * time.Second):
		t.Fatal("proxy did not return after destination closed")
	}
}

func TestProxyUnreachableDestination(t *testing.T) {
	// Use a port that nothing is listening on.
	clientConn, proxyConn := net.Pipe()
	defer clientConn.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		proxy(proxyConn, "127.0.0.1:1", nil, nil)
	}()

	// proxy should close the source and return quickly.
	select {
	case <-done:
		// success
	case <-time.After(5 * time.Second):
		t.Fatal("proxy did not return for unreachable destination")
	}

	// Verify the source side is closed — a write should fail.
	clientConn.SetWriteDeadline(time.Now().Add(time.Second))
	_, err := clientConn.Write([]byte("test"))
	if err == nil {
		t.Fatal("expected error writing to closed pipe, got nil")
	}
}

func TestStopIdempotent(t *testing.T) {
	// Create a minimal TsnetServer with real listeners (no actual tsnet.Server needed).
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	_, cancel := context.WithCancel(context.Background())
	ts := &TsnetServer{
		Hostname:  "test",
		listeners: []net.Listener{ln},
		logger:    slog.Default(),
		cancel:    cancel,
	}

	// First Stop should succeed without panic.
	ts.Stop()

	// Second Stop should be a no-op, not panic.
	ts.Stop()
}

func TestProxyLogsDialError(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	clientConn, proxyConn := net.Pipe()
	defer clientConn.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		proxy(proxyConn, "127.0.0.1:1", logger, nil)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("proxy did not return for unreachable destination")
	}

	logged := buf.String()
	if !strings.Contains(logged, "WARN") {
		t.Fatalf("expected WARN level log, got: %s", logged)
	}
	if !strings.Contains(logged, "failed to dial proxy destination") {
		t.Fatalf("expected dial error message in log, got: %s", logged)
	}
}

// mockListener is a test helper that returns pre-configured results from Accept().
type mockListener struct {
	results []acceptResult
	idx     int
}

type acceptResult struct {
	conn net.Conn
	err  error
}

func (m *mockListener) Accept() (net.Conn, error) {
	if m.idx >= len(m.results) {
		return nil, net.ErrClosed
	}
	r := m.results[m.idx]
	m.idx++
	return r.conn, r.err
}

func (m *mockListener) Close() error   { return nil }
func (m *mockListener) Addr() net.Addr { return nil }

func TestAcceptLoopTransientError(t *testing.T) {
	// Start a TCP echo server for the proxy to forward to.
	echoLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer echoLn.Close()

	go func() {
		for {
			conn, err := echoLn.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				io.Copy(conn, conn)
			}()
		}
	}()

	// Create a pipe for the valid connection that acceptLoop will proxy.
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()

	ml := &mockListener{
		results: []acceptResult{
			{nil, errors.New("transient error: too many open files")},
			{serverConn, nil},
			{nil, net.ErrClosed},
		},
	}

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	done := make(chan struct{})
	go func() {
		defer close(done)
		acceptLoop(ml, echoLn.Addr().String(), logger, nil)
	}()

	// Write through the client side and verify echo works through the proxy.
	msg := []byte("after transient error")
	if _, err := clientConn.Write(msg); err != nil {
		t.Fatal(err)
	}

	readBuf := make([]byte, len(msg))
	clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, err := io.ReadFull(clientConn, readBuf); err != nil {
		t.Fatal(err)
	}

	if string(readBuf) != string(msg) {
		t.Fatalf("expected %q, got %q", msg, readBuf)
	}

	// acceptLoop should have returned after net.ErrClosed.
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("acceptLoop did not return after net.ErrClosed")
	}

	// Verify the transient error was logged.
	logged := buf.String()
	if !strings.Contains(logged, "transient error") {
		t.Fatalf("expected transient error in log, got: %s", logged)
	}
}

func TestAcceptLoopTracksGoroutines(t *testing.T) {
	// Start a TCP echo server.
	echoLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer echoLn.Close()

	go func() {
		for {
			conn, err := echoLn.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				io.Copy(conn, conn)
			}()
		}
	}()

	// Create a pipe for the connection that acceptLoop will proxy.
	clientConn, serverConn := net.Pipe()

	ml := &mockListener{
		results: []acceptResult{
			{serverConn, nil},
			{nil, net.ErrClosed},
		},
	}

	var wg sync.WaitGroup

	done := make(chan struct{})
	go func() {
		defer close(done)
		acceptLoop(ml, echoLn.Addr().String(), nil, &wg)
	}()

	// Send data through and verify echo.
	msg := []byte("tracked goroutine")
	if _, err := clientConn.Write(msg); err != nil {
		t.Fatal(err)
	}

	readBuf := make([]byte, len(msg))
	clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, err := io.ReadFull(clientConn, readBuf); err != nil {
		t.Fatal(err)
	}

	if string(readBuf) != string(msg) {
		t.Fatalf("expected %q, got %q", msg, readBuf)
	}

	// Close client side to let proxy finish.
	clientConn.Close()

	// Wait for acceptLoop to exit.
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("acceptLoop did not return")
	}

	// WaitGroup should complete — all tracked goroutines finished.
	wgDone := make(chan struct{})
	go func() {
		defer close(wgDone)
		wg.Wait()
	}()

	select {
	case <-wgDone:
		// All proxy goroutines completed.
	case <-time.After(5 * time.Second):
		t.Fatal("WaitGroup did not complete — proxy goroutines still running")
	}
}

func TestProxyHalfClose(t *testing.T) {
	// Start a TCP server that reads all input, then sends a response, then closes.
	// This pattern requires half-close: the client must signal "done writing" while
	// still being able to read the response.
	serverLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer serverLn.Close()

	response := []byte("response-after-half-close")
	go func() {
		conn, err := serverLn.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		// Read all input until EOF (client did CloseWrite).
		data, _ := io.ReadAll(conn)
		_ = data
		// Send response back.
		conn.Write(response)
	}()

	// Set up: client -> proxy -> server
	// Use real TCP connections so CloseWrite is available.
	proxyClientLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer proxyClientLn.Close()

	// Accept one connection from the proxy side and run proxy on it.
	go func() {
		proxyConn, err := proxyClientLn.Accept()
		if err != nil {
			return
		}
		proxy(proxyConn, serverLn.Addr().String(), nil, nil)
	}()

	// Client connects to the proxy.
	clientConn, err := net.Dial("tcp", proxyClientLn.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer clientConn.Close()

	// Write data, then half-close (signal "done writing").
	if _, err := clientConn.Write([]byte("request-data")); err != nil {
		t.Fatal(err)
	}
	if tc, ok := clientConn.(*net.TCPConn); ok {
		if err := tc.CloseWrite(); err != nil {
			t.Fatal(err)
		}
	}

	// Read the response — this requires the proxy to use half-close, not full close.
	clientConn.SetReadDeadline(time.Now().Add(3 * time.Second))
	got, err := io.ReadAll(clientConn)
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != string(response) {
		t.Fatalf("expected %q, got %q", response, got)
	}
}
