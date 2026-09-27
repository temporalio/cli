#!/usr/bin/env bash

set -euo pipefail

demo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_dir="$(cd "${demo_dir}/.." && pwd)"
state_dir="${TMPDIR:-/tmp}/temporal-vts-schedule-demo-${UID}"

cli_bin="${demo_dir}/bin/temporal"
worker_bin="${demo_dir}/bin/worker"
server_pid_file="${state_dir}/server.pid"
worker_pid_file="${state_dir}/worker.pid"
server_log="${state_dir}/server.log"
worker_log="${state_dir}/worker.log"
buffer_all_schedule_id_file="${state_dir}/buffer-all-schedule.id"
buffer_all_workflow_id_file="${state_dir}/buffer-all-workflow.id"
skip_schedule_id_file="${state_dir}/skip-schedule.id"
skip_workflow_id_file="${state_dir}/skip-workflow.id"
legacy_allow_all_schedule_id_file="${state_dir}/allow-all-schedule.id"
legacy_allow_all_workflow_id_file="${state_dir}/allow-all-workflow.id"
legacy_schedule_id_file="${state_dir}/schedule.id"
legacy_workflow_id_file="${state_dir}/workflow.id"
scheduler_version="${TEMPORAL_SCHEDULER_VERSION:-v2}"
dynamic_config_file="${TEMPORAL_DYNAMIC_CONFIG_FILE:-${demo_dir}/dynamicconfig-${scheduler_version}.yaml}"

temporal_port="${TEMPORAL_PORT:-7233}"
ui_port="${TEMPORAL_UI_PORT:-8233}"
temporal_address="${TEMPORAL_ADDRESS:-127.0.0.1:${temporal_port}}"
namespace="${TEMPORAL_NAMESPACE:-fx-test}"
task_queue="${TEMPORAL_TASK_QUEUE:-fx-test-task-queue}"
buffer_all_schedule_id="${TEMPORAL_BUFFER_ALL_SCHEDULE_ID:-}"
buffer_all_workflow_id="${TEMPORAL_BUFFER_ALL_WORKFLOW_ID:-}"
skip_schedule_id="${TEMPORAL_SKIP_SCHEDULE_ID:-}"
skip_workflow_id="${TEMPORAL_SKIP_WORKFLOW_ID:-}"
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
    if [[ "$command" != *"$expected_command"* ]]; then
      local executables
      executables="$(lsof -a -p "$pid" -d txt -Fn 2>/dev/null || true)"
      [[ "$executables" == *"n${expected_command}"* ]] || return 1
    fi
  fi
}

prepare_demo_identity() {
  local run_id
  local schedule_id_prefix
  local workflow_id_prefix
  run_id="${TEMPORAL_DEMO_RUN_ID:-$(date -u +%Y%m%dT%H%M%SZ)-$$-${RANDOM}}"
  schedule_id_prefix="${TEMPORAL_SCHEDULE_ID:-one-hour-timer-${scheduler_version}-${run_id}}"
  workflow_id_prefix="${TEMPORAL_WORKFLOW_ID:-one-hour-timer-workflow-${scheduler_version}-${run_id}}"
  buffer_all_schedule_id="${TEMPORAL_BUFFER_ALL_SCHEDULE_ID:-${schedule_id_prefix}-buffer-all}"
  buffer_all_workflow_id="${TEMPORAL_BUFFER_ALL_WORKFLOW_ID:-${workflow_id_prefix}-buffer-all}"
  skip_schedule_id="${TEMPORAL_SKIP_SCHEDULE_ID:-${schedule_id_prefix}-skip}"
  skip_workflow_id="${TEMPORAL_SKIP_WORKFLOW_ID:-${workflow_id_prefix}-skip}"
  printf '%s\n' "$buffer_all_schedule_id" >"$buffer_all_schedule_id_file"
  printf '%s\n' "$buffer_all_workflow_id" >"$buffer_all_workflow_id_file"
  printf '%s\n' "$skip_schedule_id" >"$skip_schedule_id_file"
  printf '%s\n' "$skip_workflow_id" >"$skip_workflow_id_file"
}

