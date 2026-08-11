package temporalcli_test

import (
	"crypto/tls"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientConnectTimeout(t *testing.T) {
	// Start a listener that accepts connections but never responds
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			defer conn.Close()
			select {}
		}
	}()

	h := NewCommandHarness(t)

	start := time.Now()
	res := h.Execute(
		"workflow", "list",
		"--address", ln.Addr().String(),
		"--client-connect-timeout", (50 * time.Millisecond).String(),
	)
	elapsed := time.Since(start)

	require.Error(t, res.Err)
	assert.Contains(t, res.Err.Error(), "deadline exceeded")
	assert.Less(t, elapsed, time.Second, "dial should have timed out quickly")
}

// TestClientTLSServerNameWithAPIKey confirms that --tls-server-name is sent as the
// SNI when TLS is only implicitly enabled by --api-key, i.e. without also passing
// --tls.
func TestClientTLSServerNameWithAPIKey(t *testing.T) {
	serverNames := startSNIRecordingListener(t)

	h := NewCommandHarness(t)
	// Ignore any TEMPORAL_* variables set in the developer's environment.
	h.Options.EnvLookup = EnvLookupMap{}

	res := h.Execute(
		"workflow", "list",
		"--address", serverNames.address,
		"--api-key", "my-api-key",
		"--tls-server-name", "my-server-name",
		// Ignore any config file present on the developer's machine.
		"--disable-config-file",
	)
	// The handshake is intentionally aborted by the listener, so the command fails.
	require.Error(t, res.Err)

	select {
	case serverName := <-serverNames.ch:
		assert.Equal(t, "my-server-name", serverName)
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for TLS handshake")
	}
}

type sniRecordingListener struct {
	address string
	ch      chan string
}

// startSNIRecordingListener listens for TLS connections and reports the server name
// each client asks for. Handshakes are aborted once the name is recorded, so no
// certificate is needed.
func startSNIRecordingListener(t *testing.T) *sniRecordingListener {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { ln.Close() })

	res := &sniRecordingListener{address: ln.Addr().String(), ch: make(chan string, 1)}
	conf := &tls.Config{
		GetConfigForClient: func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
			select {
			case res.ch <- hello.ServerName:
			default:
			}
			return nil, fmt.Errorf("no certificate configured")
		},
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				// Reading the ClientHello invokes GetConfigForClient above.
				_ = tls.Server(conn, conf).Handshake()
			}()
		}
	}()

	return res
}
