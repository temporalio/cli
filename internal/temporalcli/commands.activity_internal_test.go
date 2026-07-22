package temporalcli

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/api/serviceerror"
)

func TestActivityNotFoundErrorPreservesCauseAndHasConservativeReport(t *testing.T) {
	cause, ok := serviceerror.NewNotFound("server detail").(*serviceerror.NotFound)
	require.True(t, ok)
	wrappedCause := fmt.Errorf("poll failed: %w", cause)
	err := &activityNotFoundError{activityID: "activity-id", cause: wrappedCause}

	var notFound *serviceerror.NotFound
	assert.ErrorAs(t, err, &notFound)
	assert.Same(t, cause, notFound)
	assert.Equal(t, "activity not found: activity-id", err.Error())

	report := err.report()
	assert.Equal(t, "standalone Activity not found", report.Summary)
	assert.Empty(t, report.Context)
	assert.Empty(t, report.Checks)
	assert.Nil(t, report.Action)
	assert.True(t, errors.Is(err, wrappedCause))
	assert.True(t, errors.Is(err, cause))
}
