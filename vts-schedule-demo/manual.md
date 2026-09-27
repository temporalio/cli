# Local Schedule guide

Run these commands from `~/projects/cli/vts-schedule-demo`. The build uses
`~/projects/sdk-go` and `~/projects/workspace2/temporal`, as selected by the
repository's `go.work` file.

## Automated demo

V2 CHASM is the default. Use the explicit V1 command only for the legacy
Workflow-backed Scheduler:

```bash
make demo-start-v1  # legacy Workflow-backed Scheduler: false / 0 / false
make demo-start     # CHASM Scheduler with time skipping enabled (default)
```

The default command creates both the `BufferAll` and `Skip` Schedules with
`--fast-forward 5h`, so their V2 time-skipping configuration is enabled at
creation. It also enables principal propagation and internal-principal
authentication so the CHASM invoker can start those Workflows. Schedule time
skipping is not applied by the V1 comparison command. Demo startup waits until
it observes a real Workflow Execution from each Schedule.

Manage either demo with:

```bash
make demo-status
make demo-describe
TEMPORAL_FAST_FORWARD=2h make demo-fast-forward
make demo-disable
make demo-logs
make demo-clean
```

The Web UI is at <http://localhost:8233>. `demo-clean` stops the server and
Worker, deletes the active Schedule, and removes generated binaries, PID files,
logs, and run metadata. Each start creates uniquely named `BufferAll` and `Skip`
Schedules; `demo-status` prints both IDs and direct Web UI links, while
`demo-describe` describes both. `BufferAll` queues every overlapping start;
`Skip` discards an overlapping start while the previous Workflow is running.

## Use your own namespace and application

### 1. Configure the namespace

Copy the V1 or V2 template and replace `fx-test` with your namespace:

```bash
cp dynamicconfig-v2.yaml my-dynamicconfig.yaml
```

Start only the server and Web UI; `--namespace` is supplied by the environment
and creates the namespace during startup:

```bash
TEMPORAL_NAMESPACE=my-namespace \
TEMPORAL_DYNAMIC_CONFIG_FILE="$PWD/my-dynamicconfig.yaml" \
make server-start
```

Server commands:

```bash
make server-status
make server-logs
make server-stop
```

Dynamic configuration is loaded at server startup. Restart after changing the
YAML. To add a namespace to an already-running server:

```bash
./bin/temporal operator namespace create \
  --address 127.0.0.1:7233 \
  --namespace my-namespace
```

### 2. Run your Worker

Your Worker must:

- Connect to `127.0.0.1:7233` in `my-namespace`.
- Register your Workflow, such as `MyWorkflow`.
- Poll the same task queue used by the Schedule, such as `my-task-queue`.

[`worker/main.go`](worker/main.go) is a minimal Go example. Run your Worker in a
separate terminal and leave it running.

### 3. Create a Schedule

Use the CLI directly:

```bash
./bin/temporal schedule create \
  --address 127.0.0.1:7233 \
  --namespace my-namespace \
  --schedule-id my-hourly-schedule \
  --interval 1h \
  --overlap-policy BufferAll \
  --workflow-id my-scheduled-workflow \
  --task-queue my-task-queue \
  --type MyWorkflow
```

### 4. Enable time skipping and fast-forward

Time-skipping-only updates preserve the existing Schedule configuration:

```bash
./bin/temporal schedule update \
  --address 127.0.0.1:7233 \
  --namespace my-namespace \
  --schedule-id my-hourly-schedule \
  --ff 5h
```

Change the final value as needed: `--ff 30m`, `--ff 2h`, or `--ff 24h`.
`--fast-forward` is the long form of `--ff`.
Use `--max-skip-count 100` only when you need to override the server default.

Disable time skipping without changing the rest of the Schedule:

```bash
./bin/temporal schedule update \
  --address 127.0.0.1:7233 \
  --namespace my-namespace \
  --schedule-id my-hourly-schedule \
  --time-skipping disabled
```

Describe the result:

```bash
./bin/temporal schedule describe \
  --address 127.0.0.1:7233 \
  --namespace my-namespace \
  --schedule-id my-hourly-schedule
```
