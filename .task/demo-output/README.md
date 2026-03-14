# Standalone Activity

This sample shows how to execute Activities directly from a Temporal Client, without a Workflow.

For full documentation, see [Standalone Activities - Python SDK](https://docs.temporal.io/develop/python/standalone-activities).

### Sample directory structure

- [my_activity.py](./my_activity.py) - Activity definition with `@activity.defn`
- [worker.py](./worker.py) - Worker that registers and runs the Activity
- [execute_activity.py](./execute_activity.py) - Execute a Standalone Activity and wait for the result
- [start_activity.py](./start_activity.py) - Start a Standalone Activity, get a handle, then wait for the result
- [list_activities.py](./list_activities.py) - List Standalone Activity Executions
- [count_activities.py](./count_activities.py) - Count Standalone Activity Executions

### Quickstart

**1. Start the Temporal dev server**

```bash
temporal server start-dev
```

**2. Run the Worker** (in a separate terminal)

```bash
uv run hello_standalone_activity/worker.py
```

**3. Execute a Standalone Activity** (in a separate terminal)

Execute and wait for the result:

```bash
uv run hello_standalone_activity/execute_activity.py
```

Or use the Temporal CLI:

```bash
temporal activity execute \
  --type compose_greeting \
  --activity-id my-standalone-activity-id \
  --task-queue my-standalone-activity-task-queue \
  --start-to-close-timeout 10s \
  --input '{"greeting": "Hello", "name": "World"}'
```

**4. Start a Standalone Activity (without waiting)**

Start, get a handle, then wait for the result:

```bash
uv run hello_standalone_activity/start_activity.py
```

Or use the Temporal CLI:

```bash
temporal activity start \
  --type compose_greeting \
  --activity-id my-standalone-activity-id \
  --task-queue my-standalone-activity-task-queue \
  --start-to-close-timeout 10s \
  --input '{"greeting": "Hello", "name": "World"}'
```

**5. List Standalone Activities**

```bash
uv run hello_standalone_activity/list_activities.py
```

Or use the Temporal CLI:

```bash
temporal activity list --query "TaskQueue = 'my-standalone-activity-task-queue'"
```

Note: `list` and `count` are only available in the [Standalone Activity prerelease CLI](https://github.com/temporalio/cli/releases/tag/v1.6.2-standalone-activity).

**6. Count Standalone Activities**

```bash
uv run hello_standalone_activity/count_activities.py
```

Or use the Temporal CLI:

```bash
temporal activity count --query "TaskQueue = 'my-standalone-activity-task-queue'"
```

### Temporal Cloud

The same code works against Temporal Cloud - just set environment variables. No code changes needed.

**Connect with mTLS:**

```bash
export TEMPORAL_ADDRESS=<your-namespace>.<your-account-id>.tmprl.cloud:7233
export TEMPORAL_NAMESPACE=<your-namespace>.<your-account-id>
export TEMPORAL_TLS_CLIENT_CERT_PATH='path/to/your/client.pem'
export TEMPORAL_TLS_CLIENT_KEY_PATH='path/to/your/client.key'
```

**Connect with an API key:**

```bash
export TEMPORAL_ADDRESS=<region>.<cloud_provider>.api.temporal.io:7233
export TEMPORAL_NAMESPACE=<your-namespace>.<your-account-id>
export TEMPORAL_API_KEY=<your-api-key>
```

Then run the worker and starter as shown above.
