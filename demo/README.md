# Schedule time-skipping demo

This demo builds the Temporal CLI with its embedded development server from the
local Go workspace, starts the Web UI and a Worker, creates two hourly Schedules
(`BufferAll` and `Skip`), and fast-forwards both by five hours.

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

V2 CHASM is the default. Use the explicit V1 command only when testing the
legacy Workflow-backed Scheduler:

```bash
make demo-start-v1  # legacy Workflow-backed Scheduler
make demo-start     # CHASM Scheduler (default)
```

The Web UI is available at <http://localhost:8233>. The development server uses
in-memory persistence, so `make demo-clean` stops the processes and removes all demo
server state. Every run generates and records unique Schedule and Workflow IDs
for both policies; `make demo-status` prints them and `make demo-describe`
automatically describes both Schedules.

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
enables CHASM Scheduler creation and routing with a 100 percent rollout only for
`fx-test`. Registered server defaults remain in effect for other namespaces.
The existing time-skipping dynamic configuration values are still supplied by
`demo.sh`.

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
