# Temporal Workflow CLI - Debugging Rules

When debugging Temporal workflows, use the `temporal workflow` CLI commands for structured, machine-readable output.

## Commands

### Find Recent Failures
```bash
temporal workflow list --failed --since 1h --output json
temporal workflow list --failed --since 1h --follow-children --leaf-only --compact-errors --output json
temporal workflow list --failed --since 1h --group-by error --output mermaid
```

### Trace a Workflow Chain
```bash
temporal workflow describe --trace-root-cause --workflow-id <id> --output json
temporal workflow describe --trace-root-cause --workflow-id <id> --output mermaid
# Note: trace always follows children automatically. Use --depth to limit.
```

### Check Event Timeline
```bash
temporal workflow show --compact --workflow-id <id> --output json
temporal workflow show --compact --workflow-id <id> --output mermaid
temporal workflow show --compact --workflow-id <id> --compact --output mermaid
```

### Check Workflow State
```bash
temporal workflow describe --pending --workflow-id <id> --output json
temporal workflow describe --pending --workflow-id <id> --output mermaid
```

## Key Flags

| Flag | Purpose |
|------|---------|
| `--output json` | Structured output for parsing |
| `--output mermaid` | Visual diagrams (flowchart, sequence, pie) |
| `--follow-children` | Include child workflows and Nexus operations |
| `--leaf-only` | Only show deepest failures (skip wrapper errors) |
| `--compact-errors` | Remove verbose error context |
| `--group-by error` | Aggregate failures by error message |
| `--group-by status` | Aggregate by workflow status |
| `--group-by namespace` | Aggregate by namespace |

## When to Use Each Command

| Situation | Command |
|-----------|---------|
| Workflow failed, need root cause | `temporal workflow describe --trace-root-cause --workflow-id <id> --output json` |
| Multiple failures, need patterns | `temporal workflow list --failed --since 1h --group-by error --output mermaid` |
| Workflow stuck, need to see pending work | `temporal workflow describe --pending --workflow-id <id> --output mermaid` |
| Race condition suspected | `temporal workflow show --compact --workflow-id <id> --output mermaid` |
| Child workflow failed | `temporal workflow describe --trace-root-cause --workflow-id <id> --output mermaid` (follows children automatically) |
| Error message too verbose | Add `--compact-errors` to any failure command |

## Output Formats

### JSON Output
Use for programmatic analysis:
```bash
temporal workflow describe --trace-root-cause --workflow-id <id> --output json | jq '.root_cause'
temporal workflow list --failed --since 1h --output json | jq '.total_count'
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
   temporal workflow list --failed --since 10m --output json
   ```

2. **Trace the failure:**
   ```bash
   temporal workflow describe --trace-root-cause --workflow-id <id> --output mermaid
   ```

3. **If child workflows involved:**
   ```bash
   # trace automatically follows children
   temporal workflow describe --trace-root-cause --workflow-id <id> --output mermaid
   ```

4. **If timing issue suspected:**
   ```bash
   temporal workflow show --compact --workflow-id <id> --output mermaid
   ```

5. **If workflow stuck:**
   ```bash
   temporal workflow describe --pending --workflow-id <id> --output mermaid
   ```

6. **Analyze failure patterns:**
   ```bash
   temporal workflow list --failed --since 1h --group-by error --output mermaid
   ```
