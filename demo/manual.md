# Manual usage

This guide runs the locally built Temporal Server and Web UI, while leaving the
Worker, Workflow, and Schedule definitions under your control. Run commands from
`~/projects/cli/demo` unless noted otherwise.

## 1. Build the local CLI and server

The repository `go.work` selects the local SDK at `~/projects/sdk-go` and the
local server at `~/projects/temporal`.

```bash
make build-server
```

The resulting CLI and embedded server binary is `./bin/temporal`.

## 2. Configure and create a namespace

Copy `dynamicconfig.yaml` and replace every `fx-test` constraint with your
namespace:

```bash
cp dynamicconfig.yaml my-dynamicconfig.yaml
```

For example, for a namespace named `my-namespace`, the copied file should
contain:

```yaml
history.enableCHASMSchedulerCreation:
  - value: true
    constraints:
      namespace: my-namespace

history.chasmSchedulerCreationRolloutPercent:
  - value: 100
    constraints:
      namespace: my-namespace

history.enableCHASMSchedulerRouting:
  - value: true
    constraints:
      namespace: my-namespace
```

Start the server and create that namespace during startup:

```bash
TEMPORAL_NAMESPACE=my-namespace \
TEMPORAL_DYNAMIC_CONFIG_FILE="$PWD/my-dynamicconfig.yaml" \
make server-start
```

The server also enables Workflow and Schedule time skipping. Open the Web UI at
<http://localhost:8233>.

If the server is already running, create another namespace with:

```bash
./bin/temporal operator namespace create \
  --address 127.0.0.1:7233 \
  --namespace my-namespace
```

Dynamic configuration is loaded when this development server starts. Restart
the server after adding a namespace constraint to the YAML file.

Useful server commands:

```bash
make server-status
make server-logs
make server-stop
```

## 3. Run your Worker and register your Workflow

Your application Worker must connect to the same address and namespace, poll a
task queue, and register the Workflow type used by the Schedule. A minimal Go
Worker looks like this:

```go
package main

import (
    "log"
    "time"

    "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"
    "go.temporal.io/sdk/workflow"
)

func MyWorkflow(ctx workflow.Context) error {
    return workflow.Sleep(ctx, time.Hour)
}

func main() {
    c, err := client.Dial(client.Options{
        HostPort:  "127.0.0.1:7233",
        Namespace: "my-namespace",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer c.Close()

    w := worker.New(c, "my-task-queue", worker.Options{})
    w.RegisterWorkflow(MyWorkflow)
    if err := w.Run(worker.InterruptCh()); err != nil {
        log.Fatal(err)
    }
}
```

Run this Worker in its own terminal and leave it running. The Workflow type in
the CLI command must match the registered function name (`MyWorkflow`), and the
task queue must match `my-task-queue`.

## 4. Create a Schedule

Use the CLI directly for Schedule configuration:

```bash
./bin/temporal schedule create \
  --address 127.0.0.1:7233 \
  --namespace my-namespace \
  --schedule-id my-hourly-schedule \
  --interval 1h \
  --overlap-policy AllowAll \
  --workflow-id my-scheduled-workflow \
  --task-queue my-task-queue \
  --type MyWorkflow
```

Confirm it in the Web UI or describe it with:

```bash
./bin/temporal schedule describe \
  --address 127.0.0.1:7233 \
  --namespace my-namespace \
  --schedule-id my-hourly-schedule
```

## 5. Enable time skipping and choose a fast-forward duration

Schedule update is a full replacement. Re-specify the interval, Workflow,
task queue, and policy along with `--fast-forward` or its `--ff` alias:

```bash
./bin/temporal schedule update \
  --address 127.0.0.1:7233 \
  --namespace my-namespace \
  --schedule-id my-hourly-schedule \
  --interval 1h \
  --overlap-policy AllowAll \
  --workflow-id my-scheduled-workflow \
  --task-queue my-task-queue \
  --type MyWorkflow \
  --ff 5h
```

Use a different duration by changing the final value, for example:

```bash
--ff 30m
--ff 2h
--ff 24h
```

Describe the Schedule again to inspect `timeSkippingConfig`, action counts, and
completion state. Stop the local server when finished:

```bash
make server-stop
```

For the fully automated sample instead, use `make demo-start` and the other
`demo-*` targets documented in `README.md`.
