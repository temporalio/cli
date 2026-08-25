# Local Schedule guide

Run these commands from `~/projects/cli/demo`. The build uses the local SDK and
server selected by the repository's `go.work` file.

## Automated demo

Choose the Scheduler implementation:

```bash
make demo-start-v1  # legacy Workflow-backed Scheduler: false / 0 / false
make demo-start-v2  # CHASM Scheduler: true / 100 / true
make demo-start     # alias for demo-start-v2
```

Manage either demo with:

```bash
make demo-status
make demo-describe
make demo-logs
make demo-clean
```

The Web UI is at <http://localhost:8233>. `demo-clean` stops the server and
Worker and removes the generated binaries, PID files, and logs.

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
  --overlap-policy AllowAll \
  --workflow-id my-scheduled-workflow \
  --task-queue my-task-queue \
  --type MyWorkflow
```

### 4. Enable time skipping and fast-forward

Schedule update performs a full replacement, so re-specify the Schedule fields:

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

Change the final value as needed: `--ff 30m`, `--ff 2h`, or `--ff 24h`.
`--fast-forward` is the long form of `--ff`.

Describe the result:

```bash
./bin/temporal schedule describe \
  --address 127.0.0.1:7233 \
  --namespace my-namespace \
  --schedule-id my-hourly-schedule
```
