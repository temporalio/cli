#!/bin/bash
# TicketDrop Demo Workload - Infinite ticket sale simulation
# Creates continuous ticket drop events for demo purposes

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║         TICKETDROP DEMO WORKLOAD GENERATOR                 ║${NC}"
echo -e "${BLUE}║         Infinite ticket sales for showcase demo            ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check prerequisites
if ! command -v temporal &> /dev/null; then
    echo -e "${RED}❌ temporal CLI not found. Install it first.${NC}"
    exit 1
fi

if ! temporal workflow list --limit 1 &>/dev/null 2>&1; then
    echo -e "${RED}❌ Temporal server not running. Start it with:${NC}"
    echo "   temporal server start-dev"
    exit 1
fi

# Build binaries
rm -rf bin/worker bin/queue-starter bin/starter
echo -e "${YELLOW}Building binaries...${NC}"
go build -o bin/worker ./worker
go build -o bin/queue-starter ./queue-starter
go build -o bin/starter ./starter
echo -e "${GREEN}✓ Build complete${NC}"

# Start worker in background with nohup to prevent it from dying
echo -e "${YELLOW}Starting worker...${NC}"
nohup ./bin/worker > /tmp/ticketdrop-worker.log 2>&1 &
WORKER_PID=$!
sleep 2

# Verify worker started
if ps -p $WORKER_PID > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Worker started (PID: $WORKER_PID)${NC}"
    echo -e "${BLUE}  Log: /tmp/ticketdrop-worker.log${NC}"
else
    echo -e "${RED}❌ Worker failed to start. Check /tmp/ticketdrop-worker.log${NC}"
    cat /tmp/ticketdrop-worker.log
    exit 1
fi

cleanup() {
    echo ""
    echo -e "${YELLOW}Shutting down...${NC}"
    kill $WORKER_PID 2>/dev/null || true
    # Also kill any other workers that might be running
    pkill -f "bin/worker" 2>/dev/null || true
    exit 0
}
trap cleanup SIGINT SIGTERM EXIT

# Event counter
EVENT_NUM=1

echo ""
echo -e "${GREEN}Starting infinite workload loop...${NC}"
echo -e "${YELLOW}Press Ctrl+C to stop${NC}"
echo ""

while true; do
    EVENT_ID="concert-$(printf '%03d' $EVENT_NUM)"
    USERS=$((RANDOM % 20 + 15))  # 15-35 users per event
    
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}🎫 Event: $EVENT_ID | Users: $USERS | Seats: 10${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    # Start the queue
    ./bin/queue-starter --action start --event "$EVENT_ID" 2>/dev/null || true
    sleep 0.5
    
    # Send users (staggered slightly for realism)
    for i in $(seq 1 $USERS); do
        ./bin/queue-starter --action join --event "$EVENT_ID" --user "fan-$i" 2>/dev/null &
        # Small delay between users (simulates staggered arrival)
        if (( i % 5 == 0 )); then
            sleep 0.1
        fi
    done
    wait
    
    echo -e "  ${GREEN}✓ All $USERS users joined queue${NC}"
    
    # Wait for queue to drain (with timeout)
    QUEUE_WF="ticket-queue-$EVENT_ID"
    WAIT_COUNT=0
    MAX_WAIT=120
    
    while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
        # Check if worker is still running, restart if needed
        if ! ps -p $WORKER_PID > /dev/null 2>&1; then
            echo -e "  ${RED}⚠ Worker died, restarting...${NC}"
            nohup ./bin/worker >> /tmp/ticketdrop-worker.log 2>&1 &
            WORKER_PID=$!
            sleep 2
        fi
        
        status=$(temporal workflow query --workflow-id "$QUEUE_WF" --type status -o json 2>/dev/null | jq -r '.queryResult[0]' 2>/dev/null || echo '{}')
        active=$(echo "$status" | jq -r '.active_count // 0' 2>/dev/null || echo "0")
        waiting=$(echo "$status" | jq -r '.queue_length // 0' 2>/dev/null || echo "0")
        
        # Check if queue is done or workflow ended
        if [ "$active" = "0" ] && [ "$waiting" = "0" ]; then
            break
        fi
        if [ "$status" = "{}" ] || [ "$status" = "null" ] || [ -z "$status" ]; then
            echo -e "  ${YELLOW}Queue workflow ended${NC}"
            break
        fi
        
        echo -e "  Processing: ${YELLOW}Active=$active${NC} | ${BLUE}Waiting=$waiting${NC}"
        sleep 3
        WAIT_COUNT=$((WAIT_COUNT + 3))
    done
    
    if [ $WAIT_COUNT -ge $MAX_WAIT ]; then
        echo -e "  ${YELLOW}⚠ Timeout - moving to next event${NC}"
    fi
    
    # Quick results summary
    completed=$(temporal workflow list --query "WorkflowType = 'TicketPurchase' AND ExecutionStatus = 'Completed'" --limit 100 2>/dev/null | grep -c "$EVENT_ID" || echo "0")
    failed=$(temporal workflow list --query "WorkflowType = 'TicketPurchase' AND ExecutionStatus = 'Failed'" --limit 100 2>/dev/null | grep -c "$EVENT_ID" || echo "0")
    
    echo -e "  ${GREEN}✅ Tickets sold: $completed${NC} | ${RED}❌ Failed: $failed${NC}"
    
    EVENT_NUM=$((EVENT_NUM + 1))
    
    # Short pause between events
    echo ""
    sleep 3
done
