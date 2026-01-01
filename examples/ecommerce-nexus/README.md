# Multi-Namespace E-Commerce Simulation with Nexus

This example demonstrates a multi-namespace e-commerce system using both Nexus endpoints and cross-namespace child workflows. It's designed to validate the `temporal workflow` CLI's tracing capabilities across different cross-service patterns.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           commerce-ns                                    │
│  ┌─────────────────┐    ┌─────────────────────┐                         │
│  │ OrderSagaWF     │───▶│ ReserveInventoryWF  │                         │
│  └────────┬────────┘    └─────────────────────┘                         │
│           │                                                              │
└───────────┼──────────────────────────────────────────────────────────────┘
            │
     ┌──────┴──────┐
     │             │
     ▼ (Nexus)     ▼ (Child WF)
┌────────────────┐ ┌────────────────────────────────────────────────────────┐
│  finance-ns    │ │                    logistics-ns                        │
│ ┌────────────┐ │ │  ┌─────────────┐    ┌──────────────────┐              │
│ │ PaymentWF  │ │ │  │ ShipOrderWF │───▶│ TrackShipmentWF  │              │
│ └─────┬──────┘ │ │  └─────────────┘    └──────────────────┘              │
│       │        │ │                                                        │
│       ▼        │ └────────────────────────────────────────────────────────┘
│ ┌────────────┐ │
│ │ FraudCheck │ │
│ └────────────┘ │
└────────────────┘
```

## Cross-Service Patterns

| From | To | Pattern | Why |
|------|-----|---------|-----|
| commerce → finance | `PaymentWorkflow` | **Nexus** | Team/compliance boundary |
| commerce → logistics | `ShipOrderWorkflow` | **Child WF** | Compare tracing |
| finance → finance | `FraudCheckWorkflow` | **Child WF** | Same namespace |

## Prerequisites

### 1. Create Namespaces (Temporal Cloud)

Create 3 namespaces in Temporal Cloud:

```
moedash.commerce-ns
moedash.finance-ns
moedash.logistics-ns
```

Or use your account ID prefix:
```
<your-account>.commerce-ns
<your-account>.finance-ns
<your-account>.logistics-ns
```

### 2. Configure Nexus Endpoints (Temporal Cloud)

In the Temporal Cloud UI, create a Nexus endpoint:

1. Go to **Nexus** > **Endpoints**
2. Create endpoint: `payment-service`
3. Target namespace: `moedash.finance-ns` (or your finance namespace)
4. Allow caller namespace: `moedash.commerce-ns`
5. Task queue: `finance-tasks`

### 3. Environment Variables

```bash
# API Key (same key works if service account has access to all namespaces)
export TEMPORAL_API_KEY="<your-api-key>"

# Namespace configuration
export COMMERCE_NS="moedash.commerce-ns"
export FINANCE_NS="moedash.finance-ns"  
export LOGISTICS_NS="moedash.logistics-ns"

# For staging
export TEMPORAL_ADDRESS="us-west-2.aws.api.tmprl-test.cloud:7233"

# Optional: for local dev server (single namespace mode)
# export TEMPORAL_ADDRESS="localhost:7233"
# export COMMERCE_NS="default"
# export FINANCE_NS="default"
# export LOGISTICS_NS="default"
```

## Running the Workers

Start all three namespace workers:

```bash
# Terminal 1: Commerce namespace
cd examples/ecommerce-nexus
go run ./commerce-ns/worker

# Terminal 2: Finance namespace
go run ./finance-ns/worker

# Terminal 3: Logistics namespace
go run ./logistics-ns/worker
```

## Running Scenarios

```bash
# All failure scenarios
go run ./starter -scenario all

# Specific scenarios
go run ./starter -scenario nexus-payment-fail    # Payment fails via Nexus
go run ./starter -scenario child-shipping-fail   # Shipping fails via Child WF
go run ./starter -scenario nexus-fraud-detect    # Fraud detection via Nexus chain
go run ./starter -scenario saga-compensation     # Saga with compensation
go run ./starter -scenario deep-chain            # 4-level cross-NS chain
```

## Testing with Agent CLI

### Cross-Namespace API Keys

For cross-namespace tracing, set namespace-specific API keys:

```bash
# Format: TEMPORAL_API_KEY_<NAMESPACE>
# Namespace names are normalized: dots/dashes → underscores, then UPPERCASED
#
# Example for moedash-finance-ns.temporal-dev:
#   → TEMPORAL_API_KEY_MOEDASH_FINANCE_NS_TEMPORAL_DEV

# Primary namespace (commerce) uses TEMPORAL_API_KEY
export TEMPORAL_API_KEY="$(cat staging-commerce-temporal-api-key.txt)"

# Finance namespace
export TEMPORAL_API_KEY_MOEDASH_FINANCE_NS_TEMPORAL_DEV="$(cat staging-finance-temporal-api-key.txt)"

# Logistics namespace
export TEMPORAL_API_KEY_MOEDASH_LOGISTICS_NS_TEMPORAL_DEV="$(cat staging-logistics-temporal-api-key.txt)"
```

### Commands

```bash
# Find failures in commerce namespace (cross-namespace traversal)
temporal workflow failures --namespace $COMMERCE_NS --since 1h \
    --follow-children --follow-namespaces $FINANCE_NS,$LOGISTICS_NS --format json

