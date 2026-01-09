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

# Build if needed
if [ ! -f bin/worker ] || [ ! -f bin/queue-starter ]; then
    echo -e "${YELLOW}Building binaries...${NC}"
    go build -o bin/worker ./worker
    go build -o bin/queue-starter ./queue-starter
    go build -o bin/starter ./starter
    echo -e "${GREEN}✓ Build complete${NC}"
fi

# Start worker in background
echo -e "${YELLOW}Starting worker...${NC}"
./bin/worker &
WORKER_PID=$!
echo -e "${GREEN}✓ Worker started (PID: $WORKER_PID)${NC}"

cleanup() {
    echo ""
    echo -e "${YELLOW}Shutting down...${NC}"
    kill $WORKER_PID 2>/dev/null || true
    exit 0
}
trap cleanup SIGINT SIGTERM

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
    MAX_WAIT=60
    
    while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
        status=$(temporal workflow query --workflow-id "$QUEUE_WF" --type status -o json 2>/dev/null | jq -r '.queryResult[0]' 2>/dev/null || echo '{}')
        active=$(echo "$status" | jq -r '.active_count // 0')
        waiting=$(echo "$status" | jq -r '.queue_length // 0')
        
        if [ "$active" = "0" ] && [ "$waiting" = "0" ]; then
            break
        fi
        
        echo -e "  Processing: ${YELLOW}Active=$active${NC} | ${BLUE}Waiting=$waiting${NC}"
        sleep 2
        WAIT_COUNT=$((WAIT_COUNT + 2))
    done
    
    # Quick results summary
    completed=$(temporal workflow list --query "WorkflowType = 'TicketPurchase' AND ExecutionStatus = 'Completed'" --limit 100 2>/dev/null | grep -c "$EVENT_ID" || echo "0")
    failed=$(temporal workflow list --query "WorkflowType = 'TicketPurchase' AND ExecutionStatus = 'Failed'" --limit 100 2>/dev/null | grep -c "$EVENT_ID" || echo "0")
    
    echo -e "  ${GREEN}✅ Tickets sold: $completed${NC} | ${RED}❌ Failed: $failed${NC}"
    
    EVENT_NUM=$((EVENT_NUM + 1))
    
    # Short pause between events
    echo ""
    sleep 3
done
