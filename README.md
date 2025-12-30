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

## Agent Commands

The `temporal agent` command group provides machine-readable, structured output optimized for AI agents, LLM tooling, and automated debugging workflows.

### Commands

- **`temporal agent failures`** - List recent workflow failures with auto-traversed root cause
- **`temporal agent trace`** - Trace a workflow through its child chain to the deepest failure  
- **`temporal agent timeline`** - Show a compact event timeline for a workflow
- **`temporal agent tool-spec`** - Output tool specifications for AI agent frameworks

### Examples

```bash
# List failures from the last hour with automatic chain traversal
temporal agent failures --namespace prod --since 1h --follow-children -o json

# Filter failures by error message (case-insensitive)
temporal agent failures --namespace prod --since 1h --error-contains "timeout" -o json

# Trace a workflow to find the deepest failure in the chain
temporal agent trace --workflow-id order-123 --namespace prod -o json

# Get a compact timeline of workflow events
temporal agent timeline --workflow-id order-123 --namespace prod --compact -o json
```

### Output

All agent commands output structured JSON designed for:
- Low token cost (compact, no redundant data)
- Easy parsing by LLMs and automated tools
- Derived views like `root_cause`, `leaf_failure`, and `chain`

Example trace output:
```json
{
  "chain": [
    {"namespace": "prod", "workflow_id": "order-123", "status": "Failed", "depth": 0},
    {"namespace": "prod", "workflow_id": "payment-456", "status": "Failed", "depth": 1, "leaf": true}
  ],
  "root_cause": {
    "type": "ActivityFailed",
    "activity": "ProcessPayment",
    "error": "connection timeout"
  },
  "depth": 1
}
```

### AI Agent Integration

The `temporal agent tool-spec` command outputs tool definitions compatible with AI agent frameworks:

```bash
# OpenAI function calling format (default)
temporal agent tool-spec --format openai

# Anthropic Claude format
temporal agent tool-spec --format claude

# LangChain tool format
temporal agent tool-spec --format langchain

# Raw function definitions
temporal agent tool-spec --format functions
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
    ["temporal", "agent", "tool-spec", "--format", "openai"],
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
    ["temporal", "agent", "tool-spec", "--format", "claude"],
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