# Trace a failed order (follows Nexus and child workflows across namespaces)
temporal workflow diagnose --workflow-id order-123 --namespace $COMMERCE_NS \
    --follow-namespaces $FINANCE_NS,$LOGISTICS_NS --format json

# Check workflow state (shows pending Nexus operations)
temporal workflow describe --pending --workflow-id order-123 --namespace $COMMERCE_NS --format json

# With leaf-only and compact errors
temporal workflow failures --namespace $COMMERCE_NS --since 1h \
    --follow-children --follow-namespaces $FINANCE_NS,$LOGISTICS_NS \
    --leaf-only --compact-errors --format json

# Group failures by error type
temporal workflow failures --namespace $COMMERCE_NS --since 1h \
    --follow-children --follow-namespaces $FINANCE_NS,$LOGISTICS_NS \
    --compact-errors --group-by error --format json

# Group by namespace to see which services are failing
temporal workflow failures --namespace $COMMERCE_NS --since 1h \
    --follow-children --follow-namespaces $FINANCE_NS,$LOGISTICS_NS \
    --group-by namespace --format json
```

## Validation Points

This example tests:

1. **Nexus call tracing** - Does `workflow diagnose` follow Nexus calls?
2. **Cross-NS child WF tracing** - Does `workflow diagnose` follow cross-namespace child workflows?
3. **Error propagation** - Do errors from Nexus calls appear in parent workflow?
4. **Leaf-only filtering** - Does `--leaf-only` work with Nexus?
5. **Compact errors** - Does `--compact-errors` strip Nexus wrapper messages?

## Quick Start (Single Namespace Mode)

For quick testing without multiple namespaces, all services run in the same namespace with different task queues. This tests child workflow tracing but not Nexus.

```bash
# Start local dev server
temporal server start-dev

# Set single namespace mode
export TEMPORAL_ADDRESS="localhost:7233"
export COMMERCE_NS="default"
export FINANCE_NS="default"
export LOGISTICS_NS="default"

# Terminal 1: Run all workers
go run ./commerce-ns/worker &
go run ./finance-ns/worker &
go run ./logistics-ns/worker &

# Terminal 2: Run scenarios
go run ./starter -scenario all
```

## Temporal Cloud with Nexus (Full Mode)

For full multi-namespace Nexus testing on staging:

### Step 1: Create Namespaces (Done)

Namespaces created on staging (us-west-2.aws.api.tmprl-test.cloud:7233):
- `moedash-commerce-ns.temporal-dev`
- `moedash-finance-ns.temporal-dev`
- `moedash-logistics-ns.temporal-dev`

### Step 2: Configure Nexus Endpoint

In Temporal Cloud UI → Nexus → Endpoints:
1. **Create endpoint**: `payment-endpoint`
2. **Target namespace**: `moedash-finance-ns.temporal-dev`
3. **Target task queue**: `finance-tasks`
4. **Allowed callers**: Add `moedash-commerce-ns.temporal-dev`

### Step 3: Run Workers (Each Needs Its Own API Key)

```bash
# Terminal 1: Commerce worker
export TEMPORAL_ADDRESS="us-west-2.aws.api.tmprl-test.cloud:7233"
export TEMPORAL_API_KEY="$(cat staging-commerce-temporal-api-key.txt)"
export COMMERCE_NS="moedash-commerce-ns.temporal-dev"
export FINANCE_NS="moedash-finance-ns.temporal-dev"
export LOGISTICS_NS="moedash-logistics-ns.temporal-dev"
export NEXUS_PAYMENT_ENDPOINT="payment-endpoint"
go run ./commerce-ns/worker

# Terminal 2: Finance worker
export TEMPORAL_ADDRESS="us-west-2.aws.api.tmprl-test.cloud:7233"
export TEMPORAL_API_KEY="$(cat staging-finance-temporal-api-key.txt)"
export FINANCE_NS="moedash-finance-ns.temporal-dev"
go run ./finance-ns/worker

# Terminal 3: Logistics worker
export TEMPORAL_ADDRESS="us-west-2.aws.api.tmprl-test.cloud:7233"
export TEMPORAL_API_KEY="$(cat staging-logistics-temporal-api-key.txt)"
export LOGISTICS_NS="moedash-logistics-ns.temporal-dev"
go run ./logistics-ns/worker
```

### Step 4: Run Scenarios

```bash
# Use commerce API key to start workflows
export TEMPORAL_ADDRESS="us-west-2.aws.api.tmprl-test.cloud:7233"
export TEMPORAL_API_KEY="$(cat staging-commerce-temporal-api-key.txt)"
export COMMERCE_NS="moedash-commerce-ns.temporal-dev"
export FINANCE_NS="moedash-finance-ns.temporal-dev"
export LOGISTICS_NS="moedash-logistics-ns.temporal-dev"
export NEXUS_PAYMENT_ENDPOINT="payment-endpoint"

go run ./starter -scenario all
```

Note: Nexus features require Temporal Cloud or self-hosted Temporal with Nexus enabled.

