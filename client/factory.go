package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gogo/status"
	"github.com/temporalio/cli/common"
	"github.com/temporalio/cli/dataconverter"
	"github.com/temporalio/cli/headersprovider"
	"github.com/urfave/cli/v2"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/api/workflowservice/v1"
	sdkclient "go.temporal.io/sdk/client"
	"go.temporal.io/server/common/auth"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/log/tag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
)

var (
	CFactory ClientFactory
)

var netClient HttpGetter = &http.Client{
	Timeout: time.Second * 10,
}

func GetSDKClient(c *cli.Context) (sdkclient.Client, error) {
	namespace, err := common.RequiredFlag(c, common.FlagNamespace)
	if err != nil {
		return nil, err
	}
	return CFactory.SDKClient(c, namespace), nil
}

// HttpGetter defines http.Client.Get(...) as an interface so we can mock it
type HttpGetter interface {
	Get(url string) (resp *http.Response, err error)
}

// ClientFactory is used to construct rpc clients
type ClientFactory interface {
	FrontendClient(c *cli.Context) workflowservice.WorkflowServiceClient
	OperatorClient(c *cli.Context) operatorservice.OperatorServiceClient
	SDKClient(c *cli.Context, namespace string) sdkclient.Client
	HealthClient(c *cli.Context) healthpb.HealthClient
}

func Init(c *cli.Context) {
	for _, c := range c.App.Commands {
		common.AddBeforeHandler(c, configureSDK)
	}
}

func configureSDK(ctx *cli.Context) error {
	endpoint := ctx.String(common.FlagCodecEndpoint)
	if endpoint != "" {
		dataconverter.SetRemoteEndpoint(
			endpoint,
			ctx.String(common.FlagNamespace),
			ctx.String(common.FlagCodecAuth),
		)
	}

	md, err := common.SplitKeyValuePairs(ctx.StringSlice(common.FlagMetadata))
	if err != nil {
		return err
	}

	headersprovider.SetGRPCHeadersProvider(md)

	return nil
}

type clientFactory struct {
	logger log.Logger
}

// NewClientFactory creates a new ClientFactory
func NewClientFactory() ClientFactory {
	logger := log.NewCLILogger()

	return &clientFactory{
		logger: logger,
	}
}

// FrontendClient builds a frontend client
func (b *clientFactory) FrontendClient(c *cli.Context) workflowservice.WorkflowServiceClient {
	connection, _ := b.createGRPCConnection(c)

	return workflowservice.NewWorkflowServiceClient(connection)
}

// FrontendClient builds an operator client
func (b *clientFactory) OperatorClient(c *cli.Context) operatorservice.OperatorServiceClient {
	connection, _ := b.createGRPCConnection(c)

	return operatorservice.NewOperatorServiceClient(connection)
}

// SDKClient builds an SDK client.
func (b *clientFactory) SDKClient(c *cli.Context, namespace string) sdkclient.Client {
	hostPort := c.String(common.FlagAddress)
	if hostPort == "" {
		hostPort = common.LocalHostPort
	}

	tlsConfig, err := b.createTLSConfig(c)
	if err != nil {
		b.logger.Fatal("Failed to configure TLS for SDK client", tag.Error(err))
	}

	sdkClient, err := sdkclient.Dial(sdkclient.Options{
		HostPort:  hostPort,
		Namespace: namespace,
		Logger:    log.NewSdkLogger(b.logger),
		Identity:  common.GetCliIdentity(),
		ConnectionOptions: sdkclient.ConnectionOptions{
			TLS: tlsConfig,
		},
		HeadersProvider: headersprovider.GetCurrent(),
	})
	if err != nil {
		b.logger.Fatal("Failed to create SDK client", tag.Error(err))
	}

	return sdkClient
}

// HealthClient builds a health client.
func (b *clientFactory) HealthClient(c *cli.Context) healthpb.HealthClient {
	connection, _ := b.createGRPCConnection(c)

	return healthpb.NewHealthClient(connection)
}

func headersProviderInterceptor(headersProvider headersprovider.HeadersProvider) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		headers, err := headersProvider.GetHeaders(ctx)
		if err != nil {
			return err
		}
		for k, v := range headers {
			ctx = metadata.AppendToOutgoingContext(ctx, k, v)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func errorInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		err := invoker(ctx, method, req, reply, cc, opts...)
		err = serviceerror.FromStatus(status.Convert(err))
		return err
	}
}

