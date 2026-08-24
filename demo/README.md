# Schedule time-skipping demo

This demo builds the Temporal CLI with its embedded development server from the
local Go workspace, starts the Web UI and a Worker, creates an hourly Schedule,
and fast-forwards it by five hours.

The repository's `go.work` must include:

```text
use (
    .
    ../sdk-go
    ../temporal
)
```

Run the demo:

```bash
cd demo
make demo-start
make demo-status
make demo-describe
make demo-logs
make demo-clean
```

The Web UI is available at <http://localhost:8233>. The development server uses
in-memory persistence, so `make demo-clean` stops the processes and removes all demo
server state.

For general usage with your own namespace, Worker, Workflow, and Schedule, see
[`manual.md`](manual.md). The `server-*` targets start only Temporal Server and
the Web UI; they do not start the demo Worker or create a Schedule.

The server loads `dynamicconfig.yaml`, which enables CHASM Scheduler creation
and routing with a 100 percent rollout only for the `fx-test` namespace. The
registered server defaults remain in effect for other namespaces. The existing
time-skipping dynamic configuration values are still supplied by `demo.sh`.

Configuration can be overridden with environment variables:

```bash
TEMPORAL_FAST_FORWARD=10h TEMPORAL_UI_PORT=9233 make demo-start
```

Enable server debug logging when verifying scheduler routing:

```bash
TEMPORAL_LOG_LEVEL=debug make demo-start
```

Use a different dynamic configuration file with:

```bash
TEMPORAL_DYNAMIC_CONFIG_FILE=/path/to/dynamicconfig.yaml make demo-start
```
