# Temporal Workflow CLI Rules for Cursor

When debugging Temporal workflows, use the `temporal workflow` CLI commands instead of logs.

## Available Commands

### Find Failures
```bash
temporal workflow list --failed --since 1h --output json
temporal workflow list --failed --since 1h --follow-children --leaf-only --compact-errors --output json
temporal workflow list --failed --since 1h --group-by error --output json
```

### Trace Workflow Chain
```bash
temporal workflow describe --trace-root-cause --workflow-id <id> --output json
temporal workflow describe --trace-root-cause --workflow-id <id> --output mermaid
```

### Get Timeline
```bash
temporal workflow show --compact --workflow-id <id> --output json
temporal workflow show --compact --workflow-id <id> --compact --output json
temporal workflow show --compact --workflow-id <id> --output mermaid
```

### Check State
```bash
temporal workflow describe --pending --workflow-id <id> --output json
temporal workflow describe --pending --workflow-id <id> --output mermaid
```

## Key Flags

- `--follow-children` - Traverse child workflows to find leaf failures
- `--leaf-only` - Show only leaf failures (de-duplicate chains)
- `--compact-errors` - Strip wrapper context, show core error
- `--group-by error|namespace|type|status` - Aggregate failures
- `--output mermaid` - Output visual diagrams
- `--output json` - Structured JSON output

## Visualization

Use `--output mermaid` to generate diagrams:
- `trace --output mermaid` → Flowchart of workflow chain
- `timeline --output mermaid` → Sequence diagram of events
- `state --output mermaid` → State diagram with pending work
- `failures --group-by error --output mermaid` → Pie chart

## Debugging Workflow

1. Start with `workflow diagnose` to see the chain and root cause
2. Use `--output mermaid` to visualize if complex
3. Use `workflow failures --leaf-only` to see actual failures
4. Use `workflow describe --pending` to check pending work on running workflows
5. Use `--group-by` to find patterns in multiple failures

## Example Session

User: "The order workflow failed"

You should:
1. Run `temporal workflow describe --trace-root-cause --workflow-id order-123 --output json`
2. If complex, add `--output mermaid` for visual
3. Identify the leaf failure and root cause
4. Explain what went wrong
5. Suggest a fix