load_demo_identity() {
  if [[ -f "$buffer_all_schedule_id_file" && -f "$buffer_all_workflow_id_file" && \
    -f "$skip_schedule_id_file" && -f "$skip_workflow_id_file" ]]; then
    buffer_all_schedule_id="$(<"$buffer_all_schedule_id_file")"
    buffer_all_workflow_id="$(<"$buffer_all_workflow_id_file")"
    skip_schedule_id="$(<"$skip_schedule_id_file")"
    skip_workflow_id="$(<"$skip_workflow_id_file")"
    return
  fi
  if [[ -n "$buffer_all_schedule_id" && -n "$buffer_all_workflow_id" && \
    -n "$skip_schedule_id" && -n "$skip_workflow_id" ]]; then
    return
  fi
  printf 'No active demo identity. Run "make demo-start" (V2) or "make demo-start-v1" first.\n' >&2
  exit 1
}

delete_demo_schedule() {
  local id_file="$1"
  [[ -f "$id_file" ]] || return 0
  local id
  id="$(<"$id_file")"
  "$cli_bin" schedule delete \
    --address "$temporal_address" \
    --namespace "$namespace" \
    --schedule-id "$id" >/dev/null 2>&1 || true
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

  for _ in {1..20}; do
    if ! kill -0 "$pid" 2>/dev/null; then
      rm -f "$pid_file"
      return
    fi
    sleep 0.25
  done

  kill -KILL "$pid" 2>/dev/null || true
  rm -f "$pid_file"
}

demo_cleanup() {
  if pid_is_running "$server_pid_file" "$cli_bin"; then
    delete_demo_schedule "$buffer_all_schedule_id_file"
    delete_demo_schedule "$skip_schedule_id_file"
    delete_demo_schedule "$legacy_allow_all_schedule_id_file"
    delete_demo_schedule "$legacy_schedule_id_file"
  fi
  stop_process "Worker" "$worker_pid_file" "$worker_bin"
  stop_process "Temporal server" "$server_pid_file" "$cli_bin"
  rm -f \
    "$server_log" \
    "$worker_log" \
    "$buffer_all_schedule_id_file" \
    "$buffer_all_workflow_id_file" \
    "$skip_schedule_id_file" \
    "$skip_workflow_id_file" \
    "$legacy_allow_all_schedule_id_file" \
    "$legacy_allow_all_workflow_id_file" \
    "$legacy_schedule_id_file" \
    "$legacy_workflow_id_file"
  rmdir "$state_dir" 2>/dev/null || true
  printf 'Cleanup complete.\n'
}

create_demo_schedule() {
  local schedule_id="$1"
  local workflow_id="$2"
  local overlap_policy="$3"
  local -a time_skipping_args=()
  if [[ "$scheduler_version" == "v2" ]]; then
    time_skipping_args=(--fast-forward "$fast_forward")
    printf 'Creating hourly %s V2 Schedule %s with time skipping enabled (%s fast-forward)...\n' \
      "$overlap_policy" "$schedule_id" "$fast_forward"
  else
    printf 'Creating hourly %s V1 Schedule %s without time skipping...\n' \
      "$overlap_policy" "$schedule_id"
  fi
  "$cli_bin" schedule create \
    --address "$temporal_address" \
    --namespace "$namespace" \
    --schedule-id "$schedule_id" \
    --interval 1h \
    --overlap-policy "$overlap_policy" \
    --workflow-id "$workflow_id" \
    --task-queue "$task_queue" \
    --type OneHourTimerWorkflow \
    "${time_skipping_args[@]}"
}

fast_forward_demo_schedule() {
  local schedule_id="$1"
  local label="$2"
  printf 'Fast-forwarding %s Schedule by %s...\n' "$label" "$fast_forward"
  "$cli_bin" schedule update \
    --address "$temporal_address" \
    --namespace "$namespace" \
    --schedule-id "$schedule_id" \
    --fast-forward "$fast_forward"
}

disable_demo_schedule() {
  local schedule_id="$1"
  local label="$2"
  printf 'Disabling time skipping for %s Schedule...\n' "$label"
  "$cli_bin" schedule update \
    --address "$temporal_address" \
    --namespace "$namespace" \
    --schedule-id "$schedule_id" \
    --time-skipping disabled
}

