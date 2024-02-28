package temporalcli_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/cli/temporalcli/devserver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"os"
	"strings"
	"testing"
)

func Test_SetsAuthHeadersFromEnvVar(t *testing.T) {
	foundEnchi := false
	grpcIceptor := func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		if strings.Contains(info.FullMethod, "ListWorkflowExecutions") {
			md, ok := metadata.FromIncomingContext(ctx)
			if ok {
				values := md.Get("authorization")
				for _, v := range values {
					if v == "Bearer enchi" {
						foundEnchi = true
					}
				}
			}
		}
		return handler(ctx, req)
	}
	devServer := StartDevServer(t, DevServerOptions{
		StartOptions: devserver.StartOptions{
			GRPCInterceptors: []grpc.UnaryServerInterceptor{grpcIceptor},
		},
	})
	defer devServer.Stop()

	cmdh := NewCommandHarness(t)
	err := os.Setenv(`TEMPORAL_CLI_AUTHORIZATION_TOKEN`, "Bearer enchi")
	require.NoError(t, err)
	defer os.Unsetenv(`TEMPORAL_CLI_AUTHORIZATION_TOKEN`)
	res := cmdh.Execute("workflow", "list", "--address", devServer.Address())
	require.NoError(t, res.Err)
	require.True(t, foundEnchi)
}
