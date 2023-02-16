package headers

import (
	"context"
	"runtime/debug"

	"google.golang.org/grpc/metadata"
)

const (
	ClientNameHeaderName              = "client-name"
	ClientVersionHeaderName           = "client-version"
	SupportedServerVersionsHeaderName = "supported-server-versions"
	SupportedFeaturesHeaderName       = "supported-features"
	SupportedFeaturesHeaderDelim      = ","
)

// Set by GoReleaser using ldflags
var Version = "DEV"

const (
	ClientNameCLI = "temporal-cli"

	// SupportedServerVersions is used by CLI and inter role communication.
	SupportedServerVersions = ">=1.0.0 <2.0.0"
)

var (
	// propagateHeaders are the headers to propagate from the frontend to other services.
	propagateHeaders = []string{
		ClientNameHeaderName,
		ClientVersionHeaderName,
		SupportedServerVersionsHeaderName,
		SupportedFeaturesHeaderName,
	}

	internalVersionHeaders = metadata.New(map[string]string{
		SupportedServerVersionsHeaderName: SupportedServerVersions,
	})

	cliVersionHeaders = metadata.New(map[string]string{
		ClientNameHeaderName:              ClientNameCLI,
		ClientVersionHeaderName:           Version,
		SupportedServerVersionsHeaderName: SupportedServerVersions,
		// TODO: This should include SupportedFeaturesHeaderName with a value that's taken
		// from the Go SDK (since the cli uses the Go SDK for most operations).
	})
)

func Init() {
	if Version == "DEV" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" {
			Version = info.Main.Version
		}
	}
}

// GetValues returns header values for passed header names.
// It always returns slice of the same size as number of passed header names.
func GetValues(ctx context.Context, headerNames ...string) []string {
	headerValues := make([]string, len(headerNames))

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for i, headerName := range headerNames {
			headerValues[i] = getSingleHeaderValue(md, headerName)
		}
	}

	return headerValues
}

// Propagate propagates version headers from incoming context to outgoing context.
// It copies all version headers to outgoing context only if they are exist in incoming context
// and doesn't exist in outgoing context already.
func Propagate(ctx context.Context) context.Context {
	if mdIncoming, ok := metadata.FromIncomingContext(ctx); ok {
		var headersToAppend []string
		mdOutgoing, mdOutgoingExist := metadata.FromOutgoingContext(ctx)
		for _, headerName := range propagateHeaders {
			if incomingValue := mdIncoming.Get(headerName); len(incomingValue) > 0 {
				if mdOutgoingExist {
					if outgoingValue := mdOutgoing.Get(headerName); len(outgoingValue) > 0 {
						continue
					}
				}
				headersToAppend = append(headersToAppend, headerName, incomingValue[0])
			}
		}
		if headersToAppend != nil {
			if mdOutgoingExist {
				ctx = metadata.AppendToOutgoingContext(ctx, headersToAppend...)
			} else {
				ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(headersToAppend...))
			}
		}
	}
	return ctx
}

// SetVersions sets headers for internal communications.
func SetVersions(ctx context.Context) context.Context {
	return metadata.NewOutgoingContext(ctx, internalVersionHeaders)
}

// SetCLIVersions sets headers for CLI requests.
func SetCLIVersions(ctx context.Context) context.Context {
	return metadata.NewOutgoingContext(ctx, cliVersionHeaders)
}

// SetVersionsForTests sets headers as they would be received from the client.
// Must be used in tests only.
func SetVersionsForTests(ctx context.Context, clientVersion, clientName, supportedServerVersions, supportedFeatures string) context.Context {
	return metadata.NewIncomingContext(ctx, metadata.New(map[string]string{
		ClientNameHeaderName:              clientName,
		ClientVersionHeaderName:           clientVersion,
		SupportedServerVersionsHeaderName: supportedServerVersions,
		SupportedFeaturesHeaderName:       supportedFeatures,
	}))
}

func getSingleHeaderValue(md metadata.MD, headerName string) string {
	values := md.Get(headerName)
	if len(values) == 0 {
		return ""
	}

	return values[0]
}
