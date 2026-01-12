package temporalcli_test

import (
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
