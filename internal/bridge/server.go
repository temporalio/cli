package bridge

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/temporalio/cli/internal/devserver"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/server/api/adminservice/v1"
	"go.temporal.io/server/service/localexecution"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	loopbackAddress       = "127.0.0.1"
	bridgeStartupTimeout  = 30 * time.Second
	bridgeShutdownTimeout = 10 * time.Second
)

type StartOptions struct {
	Configuration localexecution.BridgeConfiguration
	StateStore    *localexecution.BridgeStateStore
	FrontendPort  int
	Logger        *slog.Logger
}

type Server struct {
	cancel             context.CancelFunc
	done               <-chan error
	local              *devserver.Server
	upstreamClient     client.Client
	upstreamConnection *grpc.ClientConn
	localConnection    *grpc.ClientConn
	frontendAddress    string

	stopOnce sync.Once
}

func Start(ctx context.Context, options StartOptions) (_ *Server, retError error) {
	if err := options.Configuration.Validate(); err != nil {
		return nil, fmt.Errorf("validate bridge configuration: %w", err)
	}
	if options.StateStore == nil {
		return nil, errors.New("bridge state store is required")
	}
	if options.Logger == nil {
		return nil, errors.New("bridge logger is required")
	}
	if options.FrontendPort < 0 || options.FrontendPort > 65535 {
		return nil, errors.New("bridge frontend port must be between 0 and 65535")
	}
	frontendPort := options.FrontendPort
	if frontendPort == 0 {
		frontendPort = devserver.MustGetFreePort(loopbackAddress)
	}
	metricsPort := devserver.MustGetFreePort(loopbackAddress)

	startupContext, cancelStartup := context.WithTimeout(ctx, bridgeStartupTimeout)
	defer cancelStartup()
	upstreamClient, upstreamConnection, upstreamNamespaceID, err := connectUpstream(
		startupContext,
		options.Configuration,
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		if retError != nil {
			upstreamClient.Close()
			_ = upstreamConnection.Close()
		}
	}()

	local, err := devserver.Start(devserver.StartOptions{
		FrontendIP:             loopbackAddress,
		FrontendPort:           frontendPort,
		Namespaces:             []string{options.Configuration.Namespace},
		ClusterID:              uuid.NewString(),
		MasterClusterName:      "active",
		CurrentClusterName:     "active",
		InitialFailoverVersion: 1,
		Logger:                 options.Logger,
		LogLevel:               slog.LevelWarn,
		DatabaseFile:           options.StateStore.DatabasePath(),
		MetricsPort:            metricsPort,
		EnableGlobalNamespace:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("start local Temporal server: %w", err)
	}
	defer func() {
		if retError != nil {
			local.Stop()
		}
	}()

	frontendAddress := fmt.Sprintf("%s:%d", loopbackAddress, frontendPort)
	localConnection, localNamespaceID, err := connectLocal(
		startupContext,
		frontendAddress,
		options.Configuration.Namespace,
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		if retError != nil {
			_ = localConnection.Close()
		}
	}()

	runtime, err := localexecution.NewBridgeRuntime(localexecution.BridgeRuntimeOptions{
		Configuration:       options.Configuration,
		StateStore:          options.StateStore,
		UpstreamNamespaceID: upstreamNamespaceID,
		LocalNamespaceID:    localNamespaceID,
		UpstreamWorkflow:    upstreamClient.WorkflowService(),
		UpstreamAdmin:       adminservice.NewAdminServiceClient(upstreamConnection),
		LocalAdmin:          adminservice.NewAdminServiceClient(localConnection),
	})
	if err != nil {
		return nil, fmt.Errorf("configure bridge runtime: %w", err)
	}
	runtimeContext, cancelRuntime := context.WithCancel(ctx)
	done, err := runtime.Start(runtimeContext)
	if err != nil {
		cancelRuntime()
		return nil, fmt.Errorf("start bridge runtime: %w", err)
	}
	return &Server{
		cancel:             cancelRuntime,
		done:               done,
		local:              local,
		upstreamClient:     upstreamClient,
		upstreamConnection: upstreamConnection,
		localConnection:    localConnection,
		frontendAddress:    frontendAddress,
	}, nil
}

func (s *Server) FrontendAddress() string {
	return s.frontendAddress
}

func (s *Server) Done() <-chan error {
	return s.done
}

func (s *Server) Stop() {
	s.stopOnce.Do(func() {
		s.cancel()
		shutdownTimer := time.NewTimer(bridgeShutdownTimeout)
		select {
		case <-s.done:
			if !shutdownTimer.Stop() {
				<-shutdownTimer.C
			}
		case <-shutdownTimer.C:
		}
		_ = s.localConnection.Close()
		s.upstreamClient.Close()
		_ = s.upstreamConnection.Close()
		s.local.Stop()
	})
}

func connectUpstream(
	ctx context.Context,
	configuration localexecution.BridgeConfiguration,
) (client.Client, *grpc.ClientConn, string, error) {
	tlsConfig, err := buildTLSConfig(configuration.Upstream)
	if err != nil {
		return nil, nil, "", err
	}
	clientOptions := client.Options{
		HostPort:        configuration.Upstream.Address,
		Namespace:       configuration.Namespace,
		HeadersProvider: staticHeaders(configuration.Upstream.Headers),
	}
	if configuration.Upstream.APIKey != "" {
		clientOptions.Credentials = client.NewAPIKeyStaticCredentials(configuration.Upstream.APIKey)
	}
	if tlsConfig != nil {
		clientOptions.ConnectionOptions.TLS = tlsConfig.Clone()
	}
	upstreamClient, err := client.DialContext(ctx, clientOptions)
	if err != nil {
		return nil, nil, "", fmt.Errorf("connect upstream Temporal client: %w", err)
	}

	connection, err := grpc.NewClient(
		configuration.Upstream.Address,
		grpc.WithTransportCredentials(transportCredentials(tlsConfig)),
		grpc.WithChainUnaryInterceptor(upstreamMetadataInterceptor(configuration.Upstream)),
	)
	if err != nil {
		upstreamClient.Close()
		return nil, nil, "", fmt.Errorf("connect upstream admin client: %w", err)
	}
	namespaceID, err := describeNamespaceID(ctx, connection, configuration.Namespace)
	if err != nil {
		upstreamClient.Close()
		_ = connection.Close()
		return nil, nil, "", fmt.Errorf("describe upstream namespace: %w", err)
	}
	return upstreamClient, connection, namespaceID, nil
}

func connectLocal(
	ctx context.Context,
	address string,
	namespace string,
) (*grpc.ClientConn, string, error) {
	connection, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, "", fmt.Errorf("connect local admin client: %w", err)
	}
	namespaceID, err := describeNamespaceID(ctx, connection, namespace)
	if err != nil {
		_ = connection.Close()
		return nil, "", fmt.Errorf("describe local namespace: %w", err)
	}
	return connection, namespaceID, nil
}

