// ABOUTME: Tests for the TCP proxy helper used by the tsnet integration.
// Validates bidirectional forwarding, connection close handling, and unreachable destinations.
package devserver

import (
	"io"
	"net"
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
	go proxy(proxyConn, echoLn.Addr().String())

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
		proxy(proxyConn, closeLn.Addr().String())
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
		proxy(proxyConn, "127.0.0.1:1")
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
