# Debug Loop Test: TOCTOU Race Condition

This example tests the end-to-end AI agent debug loop with a **realistic TOCTOU (Time-of-Check to Time-of-Use) race condition** that requires workflow timeline analysis to diagnose.

## The Bug

The `ProcessOrderWorkflow` has a subtle race condition:

1. **Parallel Check Phase**: Inventory is checked for ALL items simultaneously
2. **Delay Phase**: A 200ms processing delay occurs (simulating real-world latency)
3. **Sequential Reserve Phase**: Inventory is reserved one item at a time

**Problem**: During the delay, a competing order can claim limited-stock items. All checks pass, but reservations fail because inventory state changed between check and reserve.

### Why This Bug Is Realistic

- The **error message alone is misleading**: `insufficient inventory for KEYBOARD-03: requested 1, available 0`
- A naive analysis might conclude "the inventory was wrong" or "the check was broken"
- The **inventory check DID pass** - you can verify this in the timeline
- The real issue is a **race condition** that requires timing analysis to diagnose

### What Makes Diagnosis Non-Trivial

1. The error says "available 0" but the check showed "available 1"
2. The workflow logic appears correct (check → then reserve)
3. You need to see **WHEN** events occurred to understand the race
4. You need to recognize the **parallel check + sequential reserve** anti-pattern

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

### Step 3: Run the Race Condition Scenario

```bash
go run ./starter --scenario race
```

Expected output:
```
=== RACE CONDITION SIMULATION ===
Two orders will compete for the same item (KEYBOARD-03, only 1 in stock)

Starting main order: order-123456
  Items: LAPTOP-001 x1, MOUSE-002 x2, KEYBOARD-03 x1
Main order started: order-123456 (run ID: abc...)
Competing order started: competing-123456

=== RESULTS ===
Main order FAILED: insufficient inventory for KEYBOARD-03: requested 1, available 0
Competing order SUCCEEDED

=== DEBUG CHALLENGE ===
One order's inventory check PASSED but reservation FAILED.
This is a classic TOCTOU race condition!
```

### Step 4: Diagnose with Temporal Agent CLI

```bash
# Get the trace - shows root cause
temporal agent trace --workflow-id order-123456 --namespace default --format json

# THIS IS KEY: Get the timeline to see the race condition
temporal agent timeline --workflow-id order-123456 --namespace default --format json
```

## What the Timeline Reveals

```
T+0ms     CheckInventory (LAPTOP-001) scheduled   ─┐
T+0ms     CheckInventory (MOUSE-002) scheduled    ├── Parallel checks
T+0ms     CheckInventory (KEYBOARD-03) scheduled  ─┘
T+1ms     CheckInventory (LAPTOP-001) completed: Available=true
T+2ms     CheckInventory (MOUSE-002) completed: Available=true  
T+3ms     CheckInventory (KEYBOARD-03) completed: Available=true, InStock=1  ← CHECK PASSED!
T+200ms   TimerFired (processing delay)           ← RACE WINDOW
T+205ms   ReserveInventory (LAPTOP-001) completed
T+210ms   ReserveInventory (MOUSE-002) completed
T+215ms   ReserveInventory (KEYBOARD-03) FAILED   ← RESERVE FAILED!
          Error: "insufficient inventory: requested 1, available 0"
```

**The key insight**: At T+3ms, KEYBOARD-03 showed `InStock=1`. At T+215ms, reservation failed with `available=0`. 

The competing order claimed the keyboard during the 200ms delay!

## AI Agent Diagnosis Prompt

```
A workflow failed with "insufficient inventory for KEYBOARD-03: requested 1, available 0".
The logs show the inventory check passed, but the reservation failed.

Use temporal agent to diagnose:
  temporal agent trace --workflow-id [id] --namespace default --format json
  temporal agent timeline --workflow-id [id] --namespace default --format json

Questions to answer:
1. Did the inventory check pass? What did it show?
2. How much time passed between check and reserve?
3. What could have changed the inventory during that time?
```

## Expected AI Analysis

A good AI diagnosis should identify:

1. **Timeline Analysis**: Checks at T+3ms showed `InStock=1`, reservation at T+215ms showed `available=0`
2. **Pattern Recognition**: Parallel checks + delay + sequential reserves = TOCTOU vulnerability
3. **Root Cause**: Another workflow reserved the item during the 200ms processing delay
4. **Fix Proposals**:
   - Atomic check-and-reserve in a single activity
   - Remove the delay between check and reserve
   - Use optimistic locking/versioning
   - Re-validate inventory immediately before each reservation

## The Fix

Option 1: **Atomic Operation**
```go
// Instead of separate check + reserve, do both atomically
func CheckAndReserveInventory(ctx context.Context, input CheckAndReserveInput) (*ReserveResult, error) {
    // Single transaction that checks and reserves
}
```

Option 2: **Re-validate Before Reserve**
```go
// Check again right before reserving
for _, item := range items {
    // Re-check inventory (no caching)
    check := checkInventory(item)
    if !check.Available {
        return nil, fmt.Errorf("item %s became unavailable", item.SKU)
    }
    // Reserve immediately after check
    reserve(item)
}
```

Option 3: **Remove Unnecessary Delay**
```go
// Don't sleep between check and reserve!
// The 200ms delay creates an unnecessary race window
```

## Files

| File | Purpose |
|------|---------|
| `workflows/order.go` | Order workflow with TOCTOU race condition |
| `activities/inventory.go` | Inventory check/reserve activities |
| `worker/main.go` | Worker registration |
| `starter/main.go` | Race condition scenario launcher |

## Key Learning

The `temporal agent timeline` command is essential for diagnosing race conditions because it shows:
- **When** each event occurred (precise timestamps)
- **What** the state was at each point (activity results)
- **How long** each phase took (identifying race windows)

Without the timeline, you only see "check passed" and "reserve failed" - not the crucial timing that explains the discrepancy.
