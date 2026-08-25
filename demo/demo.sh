#!/usr/bin/env bash

set -euo pipefail

demo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_dir="$(cd "${demo_dir}/.." && pwd)"
state_dir="${TMPDIR:-/tmp}/temporal-time-skipping-demo-${UID}"

cli_bin="${demo_dir}/bin/temporal"
worker_bin="${demo_dir}/bin/worker"
server_pid_file="${state_dir}/server.pid"
worker_pid_file="${state_dir}/worker.pid"
server_log="${state_dir}/server.log"
worker_log="${state_dir}/worker.log"
scheduler_version="${TEMPORAL_SCHEDULER_VERSION:-v2}"
dynamic_config_file="${TEMPORAL_DYNAMIC_CONFIG_FILE:-${demo_dir}/dynamicconfig-${scheduler_version}.yaml}"

temporal_port="${TEMPORAL_PORT:-7233}"
ui_port="${TEMPORAL_UI_PORT:-8233}"
temporal_address="${TEMPORAL_ADDRESS:-127.0.0.1:${temporal_port}}"
namespace="${TEMPORAL_NAMESPACE:-fx-test}"
task_queue="${TEMPORAL_TASK_QUEUE:-fx-test-task-queue}"
schedule_id="${TEMPORAL_SCHEDULE_ID:-one-hour-timer}"
workflow_id="${TEMPORAL_WORKFLOW_ID:-one-hour-timer-workflow}"
fast_forward="${TEMPORAL_FAST_FORWARD:-5h}"
log_level="${TEMPORAL_LOG_LEVEL:-warn}"

pid_is_running() {
  local pid_file="$1"
  local expected_command="${2:-}"
  [[ -f "$pid_file" ]] || return 1
  local pid
  pid="$(<"$pid_file")"
  [[ "$pid" =~ ^[0-9]+$ ]] || return 1
  kill -0 "$pid" 2>/dev/null || return 1

  if [[ -n "$expected_command" ]]; then
    local command
    command="$(ps -p "$pid" -o command= 2>/dev/null || true)"
    [[ "$command" == *"$expected_command"* ]] || return 1
  fi
}

stop_process() {
  local name="$1"
  local pid_file="$2"
  local expected_command="$3"

  if ! pid_is_running "$pid_file" "$expected_command"; then
    rm -f "$pid_file"
    return
  fi

  local pid
  pid="$(<"$pid_file")"
  printf 'Stopping %s (PID %s)...\n' "$name" "$pid"
  kill -INT "$pid" 2>/dev/null || true

  for _ in {1..20}; do
    if ! kill -0 "$pid" 2>/dev/null; then
      rm -f "$pid_file"
      return
    fi
    sleep 0.25
  done

  kill -TERM "$pid" 2>/dev/null || true
  rm -f "$pid_file"
}

demo_cleanup() {
  stop_process "Worker" "$worker_pid_file" "$worker_bin"
  stop_process "Temporal server" "$server_pid_file" "$cli_bin"
  rm -f "$server_log" "$worker_log"
  rmdir "$state_dir" 2>/dev/null || true
  printf 'Cleanup complete.\n'
}

check_workspace() {
  if [[ ! -f "${repo_dir}/go.work" ]]; then
    printf 'Missing %s/go.work. Configure the local SDK and server workspace first.\n' "$repo_dir" >&2
    exit 1
  fi
}

check_server_build() {
  check_workspace
  case "$scheduler_version" in
    v1 | v2) ;;
    *)
      printf 'Unsupported TEMPORAL_SCHEDULER_VERSION %q; expected v1 or v2.\n' "$scheduler_version" >&2
      exit 1
      ;;
  esac
  if [[ ! -x "$cli_bin" ]]; then
    printf 'The Temporal binary is missing. Run "make build-server" first.\n' >&2
    exit 1
  fi
  if [[ ! -f "$dynamic_config_file" ]]; then
    printf 'Dynamic configuration file not found: %s\n' "$dynamic_config_file" >&2
    exit 1
  fi
}

check_demo_build() {
  check_server_build
  if [[ ! -x "$worker_bin" ]]; then
    printf 'Demo binaries are missing. Run "make build" first.\n' >&2
    exit 1
  fi
}

wait_for_server() {
  for _ in {1..60}; do
    if "$cli_bin" operator namespace describe \
      --address "$temporal_address" \
      --namespace "$namespace" >/dev/null 2>&1; then
      return
    fi

    if ! pid_is_running "$server_pid_file" "$cli_bin"; then
      printf 'Temporal server exited during startup. See %s\n' "$server_log" >&2
      exit 1
    fi
    sleep 0.5
  done

  printf 'Timed out waiting for Temporal server. See %s\n' "$server_log" >&2
  exit 1
}

