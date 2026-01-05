#!/bin/bash
# Ticket Drop Simulation Script
# Simulates a high-traffic ticket sale with configurable users and seats

set -e

# Configuration
USERS=${1:-100}
EVENT=${2:-"drop-$(date +%s)"}
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "╔════════════════════════════════════════════════════════════╗"
echo "║              TICKET DROP SIMULATION                        ║"
echo "╠════════════════════════════════════════════════════════════╣"
echo "║  Event:     $EVENT"
echo "║  Users:     $USERS"
echo "║  Seats:     10 (hardcoded in inventory)"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Check if Temporal is running
if ! temporal workflow list --limit 1 &>/dev/null; then
    echo "❌ Temporal server not running. Start it with:"
    echo "   temporal server start-dev"
    exit 1
fi

# Build binaries
echo "Building binaries..."
cd "$SCRIPT_DIR"
go build -o bin/queue-starter ./queue-starter
go build -o bin/starter ./starter
echo "✓ Build complete"
echo ""

# Start the queue
echo "Starting queue for event: $EVENT"
./bin/queue-starter --action start --event "$EVENT" 2>/dev/null || true
sleep 1

# Send all users at once
echo ""
echo "Sending $USERS users simultaneously..."
start_time=$(date +%s)

for i in $(seq 1 $USERS); do
    ./bin/queue-starter --action join --event "$EVENT" --user "user-$i" 2>/dev/null &
done

# Wait for all signals to be sent
wait
end_time=$(date +%s)
echo "✓ All $USERS join signals sent in $((end_time - start_time)) seconds"

# Monitor progress
echo ""
echo "Monitoring queue progress..."
QUEUE_WF="ticket-queue-$EVENT"

while true; do
    status=$(temporal workflow query --workflow-id "$QUEUE_WF" --type status -o json 2>/dev/null | jq -r '.queryResult[0]')
    active=$(echo "$status" | jq -r '.active_count')
    waiting=$(echo "$status" | jq -r '.queue_length')
    
    echo "  Active: $active | Waiting: $waiting"
    
    if [ "$active" = "0" ] && [ "$waiting" = "0" ]; then
        break
    fi
    sleep 2
done

echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║                      RESULTS                               ║"
echo "╠════════════════════════════════════════════════════════════╣"

# Count results
completed=$(temporal workflow list --query "WorkflowType = 'TicketPurchase' AND ExecutionStatus = 'Completed'" --limit 500 2>/dev/null | grep "$EVENT" | wc -l | tr -d ' ')
failed=$(temporal workflow list --query "WorkflowType = 'TicketPurchase' AND ExecutionStatus = 'Failed'" --limit 500 2>/dev/null | grep "$EVENT" | wc -l | tr -d ' ')

echo "║  ✅ Got tickets:    $completed"
echo "║  ❌ Sold out:       $failed"
echo "╠════════════════════════════════════════════════════════════╣"

# Check for issues
if [ "$completed" -le 10 ]; then
    echo "║  ✓ Correct: Only 10 seats were available                  ║"
else
    echo "║  ⚠ WARNING: More than 10 tickets issued!                  ║"
fi

total=$((completed + failed))
if [ "$total" -eq "$USERS" ]; then
    echo "║  ✓ All $USERS users processed                              ║"
else
    echo "║  ⚠ Only $total of $USERS users processed                   ║"
fi

echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "To investigate:"
echo "  temporal workflow list --query \"WorkflowType = 'TicketPurchase'\" | grep $EVENT"
echo "  temporal workflow query --workflow-id $QUEUE_WF --type status"