func describeNamespaceID(ctx context.Context, connection *grpc.ClientConn, namespace string) (string, error) {
	response, err := workflowservice.NewWorkflowServiceClient(connection).DescribeNamespace(
		ctx,
		&workflowservice.DescribeNamespaceRequest{Namespace: namespace},
	)
	if err != nil {
		return "", err
	}
	if response.GetNamespaceInfo().GetId() == "" {
		return "", errors.New("namespace has no ID")
	}
	return response.GetNamespaceInfo().GetId(), nil
}

func buildTLSConfig(profile localexecution.UpstreamConnectionProfile) (*tls.Config, error) {
	if profile.TLS == nil && profile.APIKey == "" {
		return nil, nil
	}
	config := &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: profile.ServerName,
	}
	if profile.TLS == nil {
		return config, nil
	}
	if profile.TLS.ServerRootCACertificate != "" {
		roots, err := x509.SystemCertPool()
		if err != nil {
			roots = x509.NewCertPool()
		}
		if !roots.AppendCertsFromPEM([]byte(profile.TLS.ServerRootCACertificate)) {
			return nil, errors.New("upstream TLS root contains no certificates")
		}
		config.RootCAs = roots
	}
	if profile.TLS.ClientCertificate != "" {
		certificate, err := tls.X509KeyPair(
			[]byte(profile.TLS.ClientCertificate),
			[]byte(profile.TLS.ClientPrivateKey),
		)
		if err != nil {
			return nil, fmt.Errorf("load upstream TLS client identity: %w", err)
		}
		config.Certificates = []tls.Certificate{certificate}
	}
	return config, nil
}

func transportCredentials(config *tls.Config) credentials.TransportCredentials {
	if config == nil {
		return insecure.NewCredentials()
	}
	return credentials.NewTLS(config.Clone())
}

func upstreamMetadataInterceptor(
	profile localexecution.UpstreamConnectionProfile,
) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		request any,
		reply any,
		connection *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		outgoing := metadata.New(profile.Headers)
		if profile.APIKey != "" {
			outgoing.Set("authorization", "Bearer "+profile.APIKey)
		}
		ctx = metadata.NewOutgoingContext(ctx, outgoing)
		return invoker(ctx, method, request, reply, connection, opts...)
	}
}

type staticHeaders map[string]string

func (h staticHeaders) GetHeaders(context.Context) (map[string]string, error) {
	copy := make(map[string]string, len(h))
	for name, value := range h {
		copy[strings.ToLower(name)] = value
	}
	return copy, nil
}
