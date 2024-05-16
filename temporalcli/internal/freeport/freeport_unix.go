//go:build unix || windows

package freeport

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"syscall"
)

/**
 * Returns a TCP port that is available to listen on, for the given (local) host.
 *
 * This works by binding a new TCP socket on port 0, which requests the OS to
 * allocate a free port. There is no strict guarantee that the port will remain
 * available after this function returns, but it should be safe to assume that
 * a given port will not be allocated again to any process on this machine
 * within a few seconds.
 *
 * On Unix-based systems, binding to the port returned by this function requires
 * setting the `SO_REUSEADDR` socket option (Go already does that by default,
 * but other languages may not); otherwise, the OS may fail with a message such
 * as "address already in use". Windows default behavior is already appropriate
 * in this regard; on that platform, `SO_REUSEADDR` has a different meaning and
 * should not be set (setting it may have unpredictable consequences).
 */
func GetFreePort(host string) (int, error) {
	config := &net.ListenConfig{Control: reuseAddress}
	l, err := config.Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("failed to assign a free port: %v", err)
	}
	defer l.Close()
	port := l.Addr().(*net.TCPAddr).Port

	// On Linux and some BSD variants, ephemeral ports are randomized, and may
	// consequently repeat within a short time frame after the listenning end
	// has been closed. To avoid this, we make a connection to the port, then
	// close that connection from the server's side (this is very important),
	// which puts the connection in TIME_WAIT state for some time (by default,
	// 60s on Linux). While it remains in that state, the OS will not reallocate
	// that port number for bind(:0) syscalls, yet we are not prevented from
	// explicitly binding to it (thanks to SO_REUSEADDR).
	r, err := net.DialTCP("tcp", nil, l.Addr().(*net.TCPAddr))
	if err != nil {
		return 0, fmt.Errorf("failed to assign a free port: %v", err)
	}
	c, err := l.Accept()
	if err != nil {
		return 0, fmt.Errorf("failed to assign a free port: %v", err)
	}
	// Closing the socket from the server side
	c.Close()
	defer r.Close()

	return port, nil
}

func reuseAddress(network, address string, conn syscall.RawConn) error {
	return conn.Control(func(descriptor uintptr) {
		if runtime.GOOS != "windows" {
			syscall.SetsockoptInt(int(descriptor), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
		}
	})
}
