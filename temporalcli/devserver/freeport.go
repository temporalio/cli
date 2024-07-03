package devserver

import (
	"fmt"
	"net"
	"runtime"
)

// Returns a TCP port that is available to listen on, for the given (local) host.
//
// This works by binding a new TCP socket on port 0, which requests the OS to
// allocate a free port. There is no strict guarantee that the port will remain
// available after this function returns, but it should be safe to assume that
// a given port will not be allocated again to any process on this machine
// within a few seconds.
//
// On Unix-based systems, binding to the port returned by this function requires
// setting the `SO_REUSEADDR` socket option (Go already does that by default,
// but other languages may not); otherwise, the OS may fail with a message such
// as "address already in use". Windows default behavior is already appropriate
// in this regard; on that platform, `SO_REUSEADDR` has a different meaning and
// should not be set (setting it may have unpredictable consequences).
func GetFreePort(host string) (int, error) {
	host = MaybeEscapeIPv6(host)
	l, err := net.Listen("tcp", host+":0")
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
	//
	// On macOS and Windows, the above technique is not necessary, as the OS
	// allocates ephemeral ports sequentially, meaning a port number will only
	// be reused after the entire range has been exhausted. Quite the opposite,
	// given that these OSes use a significantly smaller range for ephemeral
	// ports, making an extra connection just to reserve a port might actually
	// be harmful (by hastening ephemeral port exhaustion).
	if runtime.GOOS != "darwin" && runtime.GOOS != "windows" {
		// DialTCP(..., l.Addr()) might fail if machine has IPv6 support, but
		// isn't fully configured (e.g. doesn't have a loopback interface bound
		// to ::1). For safety, rebuild address form the original host instead.
		tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
		if err != nil {
			return 0, fmt.Errorf("error resolving address: %v", err)
		}
		r, err := net.DialTCP("tcp", nil, tcpAddr)
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
	}

	return port, nil
}

// Returns a TCP port that is available to listen on, for the given (local)
// host; panics if no port is available.
//
// This works by binding a new TCP socket on port 0, which requests the OS to
// allocate a free port. There is no strict guarantee that the port will remain
// available after this function returns, but it should be safe to assume that
// a given port will not be allocated again to any process on this machine
// within a few seconds.
//
// On Unix-based systems, binding to the port returned by this function requires
// setting the `SO_REUSEADDR` socket option (Go already does that by default,
// but other languages may not); otherwise, the OS may fail with a message such
// as "address already in use". Windows default behavior is already appropriate
// in this regard; on that platform, `SO_REUSEADDR` has a different meaning and
// should not be set (setting it may have unpredictable consequences).
func MustGetFreePort(host string) int {
	port, err := GetFreePort(host)
	if err != nil {
		panic(fmt.Errorf("failed assigning ephemeral port: %w", err))
	}
	return port
}

// Asserts that the given TCP port is available to listen on, for the given
// (local) host; return an error if it is not.
func CheckPortFree(host string, port int) error {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", MaybeEscapeIPv6(host), port))
	if err != nil {
		return err
	}
	l.Close()
	return nil
}

// Escapes an IPv6 address with square brackets, if it is an IPv6 address.
func MaybeEscapeIPv6(host string) string {
	if ip := net.ParseIP(host); ip != nil && ip.To4() == nil {
		return "[" + host + "]"
	}
	return host
}
