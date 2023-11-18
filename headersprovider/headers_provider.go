package headersprovider

import (
	"context"
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

func (a grpcHeaderProvider) GetHeaders(ctx context.Context) (map[string]string, error) {
	return a.headers, nil
}

func SetGRPCHeadersProvider(headers map[string]string) {
	headersProvider = &grpcHeaderProvider{headers}
}

func SetCurrent(hp HeadersProvider) {
	headersProvider = hp
}

func GetCurrent() HeadersProvider {
	return headersProvider
}
