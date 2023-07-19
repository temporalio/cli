package headersprovider

import (
	"context"

	h "github.com/temporalio/cli/headers"
)

type HeadersProvider interface {
	GetHeaders(context.Context) (map[string]string, error)
}

var (
	headersProvider HeadersProvider = nil
)

type grpcHeaderProvider struct {
	headers map[string]string
}

func newGrpcHeaderProvider(headers map[string]string) *grpcHeaderProvider {
	provider := &grpcHeaderProvider{headers}
	provider.headers[h.CallerTypeHeaderName] = h.CallerTypeHeaderCLI
	return provider
}

func (a grpcHeaderProvider) GetHeaders(ctx context.Context) (map[string]string, error) {
	return a.headers, nil
}

func SetGRPCHeadersProvider(headers map[string]string) {
	headersProvider = newGrpcHeaderProvider(headers)
}

func SetCurrent(hp HeadersProvider) {
	headersProvider = hp
}

func GetCurrent() HeadersProvider {
	return headersProvider
}
