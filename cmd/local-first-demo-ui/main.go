package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	uiserver "github.com/temporalio/ui-server/v2/server"
	uiconfig "github.com/temporalio/ui-server/v2/server/config"
	uiserveroptions "github.com/temporalio/ui-server/v2/server/server_options"
)

func main() {
	var temporalAddress string
	var publicPath string
	flag.StringVar(&temporalAddress, "temporal-address", "", "Temporal frontend address")
	flag.StringVar(&publicPath, "public-path", "", "URL prefix used by the reverse proxy")
	flag.Parse()
	if temporalAddress == "" || publicPath == "" {
		fmt.Fprintln(os.Stderr, "--temporal-address and --public-path are required")
		os.Exit(2)
	}

	port, err := availablePort()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	server := uiserver.NewServer(uiserveroptions.WithConfigProvider(&uiconfig.Config{
		Host:                  "127.0.0.1",
		Port:                  port,
		TemporalGRPCAddress:   temporalAddress,
		PublicPath:            publicPath,
		EnableUI:              true,
		DefaultNamespace:      "local-first-demo",
		DisableNewsFetch:      true,
		DisableWriteActions:   true,
		NavCollapsedByDefault: true,
		HideLogs:              true,
		CORS:                  uiconfig.CORS{CookieInsecure: true},
	}))
	errCh := make(chan error, 1)
	go func() { errCh <- server.Start() }()
	address := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	if err := waitUntilReady(address, publicPath); err != nil {
		server.Stop()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := json.NewEncoder(os.Stdout).Encode(map[string]string{"address": address}); err != nil {
		server.Stop()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	select {
	case <-ctx.Done():
		server.Stop()
	case err := <-errCh:
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func availablePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func waitUntilReady(address string, publicPath string) error {
	client := &http.Client{Timeout: time.Second}
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		response, err := client.Get("http://" + address + publicPath + "/healthz")
		if err == nil {
			response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("Temporal UI at %s did not become ready", address)
}