server_start() {
  check_server_build
  if pid_is_running "$server_pid_file" "$cli_bin"; then
    printf 'The Temporal server is already running. Run "make server-status" or "make server-stop".\n' >&2
    exit 1
  fi

  mkdir -p "$state_dir"

  printf 'Starting Temporal server and Web UI...\n'
  nohup "$cli_bin" --log-level "$log_level" server start-dev \
    --port "$temporal_port" \
    --ui-port "$ui_port" \
    --namespace "$namespace" \
    --dynamic-config-file "$dynamic_config_file" \
    --dynamic-config-value frontend.WorkflowTimeSkippingEnabled=true \
    --dynamic-config-value frontend.ScheduleTimeSkippingEnabled=true \
    </dev/null >"$server_log" 2>&1 &
  echo "$!" >"$server_pid_file"

  trap server_stop ERR INT TERM
  wait_for_server

  trap - ERR INT TERM
  printf '\nTemporal server started.\n'
  printf 'Namespace: %s\n' "$namespace"
  printf 'Scheduler configuration: %s\n' "$scheduler_version"
  printf 'Dynamic config: %s\n' "$dynamic_config_file"
  printf 'Web UI: http://localhost:%s\n' "$ui_port"
  printf 'Server log: %s\n' "$server_log"
}

server_stop() {
  stop_process "Temporal server" "$server_pid_file" "$cli_bin"
  rm -f "$server_log"
  rmdir "$state_dir" 2>/dev/null || true
  printf 'Temporal server stopped.\n'
}

demo_start() {
  check_demo_build
  if pid_is_running "$server_pid_file" "$cli_bin" || pid_is_running "$worker_pid_file" "$worker_bin"; then
    printf 'The demo is already running. Run "make demo-status" or "make demo-clean".\n' >&2
    exit 1
  fi

  server_start
  trap demo_cleanup ERR INT TERM

  printf 'Starting the one-timer Workflow Worker...\n'
  nohup env \
    TEMPORAL_ADDRESS="$temporal_address" \
    TEMPORAL_NAMESPACE="$namespace" \
    TEMPORAL_TASK_QUEUE="$task_queue" \
    "$worker_bin" </dev/null >"$worker_log" 2>&1 &
  echo "$!" >"$worker_pid_file"

  sleep 1

  printf 'Creating the hourly Schedule...\n'
  "$cli_bin" schedule create \
    --address "$temporal_address" \
    --namespace "$namespace" \
    --schedule-id "$schedule_id" \
    --interval 1h \
    --overlap-policy AllowAll \
    --workflow-id "$workflow_id" \
    --task-queue "$task_queue" \
    --type OneHourTimerWorkflow

  printf 'Enabling time skipping and fast-forwarding by %s...\n' "$fast_forward"
  "$cli_bin" schedule update \
    --address "$temporal_address" \
    --namespace "$namespace" \
    --schedule-id "$schedule_id" \
    --interval 1h \
    --overlap-policy AllowAll \
    --workflow-id "$workflow_id" \
    --task-queue "$task_queue" \
    --type OneHourTimerWorkflow \
    --fast-forward "$fast_forward"

  trap - ERR INT TERM
  printf '\nDemo started.\n'
  printf 'Web UI: http://localhost:%s\n' "$ui_port"
  printf 'Describe after five seconds: make demo-describe\n'
  printf 'Cleanup: make demo-clean\n'
}

demo_describe() {
  check_demo_build
  if ! pid_is_running "$server_pid_file" "$cli_bin"; then
    printf 'The Temporal server is not running.\n' >&2
    exit 1
  fi

  sleep 5
  "$cli_bin" schedule describe \
    --address "$temporal_address" \
    --namespace "$namespace" \
    --schedule-id "$schedule_id" \
    --output json
}

server_status() {
  if pid_is_running "$server_pid_file" "$cli_bin"; then
    printf 'Temporal server: running (PID %s)\n' "$(<"$server_pid_file")"
    printf 'Web UI: http://localhost:%s\n' "$ui_port"
  else
    printf 'Temporal server: stopped\n'
  fi

  printf 'Server log: %s\n' "$server_log"
}

demo_status() {
  server_status

  if pid_is_running "$worker_pid_file" "$worker_bin"; then
    printf 'Worker: running (PID %s)\n' "$(<"$worker_pid_file")"
  else
    printf 'Worker: stopped\n'
  fi

  printf 'Worker log: %s\n' "$worker_log"
}

server_logs() {
  mkdir -p "$state_dir"
  touch "$server_log"
  tail -n 50 -f "$server_log"
}

demo_logs() {
  mkdir -p "$state_dir"
  touch "$server_log" "$worker_log"
  tail -n 50 -f "$server_log" "$worker_log"
}

case "${1:-}" in
  demo-start)
    demo_start
    ;;
  demo-describe)
    demo_describe
    ;;
  demo-status)
    demo_status
    ;;
  demo-logs)
    demo_logs
    ;;
  demo-clean)
    demo_cleanup
    ;;
  server-start)
    server_start
    ;;
  server-status)
    server_status
    ;;
  server-logs)
    server_logs
    ;;
  server-stop)
    server_stop
    ;;
  *)
    printf 'Usage: %s {demo-start|demo-describe|demo-status|demo-logs|demo-clean|server-start|server-status|server-logs|server-stop}\n' "$(basename "$0")" >&2
    exit 2
    ;;
esac
