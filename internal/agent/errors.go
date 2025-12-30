package agent

import (
	"regexp"
	"strings"
)

// CompactError extracts the core error message from a verbose Temporal error chain.
// It strips wrapper context like workflow IDs, run IDs, event IDs, and type annotations.
//
// Example input:
//
//	"WorkflowFailed: child workflow at depth 0 failed: child workflow execution error
//	 (type: NestedFailureWorkflow, workflowID: nested-level-1, runID: abc123,
//	 initiatedEventID: 5, startedEventID: 6): activity error (type: FailingActivity,
//	 scheduledEventID: 5, startedEventID: 6, identity: worker@host): database connection refused"
//
// Example output:
//
//	"database connection refused"
func CompactError(errorMsg string) string {
	if errorMsg == "" {
		return ""
	}

	// Strategy: Find the deepest/last meaningful error message
	// Temporal errors are typically chained with ": " separators

	// First, remove common wrapper patterns
	compacted := errorMsg

	// Remove "WorkflowFailed: " prefix
	compacted = strings.TrimPrefix(compacted, "WorkflowFailed: ")

	// Remove "Timeout: " prefix
	compacted = strings.TrimPrefix(compacted, "Timeout: ")

	// Remove parenthetical metadata like (type: X, workflowID: Y, runID: Z, ...)
	// This regex matches parenthetical expressions containing known metadata keys
	metadataPattern := regexp.MustCompile(`\s*\([^)]*(?:type:|workflowID:|runID:|initiatedEventID:|startedEventID:|scheduledEventID:|identity:|retryable:)[^)]*\)`)
	compacted = metadataPattern.ReplaceAllString(compacted, "")

	// Split by ": " to find error chain segments
	segments := strings.Split(compacted, ": ")

	// Find the most informative segment (usually the last non-empty one)
	// But skip generic wrapper messages
	wrapperPhrases := []string{
		"child workflow execution error",
		"activity error",
		"child workflow at depth",
		"leaf workflow failed at depth",
		"order failed",
		"payment error",
		"shipping error",
		"validation failed",
		"timeout workflow failed",
		"retry exhaustion",
		"activityfailed",  // from formatted output
		"workflowfailed",  // from formatted output
		"timeout",         // generic timeout prefix
	}

	// Work backwards to find the core error
	var coreError string
	for i := len(segments) - 1; i >= 0; i-- {
		segment := strings.TrimSpace(segments[i])
		if segment == "" {
			continue
		}

		// Check if this is a wrapper phrase
		isWrapper := false
		segmentLower := strings.ToLower(segment)
		for _, phrase := range wrapperPhrases {
			if strings.HasPrefix(segmentLower, strings.ToLower(phrase)) {
				isWrapper = true
				break
			}
		}

		if !isWrapper {
			coreError = segment
			break
		}
	}

	// If we didn't find a non-wrapper segment, use the last segment
	if coreError == "" && len(segments) > 0 {
		coreError = strings.TrimSpace(segments[len(segments)-1])
	}

	// Clean up any remaining metadata patterns that might be inline
	// Remove things like "(type: wrapError, retryable: true)"
	inlineMetadata := regexp.MustCompile(`\s*\(type:\s*\w+(?:,\s*retryable:\s*\w+)?\)`)
	coreError = inlineMetadata.ReplaceAllString(coreError, "")

	// Strip "ActivityName - " prefix pattern (common in formatted output)
	// ActivityName is PascalCase word followed by " - "
	activityPrefix := regexp.MustCompile(`^[A-Z][a-zA-Z]+Activity\s*-\s*`)
	coreError = activityPrefix.ReplaceAllString(coreError, "")

	// Also strip "ActivityType: " prefix
	activityTypePrefix := regexp.MustCompile(`^ActivityFailed:\s*[A-Z][a-zA-Z]+Activity\s*-\s*`)
	coreError = activityTypePrefix.ReplaceAllString(coreError, "")

	// Clean up double spaces and trim
	coreError = strings.Join(strings.Fields(coreError), " ")

	return coreError
}

// CompactErrorWithContext returns a compact error with optional context prefix.
// If the error is an activity failure, it includes the activity name.
// If it's a timeout, it indicates that.
func CompactErrorWithContext(errorMsg string) string {
	if errorMsg == "" {
		return ""
	}

	core := CompactError(errorMsg)

	// Check for specific error types and add context
	lowerMsg := strings.ToLower(errorMsg)

	// Activity timeout
	if strings.Contains(lowerMsg, "activity") && strings.Contains(lowerMsg, "timeout") {
		if !strings.Contains(strings.ToLower(core), "timeout") {
			return "activity timeout: " + core
		}
	}

	// Workflow timeout
	if strings.Contains(lowerMsg, "workflow execution timed out") {
		return "workflow timeout"
	}

	return core
}

