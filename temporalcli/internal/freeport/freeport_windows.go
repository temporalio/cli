//go:build windows

package freeport

import (
	"fmt"
	"net"
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
	l, err := net.Listen("tcp", fmt.Sprintf("%v:0", host))
	if err != nil {
		return 0, fmt.Errorf("Failed to assign a free port: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port, nil
}
