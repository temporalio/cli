package app_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"text/template"
	"time"

	"github.com/temporalio/cli/app"
	sconfig "github.com/temporalio/cli/server/config"
	"github.com/urfave/cli/v2"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
)

func TestMTLSConfig(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, thisFile, _, _ := runtime.Caller(0)
	mtlsDir := filepath.Join(thisFile, "../testdata/mtls")

	// Create temp config dir
	confDir := t.TempDir()

	// Run templated config and put in temp dir
	var buf bytes.Buffer
	tmpl, err := template.New("temporal.yaml.template").
		Funcs(template.FuncMap{"qualified": func(s string) string { return strconv.Quote(filepath.Join(mtlsDir, s)) }}).
		ParseFiles(filepath.Join(mtlsDir, "temporal.yaml.template"))
	if err != nil {
		t.Fatal(err)
	} else if err = tmpl.Execute(&buf, nil); err != nil {
		t.Fatal(err)
	} else if err = os.WriteFile(filepath.Join(confDir, "temporal.yaml"), buf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	tmpl, err = template.New("temporal-ui.yaml.template").
		Funcs(template.FuncMap{"qualified": func(s string) string { return strconv.Quote(filepath.Join(mtlsDir, s)) }}).
		ParseFiles(filepath.Join(mtlsDir, "temporal-ui.yaml.template"))
	if err != nil {
		t.Fatal(err)
	} else if err = tmpl.Execute(&buf, nil); err != nil {
		t.Fatal(err)
	} else if err = os.WriteFile(filepath.Join(confDir, "temporal-ui.yaml"), buf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}

	portProvider := sconfig.NewPortProvider()
	var (
		frontendPort = portProvider.MustGetFreePort()
		webUIPort    = portProvider.MustGetFreePort()
	)
	portProvider.Close()

	// Run in-memory using temp config
	args := []string{
		"temporal",
		"server",
		"start-dev",
		"--config", confDir,
		"--namespace", "default",
		"--log-format", "noop",
		"--port", strconv.Itoa(frontendPort),
		"--ui-port", strconv.Itoa(webUIPort),
	}
	go func() {
		temporalCLI := app.BuildApp()
		// Don't call os.Exit
		temporalCLI.ExitErrHandler = func(_ *cli.Context, _ error) {}

		if err := temporalCLI.RunContext(ctx, args); err != nil {
			fmt.Printf("CLI failed: %s\n", err)
		}
	}()

	// Load client cert/key for auth
	clientCert, err := tls.LoadX509KeyPair(
		filepath.Join(mtlsDir, "client-cert.pem"),
		filepath.Join(mtlsDir, "client-key.pem"),
	)
	if err != nil {
		t.Fatal(err)
	}
	// Load server cert for CA check
	serverCAPEM, err := os.ReadFile(filepath.Join(mtlsDir, "server-ca-cert.pem"))
	if err != nil {
		t.Fatal(err)
	}
	serverCAPool := x509.NewCertPool()
	serverCAPool.AppendCertsFromPEM(serverCAPEM)

	// Build client options and try to connect client every 100ms for 5s
	options := client.Options{
		HostPort: fmt.Sprintf("localhost:%d", frontendPort),
		ConnectionOptions: client.ConnectionOptions{
			TLS: &tls.Config{
				Certificates: []tls.Certificate{clientCert},
				RootCAs:      serverCAPool,
			},
		},
	}
	var c client.Client
	for i := 0; i < 50; i++ {
		if c, err = client.Dial(options); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		t.Fatal(err)
	}

	// Make a call
	resp, err := c.WorkflowService().DescribeNamespace(ctx, &workflowservice.DescribeNamespaceRequest{
		Namespace: "default",
	})
	if err != nil {
		t.Fatal(err)
	} else if resp.NamespaceInfo.State != enums.NAMESPACE_STATE_REGISTERED {
		t.Fatalf("Bad state: %v", resp.NamespaceInfo.State)
	}

	// Pretend to be a browser to invoke the UI API
	res, err := http.Get(fmt.Sprintf("http://localhost:%d/api/v1/namespaces?", webUIPort))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected response %s, with body %s", res.Status, string(body))
	}
}
