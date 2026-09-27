# Schedule time-skipping demo

This demo builds the Temporal CLI with its embedded development server from the
local Go workspace, starts the Web UI and a Worker, and creates two hourly V2
Schedules (`BufferAll` and `Skip`) with time skipping enabled and a five-hour
fast-forward.

The repository's `go.work` must include:

```text
use (
    .
    ../sdk-go
    ../workspace2/temporal
)
```

Run the demo:

```bash
cd vts-schedule-demo
make demo-start
TEMPORAL_FAST_FORWARD=2h make demo-fast-forward
make demo-status
make demo-describe
make demo-disable
make demo-logs
make demo-clean
```

V2 CHASM is the default. Use the explicit V1 command only when testing the
legacy Workflow-backed Scheduler:

```bash
make demo-start-v1  # legacy Workflow-backed Scheduler
make demo-start     # CHASM Scheduler (default)
```

`make demo-start` passes `--fast-forward` to both `schedule create` commands,
which enables time skipping on both V2 Schedules. Schedule time skipping is
V2-only, so `make demo-start-v1` creates the comparison Schedules without it.

The Web UI is available at <http://localhost:8233>. The development server uses
in-memory persistence, so `make demo-clean` stops the processes and removes all demo
server state. Every run generates and records unique Schedule and Workflow IDs
for both policies; `make demo-status` prints them and `make demo-describe`
automatically describes both Schedules.

`demo-fast-forward` uses a time-skipping-only update, so it preserves each
Schedule's interval, action, and overlap policy. `--fast-forward` (alias `--ff`)
automatically enables time skipping. `demo-disable` uses the matching explicit
`--time-skipping disabled` update.

`BufferAll` queues every overlapping start and runs them one at a time after the
one-hour Workflow timer finishes. `Skip` discards starts that overlap a running
Workflow. The two Schedules make that policy difference visible without
creating simultaneous `AllowAll` actions that trigger the current Web UI's
duplicate-`actualTime` rendering bug.

For general usage with your own namespace, Worker, Workflow, and Schedule, see
[`manual.md`](manual.md). The `server-*` targets start only Temporal Server and
the Web UI; they do not start the demo Worker or create a Schedule.

V1 loads `dynamicconfig-v1.yaml` with CHASM creation and routing disabled and a
zero percent rollout for `fx-test`. V2 loads `dynamicconfig-v2.yaml`, which
enables CHASM Scheduler creation/routing, Workflow and Schedule time skipping,
and principal propagation only for `fx-test`. Registered server defaults remain
in effect for other namespaces. The V2 demo also enables development-server
internal-principal authentication, which the CHASM invoker requires when it
starts a Workflow with propagated time-skipping state. Startup verifies that
both Schedules create a real Workflow Execution before reporting success.

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