wait_for_schedule_workflow() {
  local schedule_id="$1"
  local label="$2"
  local workflows
  for _ in {1..120}; do
    workflows="$("$cli_bin" workflow list \
      --address "$temporal_address" \
      --namespace "$namespace" \
      --query "TemporalScheduledById = \"${schedule_id}\"" \
      --output json 2>/dev/null || true)"
    if [[ "$workflows" == *'"workflowId"'* ]]; then
      printf 'Verified a real Workflow Execution for %s Schedule.\n' "$label"
      return
    fi
    sleep 0.5
  done

  printf 'No Workflow Execution started for %s Schedule within 60 seconds. See %s and %s\n' \
    "$label" "$server_log" "$worker_log" >&2
  exit 1
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

  local -a vts_server_args=()
  if [[ "$scheduler_version" == "v2" ]]; then
    vts_server_args=(--internal-principal-auth)
  fi

  printf 'Starting Temporal server and Web UI...\n'
  nohup "$cli_bin" --log-level "$log_level" server start-dev \
    --port "$temporal_port" \
    --ui-port "$ui_port" \
    --namespace "$namespace" \
    --dynamic-config-file "$dynamic_config_file" \
    "${vts_server_args[@]}" \
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

  prepare_demo_identity

  printf 'Starting the one-timer Workflow Worker...\n'
  nohup env \
    TEMPORAL_ADDRESS="$temporal_address" \
    TEMPORAL_NAMESPACE="$namespace" \
    TEMPORAL_TASK_QUEUE="$task_queue" \
    "$worker_bin" </dev/null >"$worker_log" 2>&1 &
  echo "$!" >"$worker_pid_file"

  sleep 1

  create_demo_schedule "$buffer_all_schedule_id" "$buffer_all_workflow_id" BufferAll
  create_demo_schedule "$skip_schedule_id" "$skip_workflow_id" Skip
  wait_for_schedule_workflow "$buffer_all_schedule_id" BufferAll
  wait_for_schedule_workflow "$skip_schedule_id" Skip

  trap - ERR INT TERM
  printf '\nDemo started.\n'
  printf 'BufferAll Schedule ID: %s\n' "$buffer_all_schedule_id"
  printf 'Skip Schedule ID: %s\n' "$skip_schedule_id"
  printf 'Web UI: http://localhost:%s\n' "$ui_port"
  printf 'BufferAll UI: http://localhost:%s/namespaces/%s/schedules/%s\n' "$ui_port" "$namespace" "$buffer_all_schedule_id"
  printf 'Skip UI: http://localhost:%s/namespaces/%s/schedules/%s\n' "$ui_port" "$namespace" "$skip_schedule_id"
  printf 'Describe after five seconds: make demo-describe\n'
  printf 'Cleanup: make demo-clean\n'
}

demo_fast_forward() {
  check_demo_build
  load_demo_identity
  fast_forward_demo_schedule "$buffer_all_schedule_id" BufferAll
  fast_forward_demo_schedule "$skip_schedule_id" Skip
}

demo_disable() {
  check_demo_build
  load_demo_identity
  disable_demo_schedule "$buffer_all_schedule_id" BufferAll
  disable_demo_schedule "$skip_schedule_id" Skip
}

demo_describe() {
  check_demo_build
  load_demo_identity
  if ! pid_is_running "$server_pid_file" "$cli_bin"; then
    printf 'The Temporal server is not running.\n' >&2
    exit 1
  fi

  sleep 5
  printf '=== BufferAll: %s ===\n' "$buffer_all_schedule_id"
  "$cli_bin" schedule describe \
    --address "$temporal_address" \
    --namespace "$namespace" \
    --schedule-id "$buffer_all_schedule_id" \
    --output json
  printf '\n=== Skip: %s ===\n' "$skip_schedule_id"
  "$cli_bin" schedule describe \
    --address "$temporal_address" \
    --namespace "$namespace" \
    --schedule-id "$skip_schedule_id" \
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

  if [[ -f "$buffer_all_schedule_id_file" ]]; then
    printf 'BufferAll Schedule ID: %s\n' "$(<"$buffer_all_schedule_id_file")"
    printf 'BufferAll UI: http://localhost:%s/namespaces/%s/schedules/%s\n' \
      "$ui_port" "$namespace" "$(<"$buffer_all_schedule_id_file")"
  fi
  if [[ -f "$skip_schedule_id_file" ]]; then
    printf 'Skip Schedule ID: %s\n' "$(<"$skip_schedule_id_file")"
    printf 'Skip UI: http://localhost:%s/namespaces/%s/schedules/%s\n' \
      "$ui_port" "$namespace" "$(<"$skip_schedule_id_file")"
  fi

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
  demo-fast-forward)
    demo_fast_forward
    ;;
  demo-disable)
    demo_disable
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
    printf 'Usage: %s {demo-start|demo-fast-forward|demo-disable|demo-describe|demo-status|demo-logs|demo-clean|server-start|server-status|server-logs|server-stop}\n' "$(basename "$0")" >&2
    exit 2
    ;;
esac
