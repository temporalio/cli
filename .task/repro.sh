#!/usr/bin/env bash
# Run: bash .task/repro.sh
#
# Builds the CLI from the currently checked-out commit and runs scenarios
# that demonstrate whether error reporting and user-facing warnings are
# coupled to the structured logger. Run on main to see the bug; run on
# the fix branch to see correct behavior.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

BINARY=/tmp/temporal-repro
ENV_FILE=$(mktemp)
STDERR_FILE=$(mktemp)
trap 'rm -f "$ENV_FILE" "$STDERR_FILE" "$BINARY"' EXIT

echo "Building temporal from $(git rev-parse --short HEAD) ($(git branch --show-current))..." >&2
go build -o "$BINARY" ./cmd/temporal

COMMIT="$(git rev-parse --short HEAD)"
BRANCH="$(git branch --show-current)"

run() {
    local stdout rc
    stdout=$("$BINARY" "$@" 2>"$STDERR_FILE") && rc=$? || rc=$?
    local stderr
    stderr=$(cat "$STDERR_FILE")
    echo '```bash'
    echo "$ temporal $*"
    echo '```'
    echo ""
    if [ -n "$stdout" ]; then
        echo "**stdout:**"
        echo '```'
        echo "$stdout"
        echo '```'
    fi
    echo "**stderr** (exit $rc):"
    if [ -z "$stderr" ]; then
        echo '```'
        echo "(empty)"
        echo '```'
    else
        echo '```'
        echo "$stderr"
        echo '```'
    fi
    echo ""
}

cat <<EOF
# Error reporting and logging: \`$BRANCH\` ($COMMIT)

## Setup

Build the CLI and create a test environment file:

\`\`\`bash
go build -o /tmp/temporal-repro ./cmd/temporal
ENV_FILE=\$(mktemp)
/tmp/temporal-repro env set --env-file \$ENV_FILE --env myenv -k foo -v bar 2>/dev/null
\`\`\`

EOF

# Seed the env file
"$BINARY" env set --env-file "$ENV_FILE" --env myenv -k foo -v bar 2>/dev/null

cat <<'EOF'
## Scenario 1: error reporting vs log level

Attempting `workflow list` against a closed port should produce a clear
error message on stderr regardless of `--log-level`. Expected: an
`Error: ...` message appears at every log level.

### Default log level
EOF
run workflow list --address 127.0.0.1:1

cat <<'EOF'
### `--log-level never`
EOF
run workflow list --address 127.0.0.1:1 --log-level never

cat <<'EOF'
---

## Scenario 2: deprecation warning vs log level

Using the deprecated positional-argument syntax for `env get` should
produce a plain-text warning on stderr regardless of `--log-level`.
Expected: a `Warning: ...` line appears at every log level, and it is
plain text (not a structured log message with `time=`/`level=` prefixes).

### Default log level
EOF
run env get --env-file "$ENV_FILE" myenv

cat <<'EOF'
### `--log-level never`
EOF
run env get --env-file "$ENV_FILE" --log-level never myenv

cat <<'EOF'
---

## Scenario 3: default log level noise

Running `env set` at the default log level should not dump structured
log lines to stderr. Expected: stderr is empty.

### Default log level
EOF
rm -f "$ENV_FILE" && touch "$ENV_FILE"
run env set --env-file "$ENV_FILE" --env myenv -k foo -v bar

cat <<'EOF'
### `--log-level info` (opt-in)
EOF
rm -f "$ENV_FILE" && touch "$ENV_FILE"
run env set --env-file "$ENV_FILE" --env myenv -k foo -v bar --log-level info

cat <<'EOF'
---

## Cleanup

```bash
rm -f /tmp/temporal-repro $ENV_FILE
```
EOF
