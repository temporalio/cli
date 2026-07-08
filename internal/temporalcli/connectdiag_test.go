package temporalcli

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/api/serviceerror"
)

// newTestCert generates a self-signed server certificate for 127.0.0.1.
func newTestCert(t *testing.T) tls.Certificate {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "connectdiag-test"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:              []string{"localhost"},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	require.NoError(t, err)
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key, Leaf: mustParseCert(t, der)}
}

func mustParseCert(t *testing.T, der []byte) *x509.Certificate {
	t.Helper()
	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)
	return cert
}

func certPool(t *testing.T, cert tls.Certificate) *x509.CertPool {
	t.Helper()
	pool := x509.NewCertPool()
	pool.AddCert(cert.Leaf)
	return pool
}

// serveTLS accepts connections and completes TLS handshakes until the
// listener closes. Returned address is host:port.
func serveTLS(t *testing.T, cfg *tls.Config) string {
	t.Helper()
	ln, err := tls.Listen("tcp", "127.0.0.1:0", cfg)
	require.NoError(t, err)
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				// Drive the handshake; keep the connection open briefly so
				// post-handshake alerts (TLS 1.3 client-cert enforcement)
				// reach the client, then close.
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
	return ln.Addr().String()
}

func testDiagnose(t *testing.T, address string, tlsCfg *tls.Config) *connectDiagnosis {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Mirror the SDK's eager GetSystemInfo failure shape.
	origErr := fmt.Errorf("failed reaching server: %w", serviceerror.NewDeadlineExceeded("context deadline exceeded"))
	return diagnoseConnection(ctx, address, tlsCfg, origErr)
}

func TestDiagnoseConnection_ClientCertRequired(t *testing.T) {
	cert := newTestCert(t)
	addr := serveTLS(t, &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool(t, cert),
	})

	// Client trusts the server CA but has no client cert.
	d := testDiagnose(t, addr, &tls.Config{RootCAs: certPool(t, cert)})
	// This test also pins the Go crypto/tls alert text ("certificate
	// required" / "bad certificate" / "handshake failure") that
	// classifyTLSAlert depends on; if it fails after a Go upgrade, revisit
	// classifyTLSAlert.
	assert.Equal(t, causeClientCertRequired, d.Cause)
	requireStage(t, d, diagOK, "TCP connection established")
	requireStage(t, d, diagFail, "server requires mTLS")
}

func TestDiagnoseConnection_ServerPlaintext(t *testing.T) {
	// Plain TCP listener that immediately writes a non-TLS response.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			_, _ = conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
			conn.Close()
		}
	}()

	d := testDiagnose(t, ln.Addr().String(), &tls.Config{InsecureSkipVerify: true})
	assert.Equal(t, causeServerPlaintext, d.Cause)
	requireStage(t, d, diagFail, "did not respond with TLS")
}

func TestDiagnoseConnection_ServerSpeaksTLS(t *testing.T) {
	cert := newTestCert(t)
	addr := serveTLS(t, &tls.Config{Certificates: []tls.Certificate{cert}})

	// No TLS configured on the client.
	d := testDiagnose(t, addr, nil)
	assert.Equal(t, causeServerSpeaksTLS, d.Cause)
	requireStage(t, d, diagFail, "server expects TLS")
}

func TestDiagnoseConnection_Refused(t *testing.T) {
	// Grab a port and close it so nothing is listening.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	require.NoError(t, ln.Close())

	d := testDiagnose(t, addr, nil)
	assert.Equal(t, causeTCPRefused, d.Cause)
	requireStage(t, d, diagFail, "TCP connection refused")
}

func TestDiagnoseConnection_DNSFailure(t *testing.T) {
	// .invalid is reserved (RFC 2606) and can never resolve.
	d := testDiagnose(t, "does-not-exist.invalid:7233", nil)
	assert.Equal(t, causeDNS, d.Cause)
	requireStage(t, d, diagFail, "DNS lookup")
}

func TestDiagnoseConnection_UnknownCA(t *testing.T) {
	cert := newTestCert(t)
	addr := serveTLS(t, &tls.Config{Certificates: []tls.Certificate{cert}})

	// Client does not trust the server's self-signed cert.
	d := testDiagnose(t, addr, &tls.Config{})
	assert.Equal(t, causeCAVerify, d.Cause)
	requireStage(t, d, diagFail, "cannot verify server certificate")
}

func TestDiagnoseConnection_HealthyTLSFallsThroughToGRPC(t *testing.T) {
	cert := newTestCert(t)
	addr := serveTLS(t, &tls.Config{Certificates: []tls.Certificate{cert}})

	// TLS is fine; the original (gRPC-level) error should be classified.
	d := testDiagnose(t, addr, &tls.Config{RootCAs: certPool(t, cert)})
	assert.Equal(t, causeTimeout, d.Cause)
	requireStage(t, d, diagOK, "TLS handshake succeeded")
	requireStage(t, d, diagFail, "gRPC connection failed")
}

func requireStage(t *testing.T, d *connectDiagnosis, status diagStatus, labelSubstr string) {
	t.Helper()
	for _, stage := range d.Stages {
		if stage.Status == status && strings.Contains(stage.Label, labelSubstr) {
			return
		}
	}
	t.Fatalf("no stage with status %v containing %q in %+v", status, labelSubstr, d.Stages)
}

