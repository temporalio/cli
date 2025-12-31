# Debug Loop Test

This example tests the end-to-end AI agent debug loop:
1. Run a workflow with an intentional bug
2. Use `temporal agent` CLI to diagnose the failure
3. Have an AI agent (Cursor) propose a fix
4. Apply the fix and verify success

## The Bug

The `ProcessOrderWorkflow` has a subtle retry configuration bug:

- **Activity:** `CheckInventory` simulates a flaky inventory service that needs 3 attempts to succeed
- **Bug:** The retry policy has `MaximumAttempts: 2`, so the activity never gets a 3rd attempt
- **Result:** The workflow fails with "activity retry limit exceeded"

## Running the Test

### Step 1: Start Local Dev Server

```bash
temporal server start-dev
```

### Step 2: Start Worker

```bash
cd examples/debug-loop
go run ./worker
```

### Step 3: Run the Buggy Workflow

```bash
go run ./starter
```

Expected output:
```
Starting order workflow: order-1234567890
Waiting for workflow completion...
Workflow FAILED: activity retry limit exceeded

=== DEBUG INSTRUCTIONS ===
Use the temporal agent CLI to diagnose this failure:

  temporal agent failures --namespace default --since 5m -o json
  temporal agent trace --workflow-id order-1234567890 --namespace default -o json
  temporal agent timeline --workflow-id order-1234567890 --namespace default --compact -o json
```

### Step 4: Diagnose with Temporal Agent CLI

```bash
# Find recent failures
temporal agent failures --namespace default --since 5m -o json

# Trace the workflow chain
temporal agent trace --workflow-id order-1234567890 --namespace default -o json

# Get compact timeline
temporal agent timeline --workflow-id order-1234567890 --namespace default --compact -o json
```

### Step 5: AI Analysis (Cursor)

Prompt Cursor with:
```
"A workflow just failed with 'activity retry limit exceeded'. 
Use the temporal agent CLI to diagnose the root cause.
The workflow ID is [workflow-id], namespace is default."
```

### Step 6: Apply Fix

The fix is to change `MaximumAttempts: 2` to `MaximumAttempts: 3` in `workflows/order.go`:

```go
// Before (buggy)
RetryPolicy: &temporal.RetryPolicy{
    MaximumAttempts: 2, // BUG: Should be 3 or more!
}

// After (fixed)
RetryPolicy: &temporal.RetryPolicy{
    MaximumAttempts: 3,
}
```

### Step 7: Verify Fix

1. Restart the worker
2. Run the starter again
3. The workflow should complete successfully

## Expected Timeline Output

```json
{
  "events": [
    {"type": "WorkflowExecutionStarted", ...},
    {"type": "ActivityTaskScheduled", "name": "ProcessPayment", ...},
    {"type": "ActivityTaskCompleted", "name": "ProcessPayment", ...},
    {"type": "ActivityTaskScheduled", "name": "CheckInventory", ...},
    {"type": "ActivityTaskFailed", "name": "CheckInventory", "attempt": 1, "error": "inventory service unavailable (attempt 1/3)"},
    {"type": "ActivityTaskScheduled", "name": "CheckInventory", ...},
    {"type": "ActivityTaskFailed", "name": "CheckInventory", "attempt": 2, "error": "inventory service unavailable (attempt 2/3)"},
    {"type": "WorkflowExecutionFailed", "error": "activity retry limit exceeded"}
  ]
}
```

The key insight from the timeline is:
1. Activity failed on attempt 1 and 2
2. Error message says "attempt X/3" (indicating 3 attempts are needed)
3. But only 2 attempts were made before the workflow failed
4. **Conclusion:** MaximumAttempts is too low

