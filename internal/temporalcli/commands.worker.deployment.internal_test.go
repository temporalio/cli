package temporalcli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScalerTypeForProvider(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		expected string
	}{
		{"aws-lambda is invoke-based -> no-sync", "aws-lambda", "no-sync"},
		{"gcp-cloud-run is worker-set-based -> rate-based", "gcp-cloud-run", "rate-based"},
		{"unknown provider falls back to no-sync", "azure-container-apps", "no-sync"},
		{"empty provider falls back to no-sync", "", "no-sync"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, scalerTypeForProvider(tt.provider))
		})
	}
}

// Every provider type computeProviderConfig can emit must have an explicit
// scaler mapping; a missing entry would silently fall back to no-sync and be
// rejected by WCI for worker-set providers.
func TestScalerTypeByProviderCoversAllProviders(t *testing.T) {
	for _, providerType := range []string{"aws-lambda", "gcp-cloud-run"} {
		_, ok := scalerTypeByProvider[providerType]
		require.Truef(t, ok, "provider %q has no scaler mapping", providerType)
	}
}
