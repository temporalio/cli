# Temporal Agent CLI Rules for Cursor

When debugging Temporal workflows, use the `temporal agent` CLI commands instead of logs.

## Available Commands

### Find Failures
```bash
temporal agent failures --since 1h --format json
temporal agent failures --since 1h --follow-children --leaf-only --compact-errors --format json
temporal agent failures --since 1h --group-by error --format json
```

### Trace Workflow Chain
```bash
temporal agent trace --workflow-id <id> --format json
temporal agent trace --workflow-id <id> --format mermaid
```

### Get Timeline
```bash
temporal agent timeline --workflow-id <id> --format json
temporal agent timeline --workflow-id <id> --compact --format json
temporal agent timeline --workflow-id <id> --format mermaid
```

### Check State
```bash
temporal agent state --workflow-id <id> --format json
temporal agent state --workflow-id <id> --format mermaid
```

## Key Flags

- `--follow-children` - Traverse child workflows to find leaf failures
- `--leaf-only` - Show only leaf failures (de-duplicate chains)
- `--compact-errors` - Strip wrapper context, show core error
- `--group-by error|namespace|type|status` - Aggregate failures
- `--format mermaid` - Output visual diagrams
- `--format json` - Structured JSON output

## Visualization

Use `--format mermaid` to generate diagrams:
- `trace --format mermaid` → Flowchart of workflow chain
- `timeline --format mermaid` → Sequence diagram of events
- `state --format mermaid` → State diagram with pending work
- `failures --group-by error --format mermaid` → Pie chart

## Debugging Workflow

1. Start with `agent trace` to see the chain and root cause
2. Use `--format mermaid` to visualize if complex
3. Use `agent failures --leaf-only` to see actual failures
4. Use `agent state` to check pending work on running workflows
5. Use `--group-by` to find patterns in multiple failures

## Example Session

User: "The order workflow failed"

You should:
1. Run `temporal agent trace --workflow-id order-123 --format json`
2. If complex, add `--format mermaid` for visual
3. Identify the leaf failure and root cause
4. Explain what went wrong
5. Suggest a fix

