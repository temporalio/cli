# Temporal Agent Demo

This demo project demonstrates the `temporal workflow` commands for AI-assisted debugging.

## Overview

The demo includes several workflow scenarios:

1. **SimpleSuccessWorkflow** - A basic successful workflow with one activity
2. **OrderWorkflow** - An order processing workflow with child workflows (PaymentWorkflow, ShippingWorkflow)
3. **NestedFailureWorkflow** - A deeply nested workflow chain that fails at the leaf level

## Setup

### Prerequisites

- Go 1.23+
- **Temporal Go SDK v1.37.0+** (required for API key authentication)

### Environment Variables

**For Temporal Cloud (Production):**
```bash
export TEMPORAL_ADDRESS="us-east-1.aws.api.temporal.io:7233"
export TEMPORAL_NAMESPACE="moedash-prod.a2dd6"
export TEMPORAL_API_KEY="$(cat ../../prod-temporal-api-key.txt)"
export TEMPORAL_TASK_QUEUE="agent-demo"
```

**For Temporal Cloud (Staging):**
```bash
export TEMPORAL_ADDRESS="us-west-2.aws.api.tmprl-test.cloud:7233"
export TEMPORAL_NAMESPACE="moedash.temporal-dev"
export TEMPORAL_API_KEY="$(cat ../../staging-temporal-api-key.txt)"
export TEMPORAL_TASK_QUEUE="agent-demo"
```
> **Note:** Staging uses a self-signed certificate. The worker/starter auto-detect staging URLs and skip TLS verification. For CLI commands, add `--tls-disable-host-verification`.

**For Local Dev Server:**
```bash
export TEMPORAL_ADDRESS="localhost:7233"
export TEMPORAL_NAMESPACE="default"
export TEMPORAL_TASK_QUEUE="agent-demo"
```

### Install Dependencies

```bash
go mod tidy
```

### SDK Version Note

This demo requires **Temporal Go SDK v1.37.0+** for proper API key authentication. Earlier SDK versions may fail with "Request unauthorized" errors even with valid credentials. The demo uses `go.temporal.io/sdk/contrib/envconfig` for client configuration, matching the CLI's approach.

## Running the Demo

### 1. Start the Worker

In one terminal:

```bash
go run ./worker
```

### 2. Start Workflows

In another terminal:

```bash
# Run all scenarios
go run ./starter -scenario all

# Or run individual scenarios:
go run ./starter -scenario success
go run ./starter -scenario payment-fail
go run ./starter -scenario shipping-fail
go run ./starter -scenario nested-fail
```

## Using Temporal Workflow Commands

After workflows have run, use the agent commands to analyze them.

> **For staging:** Add `--tls-disable-host-verification` to all commands.

### List Recent Failures

```bash
temporal workflow list --failed \
    --address $TEMPORAL_ADDRESS \
    --namespace $TEMPORAL_NAMESPACE \
    --api-key $TEMPORAL_API_KEY \
    --tls \
    --since 1h \
    --follow-children \
    --format json | jq
```

### Trace a Workflow Chain

```bash
# Find the deepest failure in an order workflow
temporal workflow describe --trace-root-cause \
    --address $TEMPORAL_ADDRESS \
    --namespace $TEMPORAL_NAMESPACE \
    --api-key $TEMPORAL_API_KEY \
    --tls \
    -w order-payment-fail-XXXXXX \
    --format json | jq

# Trace the nested failure workflow (3 levels deep)
temporal workflow describe --trace-root-cause \
    --address $TEMPORAL_ADDRESS \
    --namespace $TEMPORAL_NAMESPACE \
    --api-key $TEMPORAL_API_KEY \
    --tls \
    -w nested-failure-XXXXXX \
    --format json | jq
```

### Get Workflow Timeline

```bash
temporal workflow show --compact \
    --address $TEMPORAL_ADDRESS \
    --namespace $TEMPORAL_NAMESPACE \
    --api-key $TEMPORAL_API_KEY \
    --tls \
    -w order-success-XXXXXX \
    --compact \
    --format json | jq
```

## Workflow Scenarios

### Payment Failure Chain

```
OrderWorkflow (ORD-XXX-X)
  └── PaymentWorkflow (payment-ORD-XXX-X)
        └── ProcessPaymentActivity → FAILS: "payment gateway connection timeout"
```

### Shipping Failure Chain

```
OrderWorkflow (ORD-XXX-Y)
  └── PaymentWorkflow (payment-ORD-XXX-Y) → SUCCESS
  └── ShippingWorkflow (shipping-ORD-XXX-Y)
        └── ShipOrderActivity → FAILS: "warehouse inventory depleted"
```

### Nested Failure Chain

```
NestedFailureWorkflow (depth=0)
  └── NestedFailureWorkflow (depth=1)
        └── NestedFailureWorkflow (depth=2)
              └── NestedFailureWorkflow (depth=3)
                    └── FailingActivity → FAILS: "database connection refused"
```

The `temporal workflow describe --trace-root-cause` command will automatically traverse this entire chain
and identify the leaf failure with its root cause.

