# TicketDrop Example

A distributed ticket sales system demonstrating Temporal workflow patterns and `temporal agent` CLI debugging.

**Scenario:** 50,000 fans trying to buy 500 concert tickets in 10 seconds. Race conditions, timeouts, and cascading failures guaranteed.

## What You'll Learn

- Building concurrent-safe reservation systems
- Saga pattern for compensating transactions
- Queue management for fair ordering
- Race condition debugging with `temporal agent timeline`
- Failure analysis with `temporal agent failures --group-by`

## Prerequisites

```bash
# Start Temporal server
temporal server start-dev

# Verify agent commands work
temporal agent --help
```

## Quick Start

See [PLAN.md](./PLAN.md) for the step-by-step guide to build this with AI assistance.

## Architecture

```
User Request → Queue Manager → Ticket Purchase Workflow
                                    │
                                    ├── ReserveSeat (activity)
                                    │       ↓
                                    ├── ProcessPayment (activity)
                                    │       ↓
                                    ├── IssueTicket (activity)
                                    │       ↓
                                    └── SendConfirmation (child workflow)
                                            ├── Email (activity)
                                            └── SMS (activity)
```

## Key Debugging Scenarios

### 1. Double-Booking Race Condition

Two users grab the same seat simultaneously:

```bash
# Check the timeline for both purchases
temporal agent timeline --workflow-id purchase-user-1 --format mermaid
temporal agent timeline --workflow-id purchase-user-2 --format mermaid
```

### 2. Payment Stuck

Payment gateway timing out:

```bash
# See what's pending
temporal agent state --workflow-id purchase-xyz --format mermaid
```

### 3. Load Test Analysis

After running 100 concurrent users:

```bash
# See failure distribution
temporal agent failures --since 5m --group-by error --format mermaid
```

## Files

- `PLAN.md` - Step-by-step building guide with prompts
- `.cursorrules` - AI assistant configuration for debugging

## Related Examples

- [ai-research-agent](../ai-research-agent/) - Another AI-guided tutorial
- [debug-loop](../debug-loop/) - E2E debugging demonstration

