# Temporal Agent CLI - Debugging Rules

When debugging Temporal workflows, use the `temporal agent` CLI commands for structured, machine-readable output.

## Commands

### Find Recent Failures
```bash
temporal agent failures --since 1h --format json
temporal agent failures --since 1h --follow-children --leaf-only --compact-errors --format json
temporal agent failures --since 1h --group-by error --format mermaid
```

### Trace a Workflow Chain
```bash
temporal agent trace --workflow-id <id> --format json
temporal agent trace --workflow-id <id> --format mermaid
# Note: trace always follows children automatically. Use --depth to limit.
```

### Check Event Timeline
```bash
temporal agent timeline --workflow-id <id> --format json
temporal agent timeline --workflow-id <id> --format mermaid
temporal agent timeline --workflow-id <id> --compact --format mermaid
```

### Check Workflow State
```bash
temporal agent state --workflow-id <id> --format json
temporal agent state --workflow-id <id> --format mermaid
```

## Key Flags

| Flag | Purpose |
|------|---------|
| `--format json` | Structured output for parsing |
| `--format mermaid` | Visual diagrams (flowchart, sequence, pie) |
| `--follow-children` | Include child workflows and Nexus operations |
| `--leaf-only` | Only show deepest failures (skip wrapper errors) |
| `--compact-errors` | Remove verbose error context |
| `--group-by error` | Aggregate failures by error message |
| `--group-by status` | Aggregate by workflow status |
| `--group-by namespace` | Aggregate by namespace |

## When to Use Each Command

| Situation | Command |
|-----------|---------|
| Workflow failed, need root cause | `temporal agent trace --workflow-id <id> --format json` |
| Multiple failures, need patterns | `temporal agent failures --since 1h --group-by error --format mermaid` |
| Workflow stuck, need to see pending work | `temporal agent state --workflow-id <id> --format mermaid` |
| Race condition suspected | `temporal agent timeline --workflow-id <id> --format mermaid` |
| Child workflow failed | `temporal agent trace --workflow-id <id> --format mermaid` (follows children automatically) |
| Error message too verbose | Add `--compact-errors` to any failure command |

## Output Formats

### JSON Output
Use for programmatic analysis:
```bash
temporal agent trace --workflow-id <id> --format json | jq '.root_cause'
temporal agent failures --since 1h --format json | jq '.total_count'
```

### Mermaid Output
Use for visualization:
- `trace` → Flowchart showing workflow chain
- `timeline` → Sequence diagram showing events
- `failures --group-by` → Pie chart showing distribution
- `state` → State diagram showing pending work

## Debugging Workflow

1. **Find what failed:**
   ```bash
   temporal agent failures --since 10m --format json
   ```

2. **Trace the failure:**
   ```bash
   temporal agent trace --workflow-id <id> --format mermaid
   ```

3. **If child workflows involved:**
   ```bash
   # trace automatically follows children
   temporal agent trace --workflow-id <id> --format mermaid
   ```

4. **If timing issue suspected:**
   ```bash
   temporal agent timeline --workflow-id <id> --format mermaid
   ```

5. **If workflow stuck:**
   ```bash
   temporal agent state --workflow-id <id> --format mermaid
   ```

6. **Analyze failure patterns:**
   ```bash
   temporal agent failures --since 1h --group-by error --format mermaid
   ```
