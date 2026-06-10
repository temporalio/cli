package temporalcli_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/temporalio/cli/internal/systemnexusgen"
)

// TestSystemNexusGenUpToDate fails if commands.system_nexus.gen.go is stale,
// e.g. after a go.temporal.io/api bump adds or changes a system Nexus operation.
// Run 'make gen' to regenerate.
func TestSystemNexusGenUpToDate(t *testing.T) {
	ops, err := systemnexusgen.Parse("go.temporal.io/api/workflowservice/v1/workflowservicenexus")
	require.NoError(t, err)

	want, err := systemnexusgen.GenerateCode("temporalcli", ops)
	require.NoError(t, err)

	got, err := os.ReadFile("commands.system_nexus.gen.go")
	require.NoError(t, err)

	require.Equal(t, string(want), string(got),
		"commands.system_nexus.gen.go is out of date; run 'make gen'")
}
