package systemnexusgen_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/temporalio/cli/internal/systemnexusgen"
)

func TestParseDiscoversWorkflowService(t *testing.T) {
	ops, err := systemnexusgen.Parse("go.temporal.io/api/workflowservice/v1/workflowservicenexus")
	require.NoError(t, err)

	var got *systemnexusgen.Operation
	for i := range ops {
		if ops[i].OpFieldName == "SignalWithStartWorkflowExecution" {
			got = &ops[i]
			break
		}
	}
	require.NotNil(t, got, "expected SignalWithStartWorkflowExecution operation")

	require.Equal(t, "go.temporal.io/api/workflowservice/v1/workflowservicenexus", got.NexusPkgPath)
	require.Equal(t, "workflowservicenexus", got.NexusPkgName)
	require.Equal(t, "TemporalAPIWorkflowserviceV1WorkflowService", got.ServiceVarName)
	require.Equal(t, systemnexusgen.ProtoType{
		PkgPath: "go.temporal.io/api/workflowservice/v1",
		PkgName: "workflowservice",
		Name:    "SignalWithStartWorkflowExecutionRequest",
	}, got.Request)
	require.Equal(t, systemnexusgen.ProtoType{
		PkgPath: "go.temporal.io/api/workflowservice/v1",
		PkgName: "workflowservice",
		Name:    "SignalWithStartWorkflowExecutionResponse",
	}, got.Response)
}
