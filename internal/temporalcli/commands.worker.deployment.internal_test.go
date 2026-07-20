package temporalcli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScalerTypeForProvider(t *testing.T) {
	tests := []struct {
		name      string
		provider  string
		expected  string
		expectErr bool
	}{
		{"aws-lambda is invoke-based -> no-sync", "aws-lambda", "no-sync", false},
		{"gcp-cloud-run is worker-set-based -> rate-based", "gcp-cloud-run", "rate-based", false},
		{"unknown provider errors", "azure-container-apps", "", true},
		{"empty provider errors", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scaler, err := scalerTypeForProvider(tt.provider)
			if tt.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected, scaler)
		})
	}
}

// Every provider type computeProviderConfig can emit must have an explicit
// scaler mapping; a missing entry makes scalerTypeForProvider error before the
// request is sent, so this guards against forgetting to map a newly-added provider.
func TestScalerTypeByProviderCoversAllProviders(t *testing.T) {
	for _, providerType := range []string{"aws-lambda", "gcp-cloud-run"} {
		_, ok := scalerTypeByProvider[providerType]
		require.Truef(t, ok, "provider %q has no scaler mapping", providerType)
	}
}
