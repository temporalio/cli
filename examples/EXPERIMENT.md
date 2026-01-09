# Agent Debugging Experiment

This document describes experiments to validate that the `temporal workflow` CLI commands enable AI agents to debug Temporal workflow failures using structured output instead of logs.

## Hypothesis

AI agents can query failures, trace nested workflow chains across namespaces, and get compact timelines and state — without scraping logs or manually traversing workflows.

## Experiment 1: Basic Agent Commands

### Environment

| Setting | Value |
|---------|-------|
| Temporal Environment | Staging (`us-west-2.aws.api.tmprl-test.cloud:7233`) |
| Namespace | `moedash.temporal-dev` |
| CLI Version | Built from source |
| AI Agent | Claude Code (Cursor) |

### Failure Scenarios

| Scenario | Command | Expected Failure |
|----------|---------|------------------|
| Success | `go run ./starter -scenario success` | No failure (control) |
| Payment Fail | `go run ./starter -scenario payment-fail` | Activity fails with "payment gateway connection timeout" |
| Shipping Fail | `go run ./starter -scenario shipping-fail` | Activity fails with "warehouse inventory depleted" |
| Nested Fail | `go run ./starter -scenario nested-fail` | 3-level deep child workflow chain, leaf fails with "database connection refused" |
| Timeout | `go run ./starter -scenario timeout` | Activity times out (5s activity with 2s timeout) |
| Retry Exhaustion | `go run ./starter -scenario retry-exhaustion` | Activity fails 5 times then exhausts retries |
| Multi-Child | `go run ./starter -scenario multi-child` | 3 parallel children, only "validation" child fails |

### Results: 2025-12-29

| Test | Tool Used | Root Cause Found | Score | Notes |
|------|-----------|------------------|-------|-------|
| Test 1 | `workflow failures` | 6/6 failure types identified | 95/100 | All failures found with clear root causes |
| Test 2 | `workflow diagnose` | "database connection refused" at depth 3 | 100/100 | Perfect chain traversal |
| Test 3 | `workflow show --compact` | ValidationWorkflow failed with invalid SKU | 100/100 | Clear child workflow timeline |
| Test 4 | `workflow diagnose` | "activity StartToClose timeout" | 100/100 | Correctly identified timeout vs app error |
| Test 5 | `workflow failures --error-contains` | Found 2 timeout-related failures | 100/100 | Filter worked correctly |

**Overall Score:** 99/100

---

## Experiment 2: Multi-Namespace Nexus Traversal

### Environment

| Setting | Value |
|---------|-------|
| Temporal Environment | Staging (`us-west-2.aws.api.tmprl-test.cloud:7233`) |
| Namespaces | `moedash-commerce-ns.temporal-dev`, `moedash-finance-ns.temporal-dev`, `moedash-logistics-ns.temporal-dev` |
| Example | `examples/ecommerce-nexus/` |

### Scenarios Tested

| Scenario | Chain | Expected Failure |
|----------|-------|------------------|
| Nexus Payment Fail | commerce → finance (Nexus) | Fraud detection fails |
| Child Shipping Fail | commerce → logistics (child workflow) | Shipping carrier error |
| Deep Chain | commerce → finance → fraud-check | 3-level Nexus + child chain |

### Results: 2025-12-30

| Metric | Target | Result | Status |
|--------|--------|--------|--------|
| Time to first failure found | < 30 seconds | 3.1 seconds | ✅ PASS |
| Root cause accuracy | 100% | 100% (all failures correctly identified) | ✅ PASS |
| Chain depth accuracy | 100% | 100% (depth 2 for Nexus chains) | ✅ PASS |
| Cross-NS traversal success | 100% | 100% (commerce-ns → finance-ns) | ✅ PASS |
| Token efficiency | < 1000 bytes per failure | 685 bytes/failure | ✅ PASS |

### Key Findings

- Cross-namespace Nexus traversal correctly followed fraud workflows from commerce-ns to finance-ns
- `--compact-errors` effectively stripped verbose wrapper messages
- `--leaf-only` reduced results by 69%, eliminating duplicate parent/child entries
- Namespace-specific API keys worked seamlessly via `TEMPORAL_API_KEY_<NAMESPACE>` pattern

---

## Experiment 3: Blind AI Diagnosis (TOCTOU Race Condition)

### Environment

| Setting | Value |
|---------|-------|
| Temporal Environment | Local dev server |
| Namespace | `default` |
| Example | `examples/debug-loop-fresh/` (hint-free version) |
| AI Agent | Claude (separate LLM session) |

### The Challenge

The `debug-loop-fresh` example contains a TOCTOU race condition with all hints removed. The LLM was given only:

> "I've created a sample example under `examples/debug-loop-fresh`, and I want you to find and fix its issue with the use of temporal workflow CLI"

### LLM's Diagnosis Process

1. **Ran the scenario** - Started worker and triggered race condition
2. **Used `temporal workflow describe --trace-root-cause`** - Found `ReserveInventory` failed for KEYBOARD-03
3. **Used `temporal workflow show --compact`** - Analyzed timestamps of both workflows
4. **Built a race timeline** - Correlated events across both orders:

| Time | Main Order | Competing Order |
|------|------------|-----------------|
| 03:37:04.708 | CheckInventory (all 3) ✓ | |
| 03:37:04.711 | | CheckInventory ✓ |
| 03:37:05.723 | | **ReserveInventory ✓** (takes keyboard) |
| 03:37:05.730 | Reserve KEYBOARD **FAILED** | Completed ✓ |

