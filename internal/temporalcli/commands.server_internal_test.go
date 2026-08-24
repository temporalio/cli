package temporalcli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/server/common/dynamicconfig"
)

func TestLoadDynamicConfigFile(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "dynamicconfig.yaml")
	require.NoError(t, os.WriteFile(filename, []byte(`
history.enableCHASMSchedulerCreation:
  - value: true
    constraints:
      namespace: fx-test
`), 0o600))

	values, err := loadDynamicConfigFile(filename)
	require.NoError(t, err)

	constrainedValues, ok := values["history.enablechasmschedulercreation"].([]dynamicconfig.ConstrainedValue)
	require.True(t, ok)
	require.Len(t, constrainedValues, 1)
	require.Equal(t, true, constrainedValues[0].Value)
	require.Equal(t, "fx-test", constrainedValues[0].Constraints.Namespace)
}
