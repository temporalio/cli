package temporalcli_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// startBlackHoleListener starts a listener that accepts connections but never
// responds.
func startBlackHoleListener(t *testing.T) net.Listener {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { ln.Close() })
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
	return ln
}

// startMTLSListener starts a TLS listener that requires client certificates,
// returning its address and the server CA cert as PEM.
func startMTLSListener(t *testing.T) (addr, caPEM string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "client-test"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	require.NoError(t, err)
	leaf, err := x509.ParseCertificate(der)
	require.NoError(t, err)
	pool := x509.NewCertPool()
	pool.AddCert(leaf)

	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{
		Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    pool,
	})
	require.NoError(t, err)
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				if tlsConn, ok := conn.(*tls.Conn); ok {
					_ = tlsConn.HandshakeContext(context.Background())
					buf := make([]byte, 1)
					_ = tlsConn.SetReadDeadline(time.Now().Add(2 * time.Second))
					_, _ = tlsConn.Read(buf)
				}
				conn.Close()
			}()
		}
	}()
	return ln.Addr().String(), string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}

func TestClientConnectTimeout(t *testing.T) {
	ln := startBlackHoleListener(t)

	h := NewCommandHarness(t)

	start := time.Now()
	res := h.Execute(
		"workflow", "list",
		"--address", ln.Addr().String(),
		"--client-connect-timeout", (50 * time.Millisecond).String(),
	)
	elapsed := time.Since(start)

	require.Error(t, res.Err)
	assert.Contains(t, res.Err.Error(), "failed connecting to Temporal server at")
	assert.Contains(t, res.Err.Error(), "deadline exceeded")
	// The friendly cause attached to the connect timeout must survive.
	assert.Contains(t, res.Err.Error(), "command timed out after 50ms")
	assert.Less(t, elapsed, 4*time.Second, "diagnosis should respect its independent three-second cap")
}

func TestConnectDiagnosis_MTLSRequired(t *testing.T) {
	addr, caPEM := startMTLSListener(t)

	h := NewCommandHarness(t)
	res := h.Execute(
		"workflow", "list",
		"--address", addr,
		"--tls",
		"--tls-ca-data", caPEM,
		"--client-connect-timeout", "2s",
	)

	require.Error(t, res.Err)
	msg := res.Stderr.String()
	assert.Contains(t, msg, "Connecting to "+addr)
	assert.Contains(t, msg, "✓ TCP connection established")
	assert.Contains(t, msg, "✗ TLS handshake failed: server requires mTLS")
	assert.Contains(t, msg, "tls.client_cert_path --value YourCert.pem")
	assert.Contains(t, msg, "tls.client_key_path --value YourKey.pem")
}

func TestConnectDiagnosis_Refused(t *testing.T) {
	// Grab a port and close it so nothing is listening.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	require.NoError(t, ln.Close())

	h := NewCommandHarness(t)
	res := h.Execute(
		"workflow", "list",
		"--address", addr,
		"--client-connect-timeout", "2s",
	)

	require.Error(t, res.Err)
	msg := res.Stderr.String()
	assert.Contains(t, msg, "connection refused")
	assert.Contains(t, msg, "✗ TCP connection refused")
	assert.Contains(t, msg, "Nothing is listening at "+addr)
}

func TestConnectDiagnosis_DNSFailure(t *testing.T) {
	h := NewCommandHarness(t)
	res := h.Execute(
		"workflow", "list",
		// .invalid is reserved (RFC 2606) and can never resolve.
		"--address", "does-not-exist.invalid:7233",
		"--client-connect-timeout", "5s",
	)

	require.Error(t, res.Err)
	msg := res.Stderr.String()
	assert.Contains(t, msg, "could not resolve host")
	assert.Contains(t, msg, "✗ DNS lookup")
	assert.Contains(t, msg, `Could not resolve "does-not-exist.invalid"`)
}

func TestConnectDiagnosis_JSONOutputHasNoANSI(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	require.NoError(t, ln.Close())

	h := NewCommandHarness(t)
	res := h.Execute(
		"workflow", "list",
		"--address", addr,
		"--client-connect-timeout", "2s",
		"-o", "json",
	)

	require.Error(t, res.Err)
	assert.NotContains(t, res.Stderr.String(), "\x1b[", "diagnosis must not contain ANSI escapes in JSON mode")
	assert.Empty(t, res.Stdout.String(), "connection failures must not write to stdout")
}

func TestConnectDiagnosis_Disabled(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	require.NoError(t, ln.Close())

	h := NewCommandHarness(t)
	h.Options.EnvLookup = EnvLookupMap{"TEMPORAL_CLI_DISABLE_CONNECT_DIAGNOSIS": "1"}
	res := h.Execute(
		"workflow", "list",
		"--address", addr,
		"--client-connect-timeout", "2s",
	)

	require.Error(t, res.Err)
	msg := res.Stderr.String()
	assert.Contains(t, msg, "failed connecting to Temporal server at "+addr)
	assert.NotContains(t, msg, "✗", "diagnosis must be suppressed when disabled")
}

func TestConnectDiagnosis_CertFileMissing(t *testing.T) {
	h := NewCommandHarness(t)
	res := h.Execute(
		"workflow", "list",
		"--address", "127.0.0.1:7233",
		"--tls-cert-path", "/definitely/does/not/exist.pem",
		"--tls-key-path", "/definitely/does/not/exist.key",
	)

	require.Error(t, res.Err)
	msg := res.Stderr.String()
	assert.Contains(t, msg, "cannot read file")
	assert.Contains(t, msg, "/definitely/does/not/exist")
	assert.NotContains(t, msg, "✓", "no probe stages expected when the dial never happened")
}

func TestConnectDiagnosis_ProfileAddressNamed(t *testing.T) {
	// When the failing address comes from a config profile, the suggestion
	// should say so and point at `temporal config get`.
	f, err := os.CreateTemp("", "")
	require.NoError(t, err)
	t.Cleanup(func() { os.Remove(f.Name()) })
	_, err = f.WriteString(`
[profile.default]
address = "does-not-exist.invalid:7233"
`)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	h := NewCommandHarness(t)
	h.Options.EnvLookup = EnvLookupMap{"TEMPORAL_CONFIG_FILE": f.Name()}
	res := h.Execute(
		"workflow", "list",
		"--client-connect-timeout", "5s",
	)

	require.Error(t, res.Err)
	msg := res.Stderr.String()
	assert.Contains(t, msg, "failed connecting to Temporal server at does-not-exist.invalid:7233")
	assert.Contains(t, msg, `The address comes from config profile "default"`)
	assert.Contains(t, msg, "temporal config get --prop address --profile default")
}

func TestConnectDiagnosis_CommandTimeoutCauseSurvives(t *testing.T) {
	ln := startBlackHoleListener(t)

	h := NewCommandHarness(t)
	res := h.Execute(
		"workflow", "list",
		"--address", ln.Addr().String(),
		"--command-timeout", "1s",
	)

	require.Error(t, res.Err)
	assert.Contains(t, res.Err.Error(), "command timed out after 1s")
}