func TestSuggestFix(t *testing.T) {
	cloudMeta := connectMeta{
		Args:    []string{"workflow", "list", "--address", "foo.bar.tmprl.cloud:7233"},
		Address: "foo.bar.tmprl.cloud:7233",
	}
	tests := []struct {
		name     string
		diag     *connectDiagnosis
		meta     connectMeta
		contains []string
	}{
		{
			name:     "mTLS on cloud endpoint",
			diag:     &connectDiagnosis{Cause: causeClientCertRequired},
			meta:     cloudMeta,
			contains: []string{"Temporal Cloud", "--tls-cert-path YourCert.pem", "--tls-key-path YourKey.pem", "temporal config set --prop tls.client_cert_path"},
		},
		{
			name:     "mTLS on generic endpoint",
			diag:     &connectDiagnosis{Cause: causeClientCertRequired},
			meta:     connectMeta{Args: []string{"workflow", "list"}, Address: "myhost:7233"},
			contains: []string{"requires client certificates", "--tls-cert-path YourCert.pem"},
		},
		{
			name:     "refused on local default port",
			diag:     &connectDiagnosis{Cause: causeTCPRefused},
			meta:     connectMeta{Address: "127.0.0.1:7233"},
			contains: []string{"temporal server start-dev"},
		},
		{
			name:     "refused elsewhere",
			diag:     &connectDiagnosis{Cause: causeTCPRefused},
			meta:     connectMeta{Address: "myhost:9999"},
			contains: []string{"Nothing is listening at myhost:9999"},
		},
		{
			name:     "dns failure from profile address",
			diag:     &connectDiagnosis{Cause: causeDNS},
			meta:     connectMeta{Address: "typo.example.com:7233", AddressSource: "profile", ProfileName: "prod"},
			contains: []string{`Could not resolve "typo.example.com"`, `profile "prod"`, "temporal config get --prop address"},
		},
		{
			name:     "server speaks TLS",
			diag:     &connectDiagnosis{Cause: causeServerSpeaksTLS},
			meta:     connectMeta{Args: []string{"workflow", "list"}, Address: "myhost:7233"},
			contains: []string{"Add --tls"},
		},
		{
			name:     "server plaintext",
			diag:     &connectDiagnosis{Cause: causeServerPlaintext},
			meta:     connectMeta{Address: "myhost:7233", TLSConfigured: true},
			contains: []string{"does not appear to use TLS"},
		},
		{
			name:     "api key rejected",
			diag:     &connectDiagnosis{Cause: causeAuth},
			meta:     connectMeta{Address: "us-west-2.aws.api.temporal.io:7233", HasAPIKey: true},
			contains: []string{"rejected the provided API key"},
		},
		{
			name:     "cert file unreadable",
			diag:     &connectDiagnosis{Cause: causeCertFileUnreadable, Detail: "/nope.pem"},
			meta:     connectMeta{Address: "myhost:7233"},
			contains: []string{`Cannot read "/nope.pem"`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := suggestFix(tt.diag, tt.meta)
			for _, want := range tt.contains {
				assert.Contains(t, got, want)
			}
		})
	}
}

func TestReconstructCommand(t *testing.T) {
	got := reconstructCommand(
		[]string{"workflow", "list", "--address", "foo:7233", "--query", "WorkflowType = 'x'"},
		"--tls-cert-path YourCert.pem",
	)
	assert.Equal(t,
		"temporal workflow list \\\n"+
			"    --address foo:7233 \\\n"+
			"    --query \"WorkflowType = 'x'\" \\\n"+
			"    --tls-cert-path YourCert.pem",
		got)
}

func TestConnectErrorRendering(t *testing.T) {
	origErr := fmt.Errorf("failed reaching server: context deadline exceeded")
	d := &connectDiagnosis{
		Address: "foo.bar.tmprl.cloud:7233",
		Cause:   causeClientCertRequired,
		Stages: []diagStage{
			{diagOK, "DNS resolved (3 addresses)"},
			{diagOK, "TCP connection established"},
			{diagFail, "TLS handshake failed: server requires mTLS, no valid client certificate was provided"},
		},
	}
	err := newConnectError(d, connectMeta{
		Args:    []string{"workflow", "list", "--address", "foo.bar.tmprl.cloud:7233"},
		Address: "foo.bar.tmprl.cloud:7233",
	}, origErr)

	msg := err.Error()
	assert.Contains(t, msg, "failed connecting to Temporal server at foo.bar.tmprl.cloud:7233: TLS handshake failed: server requires client certificate (mTLS)")
	assert.Contains(t, msg, "Connecting to foo.bar.tmprl.cloud:7233")
	assert.Contains(t, msg, "✓ DNS resolved (3 addresses)")
	assert.Contains(t, msg, "✗ TLS handshake failed")
	assert.Contains(t, msg, "--tls-cert-path YourCert.pem")
	// Unwrap preserves the original error.
	assert.ErrorContains(t, err.Unwrap(), "context deadline exceeded")
}
