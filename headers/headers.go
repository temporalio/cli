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
var Version = "0.0.0-DEV"

const (
	ClientNameCLI = "temporal-cli"

	// SupportedServerVersions is used by CLI and inter role communication.
	SupportedServerVersions = ">=1.0.0 <2.0.0"
)

var (
	cliVersionHeaders = metadata.New(map[string]string{
		ClientNameHeaderName:              ClientNameCLI,
		ClientVersionHeaderName:           Version,
		SupportedServerVersionsHeaderName: SupportedServerVersions,
		// TODO: This should include SupportedFeaturesHeaderName with a value that's taken
		// from the Go SDK (since the cli uses the Go SDK for most operations).
	})
)

func Init() {
	if Version == "0.0.0-DEV" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" {
			Version = info.Main.Version
		}
	}
}

// SetCLIVersions sets headers for CLI requests.
func SetCLIVersions(ctx context.Context) context.Context {
	return metadata.NewOutgoingContext(ctx, cliVersionHeaders)
}
