# TicketDrop Live Demo

## 🎯 Demo Goal (1 minute)

Show an AI agent (Cursor/Claude) debugging a production issue in real-time:
1. **See failures happening** (workload running in background)
2. **Ask AI to investigate** (one prompt)
3. **Watch AI find root cause** (uses `temporal workflow diagnose`)
4. **AI suggests fix** (pinpoints the buggy code)

---

## 📋 Pre-Demo Setup (5 minutes before)

### Terminal 1: Start Temporal Server
```bash
temporal server start-dev
```

### Terminal 2: Introduce the Bug & Start Workload
```bash
cd examples/ticketdrop-impl

# Introduce the bug (swap activities file)
cp activities.go activities_good.go.bak
cp activities_buggy.go.template activities.go

# Rebuild and start infinite workload
chmod +x demo-workload.sh
./demo-workload.sh
```

You should see events cycling with failures:
```
🎫 Event: concert-001 | Users: 25 | Seats: 10
  ✓ All 25 users joined queue
  Processing: Active=10 | Waiting=15
  ✅ Tickets sold: 4 | ❌ Failed: 6   <-- Failures happening!
```

### Terminal 3: Open Cursor IDE
```bash
cd examples/ticketdrop-impl
cursor .
```

Keep the workload running in the background during the demo.

---

## 🎬 Live Demo Script (60 seconds)

### [0:00] Set the Scene
> "We have a ticket sales system. It's live, processing purchases. 
> But we're seeing failures. Let's ask AI to investigate."

*Show Terminal 2 with workload running and failures appearing*

### [0:10] Ask AI to Investigate

In Cursor, open a new chat and type:

```
We're seeing ticket purchase failures in production. 
Can you find out what's failing and why?
```

### [0:20] Watch AI Work

The AI will run:
```bash
temporal workflow failures --since 5m --follow-children --leaf-only --compact-errors
```

It will see output like:
```json
{
  "failures": [
    {
      "root_workflow": { "workflow_id": "purchase-concert-001-fan-6" },
      "root_cause": "ActivityFailed: ProcessPayment - payment gateway error: premium tier not configured",
      "depth": 0
    }
  ],
  "total_count": 12
}
```

### [0:35] AI Diagnoses the Pattern

AI might run:
```bash
temporal workflow failures --since 5m --group-by error
```

Output shows: **All failures are "premium tier not configured"**

### [0:45] AI Finds the Bug

The AI will search the codebase and find `isPremiumSeat()`:

```go
// 🐛 BUG: This incorrectly matches user IDs like "fan-6", "fan-7"
func isPremiumSeat(reservationID string) bool {
    parts := strings.Split(reservationID, "-")
    for _, part := range parts {
        if num, err := strconv.Atoi(part); err == nil {
            if num >= 6 && num <= 10 {
                return true // BUG: False positive!
            }
        }
    }
    return false
}
```

### [0:55] AI Suggests Fix

> "The `isPremiumSeat` function parses the reservation ID incorrectly. 
> It's matching user IDs (fan-6, fan-7...) instead of seat numbers.
> The fix is to check the actual seat number, not the reservation ID."

### [1:00] Done!

> "In 60 seconds, AI found the root cause across our distributed workflow,
> traced through the failure chain, identified the pattern, and found the bug.
> That's what Temporal + AI-native CLI gives you."

---

## 🛠 Post-Demo Cleanup

```bash
# Stop workload (Ctrl+C in Terminal 2)

# Restore good code
cd examples/ticketdrop-impl
cp activities_good.go.bak activities.go
rm activities_good.go.bak
```

---

## 💡 Demo Tips

### If AI doesn't use the CLI commands:
Prompt it: *"Use the temporal CLI to check for workflow failures"*

### If you want visual output:
Ask: *"Show me a visualization of the failures"*

The AI will run:
```bash
temporal workflow failures --since 5m --group-by error --format mermaid
```

### If you have more time (2-minute version):
Add: *"Now fix the bug"*

Watch AI:
1. Edit `activities.go`
2. Fix the `isPremiumSeat` function
3. Suggest rebuilding the worker

---

## 📊 Expected Failure Pattern

With the buggy code:
- **fan-1 through fan-5**: Usually succeed (unless random payment decline)
- **fan-6 through fan-10**: Always fail ("premium tier not configured")
- **fan-11+**: Succeed (no digits 6-10 in their ID)

This creates a consistent ~50% failure rate that's easy to diagnose.

---

## 🎯 Key Talking Points

1. **"One command to find all failures"** - No clicking through UI
2. **"Automatic root cause traversal"** - Follows child workflows automatically  
3. **"Pattern analysis built-in"** - `--group-by error` shows it's all the same bug
4. **"AI-readable output"** - JSON that AI can parse and reason about
5. **"Visual when you need it"** - `--format mermaid` for diagrams

---

## 🔧 Troubleshooting

### No failures appearing?
- Check worker is running: `ps aux | grep worker`
- Check Temporal server: `temporal workflow list`

### AI not finding the bug?
- Make sure `CLAUDE.md` is in the project root (Cursor rules)
- Check the cursor rules are loaded (Settings → Cursor Rules)

### Too many failures to read?
- Add `--limit 5` to reduce output
- Use `--group-by error` for summary view
