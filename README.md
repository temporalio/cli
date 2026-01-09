# Temporal CLI

Temporal command-line interface and development server.

**[DOCUMENTATION](https://docs.temporal.io/cli)**

## Quick Install

Reference [the documentation](https://docs.temporal.io/cli) for detailed install information.

### Install via Homebrew

    brew install temporal

### Install via download

1. Download the version for your OS and architecture:
    - [Linux amd64](https://temporal.download/cli/archive/latest?platform=linux&arch=amd64)
    - [Linux arm64](https://temporal.download/cli/archive/latest?platform=linux&arch=arm64)
    - [macOS amd64](https://temporal.download/cli/archive/latest?platform=darwin&arch=amd64)
    - [macOS arm64](https://temporal.download/cli/archive/latest?platform=darwin&arch=arm64) (Apple silicon)
    - [Windows amd64](https://temporal.download/cli/archive/latest?platform=windows&arch=amd64)
2. Extract the downloaded archive.
3. Add the `temporal` binary to your `PATH` (`temporal.exe` for Windows).

### Run via Docker

[Temporal CLI on DockerHub](https://hub.docker.com/r/temporalio/temporal)

    docker run --rm temporalio/temporal --help

Note that for dev server to be accessible from host system, it needs to listen on external IP and the ports need to be forwarded:

    docker run --rm -p 7233:7233 -p 8233:8233 temporalio/temporal:latest server start-dev --ip 0.0.0.0
    # UI is now accessible from host at http://localhost:8233/

### Build

1. Install [Go](https://go.dev/)
2. Clone repository
3. Switch to cloned directory, and run `go build ./cmd/temporal`

The executable will be at `temporal` (`temporal.exe` for Windows).

## Usage

Reference [the documentation](https://docs.temporal.io/cli) for detailed usage information.

## AI-Optimized Debugging Commands

The CLI includes workflow commands optimized for AI agents, LLM tooling, and automated debugging:

### Enhanced Commands

- **`temporal workflow list --failed`** - List recent workflow failures with auto-traversed root cause
- **`temporal workflow describe --trace-root-cause`** - Trace a workflow through its child chain to the deepest failure
- **`temporal workflow show --compact`** - Show a compact event timeline
- **`temporal workflow show --output mermaid`** - Generate a sequence diagram
- **`temporal workflow describe --pending`** - Show pending activities, children, and Nexus operations
- **`temporal workflow describe --output mermaid`** - Generate a state diagram
- **`temporal tool-spec`** - Output tool specifications for AI agent frameworks

### Examples

```bash
# List failures from the last hour with automatic chain traversal
temporal workflow list --failed --namespace prod --since 1h --follow-children

# Filter failures by error message (case-insensitive)
temporal workflow list --failed --namespace prod --since 1h --error-contains "timeout"

# Show only leaf failures (de-duplicate parent/child chains)
temporal workflow list --failed --namespace prod --since 1h --follow-children --leaf-only

# Compact error messages (strip wrapper context, show core error)
temporal workflow list --failed --namespace prod --since 1h --follow-children --compact-errors

# Combine leaf-only and compact-errors for cleanest output
temporal workflow list --failed --namespace prod --since 1h --follow-children --leaf-only --compact-errors

# Group failures by error type for quick summary
temporal workflow list --failed --namespace prod --since 24h --follow-children --compact-errors --group-by error

# Group failures by namespace to see which services are failing
temporal workflow list --failed --namespace prod --since 24h --follow-children --group-by namespace

# Trace a workflow to find the deepest failure in the chain
temporal workflow describe --trace-root-cause --workflow-id order-123 --namespace prod

# Get a compact timeline of workflow events
temporal workflow show --workflow-id order-123 --namespace prod --compact

# Get current workflow state (pending activities, child workflows)
temporal workflow describe --workflow-id order-123 --namespace prod --pending

# Cross-namespace traversal (Nexus/child workflows in other namespaces)
TEMPORAL_API_KEY_FINANCE_NS="$FINANCE_KEY" \
temporal workflow describe --trace-root-cause --workflow-id order-123 --namespace commerce-ns \
  --follow-namespaces finance-ns
```

### Cross-Namespace Traversal

When tracing workflows that span multiple namespaces (via Nexus or child workflows), you can provide namespace-specific API keys using environment variables:

```bash
# Format: TEMPORAL_API_KEY_<NAMESPACE>
# Namespace names are normalized: dots/dashes → underscores, then UPPERCASED
#
# Examples of namespace → environment variable:
#   finance-ns              → TEMPORAL_API_KEY_FINANCE_NS
#   moedash-finance-ns      → TEMPORAL_API_KEY_MOEDASH_FINANCE_NS
#   finance.temporal-dev    → TEMPORAL_API_KEY_FINANCE_TEMPORAL_DEV

# Primary namespace uses TEMPORAL_API_KEY
export TEMPORAL_API_KEY="primary-ns-key"

# Additional namespaces use namespace-specific keys
export TEMPORAL_API_KEY_FINANCE_NS="finance-ns-key"
export TEMPORAL_API_KEY_LOGISTICS_NS="logistics-ns-key"

# Trace root cause across namespaces (follows Nexus operations and child workflows)
temporal workflow describe --trace-root-cause --workflow-id order-123 --namespace commerce-ns \
  --follow-namespaces finance-ns,logistics-ns

# List failures with cross-namespace traversal
temporal workflow list --failed --namespace commerce-ns --since 1h \
  --follow-children --follow-namespaces finance-ns,logistics-ns \
  --leaf-only --compact-errors
```

### Mermaid Visualization

Commands support `--output mermaid` to generate visual diagrams:

```bash
# Visualize workflow chain as a flowchart
temporal workflow describe --trace-root-cause --workflow-id order-123 --namespace prod --output mermaid

# Visualize timeline as a sequence diagram  
temporal workflow show --workflow-id order-123 --namespace prod --output mermaid

# Visualize current state with pending activities
temporal workflow describe --workflow-id order-123 --namespace prod --pending --output mermaid

# Visualize failures as a pie chart (when grouped)
temporal workflow list --failed --namespace prod --since 1h --group-by error --output mermaid

# Visualize failures as a flowchart (when not grouped)
temporal workflow list --failed --namespace prod --since 1h --follow-children --output mermaid
```

The mermaid output renders directly in:
- Cursor AI and VS Code with Mermaid extension
- GitHub markdown files and comments
- Notion pages
- Any markdown preview with Mermaid support

Example diagnose output with `--output mermaid`:
```
graph TD
    W0[🔄 OrderWorkflow<br/>Failed]
    W1[❌ PaymentWorkflow<br/>Failed<br/>🎯 LEAF]
    W0 -->|failed| W1
    RC(((connection timeout)))
    W1 -.->|root cause| RC
    style RC fill:#ff6b6b,stroke:#c92a2a,color:#fff
```

### AI Agent Integration

The `temporal tool-spec` command outputs tool definitions compatible with AI agent frameworks:

```bash
# OpenAI function calling format (default)
temporal tool-spec --format openai

# Anthropic Claude format
temporal tool-spec --format claude

# LangChain tool format
temporal tool-spec --format langchain

# Raw function definitions
temporal tool-spec --format functions
```

These tool specs can be used to integrate Temporal debugging capabilities into AI agents, allowing them to:
- Query recent failures and their root causes
- Trace workflow chains to find the deepest failure
- Get compact workflow timelines

Example OpenAI integration:
```python
import subprocess
import json

# Get tool specs
result = subprocess.run(
    ["temporal", "tool-spec", "--format", "openai"],
    capture_output=True, text=True
)
tools = json.loads(result.stdout)

# Use with OpenAI API
response = openai.chat.completions.create(
    model="gpt-4",
    messages=[{"role": "user", "content": "Find recent failures in the prod namespace"}],
    tools=tools
)
```

Example Claude integration:
```python
import subprocess
import json
import anthropic

# Get tool specs
result = subprocess.run(
    ["temporal", "tool-spec", "--format", "claude"],
    capture_output=True, text=True
)
tools = json.loads(result.stdout)

# Use with Anthropic API
client = anthropic.Anthropic()
response = client.messages.create(
    model="claude-sonnet-4-20250514",
    max_tokens=1024,
    tools=tools,
    messages=[{"role": "user", "content": "Find recent failures in the prod namespace"}]
)
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).