func (b *clientFactory) createGRPCConnection(c *cli.Context) (*grpc.ClientConn, error) {
	hostPort := c.String(common.FlagAddress)
	if hostPort == "" {
		hostPort = common.LocalHostPort
	}

	tlsConfig, err := b.createTLSConfig(c)
	if err != nil {
		return nil, err
	}

	grpcSecurityOptions := grpc.WithTransportCredentials(insecure.NewCredentials())

	if tlsConfig != nil {
		grpcSecurityOptions = grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
	}

	interceptors := []grpc.UnaryClientInterceptor{
		errorInterceptor(),
		headersProviderInterceptor(headersprovider.GetCurrent()),
	}

	dialOpts := []grpc.DialOption{
		grpcSecurityOptions,
		grpc.WithChainUnaryInterceptor(interceptors...),
	}

	connection, err := grpc.Dial(hostPort, dialOpts...)
	if err != nil {
		b.logger.Fatal("Failed to create connection", tag.Error(err))
		return nil, err
	}
	return connection, nil
}

func (b *clientFactory) createTLSConfig(c *cli.Context) (*tls.Config, error) {
	certPath := c.String(common.FlagTLSCertPath)
	keyPath := c.String(common.FlagTLSKeyPath)
	caPath := c.String(common.FlagTLSCaPath)
	disableHostNameVerificationS := c.String(common.FlagTLSDisableHostVerification)
	disableHostNameVerification, err := strconv.ParseBool(disableHostNameVerificationS)
	if err != nil {
		return nil, fmt.Errorf("unable to read TLS disable host verification flag: %w", err)
	}
	enableTLSS := c.String(common.FlagTLS)
	enableTLS, err := strconv.ParseBool(enableTLSS)
	if err != nil {
		return nil, fmt.Errorf("unable to read TLS flag: %w", err)
	}

	serverName := c.String(common.FlagTLSServerName)

	var host string
	var cert *tls.Certificate
	var caPool *x509.CertPool

	if caPath != "" {
		caCertPool, err := fetchCACert(caPath)
		if err != nil {
			b.logger.Fatal("Failed to load server CA certificate", tag.Error(err))
			return nil, err
		}
		caPool = caCertPool
	}
	if certPath != "" {
		myCert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			b.logger.Fatal("Failed to load client certificate", tag.Error(err))
			return nil, err
		}
		cert = &myCert
	}
	// If we are given arguments to verify either server or client, configure TLS
	if caPool != nil || cert != nil {
		if serverName != "" {
			host = serverName
		} else {
			hostPort := c.String(common.FlagAddress)
			if hostPort == "" {
				hostPort = common.LocalHostPort
			}
			// Ignoring error as we'll fail to dial anyway, and that will produce a meaningful error
			host, _, _ = net.SplitHostPort(hostPort)
		}
		tlsConfig := auth.NewTLSConfigForServer(host, !disableHostNameVerification)
		if caPool != nil {
			tlsConfig.RootCAs = caPool
		}
		if cert != nil {
			tlsConfig.Certificates = []tls.Certificate{*cert}
		}

		return tlsConfig, nil
	}
	// If we are given a server name, set the TLS server name for DNS resolution
	if serverName != "" {
		host = serverName
		tlsConfig := auth.NewTLSConfigForServer(host, !disableHostNameVerification)
		return tlsConfig, nil
	}
	// If we are given a TLS flag, set the TLS server name from the address
	if enableTLS {
		hostPort := c.String(common.FlagAddress)
		if hostPort == "" {
			hostPort = common.LocalHostPort
		}
		// Ignoring error as we'll fail to dial anyway, and that will produce a meaningful error
		host, _, _ = net.SplitHostPort(hostPort)
		tlsConfig := auth.NewTLSConfigForServer(host, !disableHostNameVerification)
		return tlsConfig, nil
	}

	return nil, nil
}

func fetchCACert(pathOrUrl string) (caPool *x509.CertPool, err error) {
	caPool = x509.NewCertPool()
	var caBytes []byte

	if strings.HasPrefix(pathOrUrl, "http://") {
		return nil, errors.New("HTTP is not supported for CA cert URLs. Provide HTTPS URL")
	}

	if strings.HasPrefix(pathOrUrl, "https://") {
		resp, err := netClient.Get(pathOrUrl)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		caBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
	} else {
		caBytes, err = os.ReadFile(pathOrUrl)
		if err != nil {
			return nil, err
		}
	}

	if !caPool.AppendCertsFromPEM(caBytes) {
		return nil, errors.New("unknown failure constructing cert pool for ca")
	}
	return caPool, nil
}