5. **Proposed the fix** - Atomic `CheckAndReserveInventory` activity
6. **Verified the fix** - Both orders now behave deterministically

### Results

| Metric | Result |
|--------|--------|
| Root cause identified | ✅ TOCTOU race condition |
| Timeline analysis used | ✅ Cross-workflow timing correlation |
| Fix proposed | ✅ Atomic check-and-reserve |
| Fix verified | ✅ Deterministic behavior |
| Human intervention needed | ❌ None |

**This validates the core thesis:** An LLM can autonomously diagnose complex timing bugs using only `temporal workflow` CLI output.

---

## Features Implemented

Based on experiment findings, the following improvements were made:

### Phase 1: Core Commands

| Feature | Status | Command |
|---------|--------|---------|
| Find recent failures | ✅ Done | `temporal workflow list --failed` |
| Trace workflow chain | ✅ Done | `temporal workflow describe --trace-root-cause` |
| Workflow timeline | ✅ Done | `temporal workflow show --compact` |

### Phase 2: Filtering & Compaction

| Feature | Status | Flag/Command |
|---------|--------|--------------|
| Error message filter | ✅ Done | `--error-contains` |
| Multiple status values | ✅ Done | `--status Failed,TimedOut` |
| Leaf-only failures | ✅ Done | `--leaf-only` |
| Compact error messages | ✅ Done | `--compact-errors` |
| Follow child workflows | ✅ Done | `--follow-children` |

### Phase 3: State & Aggregation

| Feature | Status | Flag/Command |
|---------|--------|--------------|
| Workflow state | ✅ Done | `temporal workflow describe --pending` |
| Pending activities | ✅ Done | Included in state output |
| Pending Nexus operations | ✅ Done | Included in state output |
| Group failures by type | ✅ Done | `--group-by type\|namespace\|status\|error` |

### Phase 4: Cross-Namespace

| Feature | Status | Notes |
|---------|--------|-------|
| Nexus chain traversal | ✅ Done | Follows Nexus operations across namespaces |
| Namespace-specific API keys | ✅ Done | `TEMPORAL_API_KEY_<NAMESPACE>` env vars |
| Cross-NS documentation | ✅ Done | Added to README and examples |

### Phase 5: AI Tool Specs

| Feature | Status | Format |
|---------|--------|--------|
| OpenAI function spec | ✅ Done | `temporal tool-spec --format openai` |
| LangChain tool spec | ✅ Done | `temporal tool-spec --format langchain` |
| Claude tool spec | ✅ Done | `temporal tool-spec --format claude` |

### Phase 6: Visualization

| Feature | Status | Flag/Command |
|---------|--------|--------------|
| Trace flowchart | ✅ Done | `temporal workflow describe --trace-root-cause --output mermaid` |
| Timeline sequence diagram | ✅ Done | `temporal workflow show --compact --output mermaid` |
| State diagram | ✅ Done | `temporal workflow describe --pending --output mermaid` |
| Failures pie chart | ✅ Done | `temporal workflow list --failed --group-by error --output mermaid` |
| Failures flowchart | ✅ Done | `temporal workflow list --failed --output mermaid` |

---

## Comparison: Agent Commands vs Log-Based Debugging

| Aspect | Agent Commands | Log-Based (LogQL/grep) |
|--------|----------------|------------------------|
| Time to root cause | ~3-5 seconds | 5-30 minutes |
| Token consumption | ~500 tokens per query | ~5000+ tokens |
| Accuracy | 100% (structured data) | Variable |
| Domain knowledge required | Minimal | High |
| Manual steps | 1 command | 5+ steps |
| Cross-namespace correlation | Automatic | Manual |
| Race condition diagnosis | Timeline timestamps | Nearly impossible |

---

## Success Criteria Validation

| Criterion | Status | Evidence |
|-----------|--------|----------|
| AI finds failures without LogQL | ✅ | All experiments used `temporal workflow` only |
| Root cause accuracy | ✅ | 100% in all tests |
| Low token cost | ✅ | ~10x reduction vs logs |
| Cross-namespace traversal | ✅ | Nexus chains fully traced |
| Timing bug diagnosis | ✅ | Race condition identified from timeline |
| Autonomous fix proposal | ✅ | LLM proposed correct atomic operation fix |

---

## Conclusion

The `temporal workflow` CLI commands successfully achieve the goals:

1. **Agent-native feedback loop**: AI agents effectively debug Temporal workflow failures using structured output
2. **No logs required**: All debugging done via `temporal workflow` commands
3. **Automatic chain traversal**: Traces follow child workflows and Nexus operations across namespaces
4. **Root cause extraction**: Leaf failures clearly identified with `--leaf-only`
5. **Error compaction**: `--compact-errors` strips wrapper context for cleaner output
6. **Timing analysis**: Timeline timestamps enable race condition diagnosis
7. **Low token cost**: Structured JSON is ~10x more efficient than raw logs
8. **Autonomous debugging**: LLM successfully diagnosed and fixed a TOCTOU bug without hints
9. **Mermaid visualization**: `--output mermaid` generates visual diagrams for human-in-the-loop debugging

**Temporal's execution history + agent-optimized CLI = effective AI debugging feedback loop.**
