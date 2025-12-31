package agent

import (
	"fmt"
	"strings"
)

// MermaidFormat specifies the type of mermaid diagram to generate.
type MermaidFormat string

const (
	MermaidFormatFlowchart MermaidFormat = "flowchart"
	MermaidFormatSequence  MermaidFormat = "sequence"
	MermaidFormatGantt     MermaidFormat = "gantt"
)

// TraceToMermaid generates a mermaid flowchart from a TraceResult.
func TraceToMermaid(result *TraceResult) string {
	if result == nil || len(result.Chain) == 0 {
		return "graph TD\n    A[No workflows in chain]"
	}

	var sb strings.Builder
	sb.WriteString("graph TD\n")

	// Generate node definitions
	for i, node := range result.Chain {
		nodeID := fmt.Sprintf("W%d", i)
		label := formatNodeLabel(node)
		style := getNodeStyle(node)

		sb.WriteString(fmt.Sprintf("    %s[%s]%s\n", nodeID, label, style))
	}

	// Generate edges
	for i := 0; i < len(result.Chain)-1; i++ {
		fromID := fmt.Sprintf("W%d", i)
		toID := fmt.Sprintf("W%d", i+1)

		// Use different arrow for failed transitions
		if result.Chain[i+1].Status == "Failed" || result.Chain[i+1].Status == "TimedOut" {
			sb.WriteString(fmt.Sprintf("    %s -->|failed| %s\n", fromID, toID))
		} else {
			sb.WriteString(fmt.Sprintf("    %s --> %s\n", fromID, toID))
		}
	}

	// Add root cause node if present
	if result.RootCause != nil {
		rcID := "RC"
		lastNodeID := fmt.Sprintf("W%d", len(result.Chain)-1)
		rcLabel := formatRootCauseLabel(result.RootCause)
		sb.WriteString(fmt.Sprintf("    %s(((%s)))\n", rcID, rcLabel))
		sb.WriteString(fmt.Sprintf("    %s -.->|root cause| %s\n", lastNodeID, rcID))
		sb.WriteString(fmt.Sprintf("    style %s fill:#ff6b6b,stroke:#c92a2a,color:#fff\n", rcID))
	}

	return sb.String()
}

// TimelineToMermaid generates a mermaid sequence diagram from a TimelineResult.
func TimelineToMermaid(result *TimelineResult) string {
	if result == nil || len(result.Events) == 0 {
		return "sequenceDiagram\n    Note over Workflow: No events"
	}

	var sb strings.Builder
	sb.WriteString("sequenceDiagram\n")

	// Collect participants (unique activity types, child workflows, etc.)
	participants := make(map[string]bool)
	participants["Workflow"] = true

	for _, event := range result.Events {
		switch event.Category {
		case "activity":
			if event.Name != "" {
				participants[sanitizeParticipant(event.Name)] = true
			}
		case "child_workflow":
			if event.Name != "" {
				participants[sanitizeParticipant(event.Name)] = true
			}
		case "timer":
			participants["Timer"] = true
		case "nexus":
			if event.Name != "" {
				participants[sanitizeParticipant("Nexus:"+event.Name)] = true
			}
		}
	}

	// Declare participants
	sb.WriteString("    participant Workflow\n")
	for p := range participants {
		if p != "Workflow" {
			sb.WriteString(fmt.Sprintf("    participant %s\n", p))
		}
	}

	// Generate events
	for _, event := range result.Events {
		line := formatSequenceEvent(event)
		if line != "" {
			sb.WriteString(fmt.Sprintf("    %s\n", line))
		}
	}

	return sb.String()
}

// StateToMermaid generates a mermaid flowchart showing current workflow state.
func StateToMermaid(result *WorkflowStateResult) string {
	if result == nil {
		return "graph TD\n    A[No workflow state]"
	}

	var sb strings.Builder
	sb.WriteString("graph TD\n")

	// Main workflow node
	statusIcon := getStatusIcon(result.Status)
	wfLabel := fmt.Sprintf("%s %s<br/>%s", statusIcon, result.WorkflowType, result.Status)
	sb.WriteString(fmt.Sprintf("    WF[%s]\n", wfLabel))

	nodeCounter := 0

	// Pending activities
	if len(result.PendingActivities) > 0 {
		sb.WriteString("    subgraph Activities[\"Pending Activities\"]\n")
		for _, act := range result.PendingActivities {
			nodeID := fmt.Sprintf("A%d", nodeCounter)
			nodeCounter++
			label := fmt.Sprintf("%s<br/>attempt %d", act.ActivityType, act.Attempt)
			if act.LastFailure != "" {
				label += fmt.Sprintf("<br/>❌ %s", truncate(act.LastFailure, 30))
			}
			sb.WriteString(fmt.Sprintf("        %s[%s]\n", nodeID, label))
			sb.WriteString(fmt.Sprintf("    WF --> %s\n", nodeID))
		}
		sb.WriteString("    end\n")
	}

	// Pending child workflows
	if len(result.PendingChildWorkflows) > 0 {
		sb.WriteString("    subgraph Children[\"Pending Child Workflows\"]\n")
		for _, child := range result.PendingChildWorkflows {
			nodeID := fmt.Sprintf("C%d", nodeCounter)
			nodeCounter++
			label := fmt.Sprintf("%s<br/>%s", child.WorkflowType, truncate(child.WorkflowID, 20))
			sb.WriteString(fmt.Sprintf("        %s[%s]\n", nodeID, label))
			sb.WriteString(fmt.Sprintf("    WF --> %s\n", nodeID))
		}
		sb.WriteString("    end\n")
	}

	// Pending Nexus operations
	if len(result.PendingNexusOperations) > 0 {
		sb.WriteString("    subgraph Nexus[\"Pending Nexus Operations\"]\n")
		for _, nex := range result.PendingNexusOperations {
			nodeID := fmt.Sprintf("N%d", nodeCounter)
			nodeCounter++
			label := fmt.Sprintf("%s.%s<br/>%s", nex.Service, nex.Operation, nex.State)
			if nex.LastFailure != "" {
				label += fmt.Sprintf("<br/>❌ %s", truncate(nex.LastFailure, 30))
			}
			sb.WriteString(fmt.Sprintf("        %s[%s]\n", nodeID, label))
			sb.WriteString(fmt.Sprintf("    WF --> %s\n", nodeID))
		}
		sb.WriteString("    end\n")
	}

	// Style the main workflow node based on status
	switch result.Status {
	case "Running":
		sb.WriteString("    style WF fill:#74c0fc,stroke:#1c7ed6\n")
	case "Completed":
		sb.WriteString("    style WF fill:#8ce99a,stroke:#37b24d\n")
	case "Failed", "TimedOut":
		sb.WriteString("    style WF fill:#ff8787,stroke:#e03131\n")
	case "Canceled":
		sb.WriteString("    style WF fill:#ffd43b,stroke:#f59f00\n")
	}

	return sb.String()
}

// FailuresToMermaid generates a mermaid pie chart or flowchart from FailuresResult.
func FailuresToMermaid(result *FailuresResult) string {
	if result == nil || (len(result.Failures) == 0 && len(result.Groups) == 0) {
		return "graph TD\n    A[No failures found]"
	}

	// If grouped, show as pie chart
	if len(result.Groups) > 0 {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("pie title Failures by %s\n", result.GroupedBy))
		for _, group := range result.Groups {
			sb.WriteString(fmt.Sprintf("    %q : %d\n", truncate(group.Key, 30), group.Count))
		}
		return sb.String()
	}

	// Otherwise show as flowchart of failure chains
	var sb strings.Builder
	sb.WriteString("graph LR\n")

	for i, failure := range result.Failures {
		if i >= 10 { // Limit to 10 for readability
			sb.WriteString(fmt.Sprintf("    MORE[+%d more...]\n", len(result.Failures)-10))
			break
		}

		// Show chain as connected nodes
		chainID := fmt.Sprintf("F%d", i)
		rootLabel := truncate(failure.RootWorkflow.WorkflowID, 15)
		sb.WriteString(fmt.Sprintf("    %s_root[%s]\n", chainID, rootLabel))

		if failure.Depth > 0 && failure.LeafFailure != nil {
			leafLabel := truncate(failure.LeafFailure.WorkflowID, 15)
			sb.WriteString(fmt.Sprintf("    %s_leaf[%s]\n", chainID, leafLabel))
			sb.WriteString(fmt.Sprintf("    %s_root -->|depth %d| %s_leaf\n", chainID, failure.Depth, chainID))

			// Root cause
			rcLabel := truncate(failure.RootCause, 25)
			sb.WriteString(fmt.Sprintf("    %s_rc((%s))\n", chainID, rcLabel))
			sb.WriteString(fmt.Sprintf("    %s_leaf -.-> %s_rc\n", chainID, chainID))
		} else {
			rcLabel := truncate(failure.RootCause, 25)
			sb.WriteString(fmt.Sprintf("    %s_rc((%s))\n", chainID, rcLabel))
			sb.WriteString(fmt.Sprintf("    %s_root -.-> %s_rc\n", chainID, chainID))
		}
	}

	return sb.String()
}

// Helper functions

func formatNodeLabel(node WorkflowChainNode) string {
	statusIcon := getStatusIcon(node.Status)
	wfType := node.WorkflowType
	if wfType == "" {
		wfType = "Workflow"
	}
	label := fmt.Sprintf("%s %s<br/>%s", statusIcon, truncate(wfType, 20), node.Status)
	if node.IsLeaf {
		label += "<br/>🎯 LEAF"
	}
	return label
}

func formatRootCauseLabel(rc *RootCause) string {
	if rc.Activity != "" {
		return fmt.Sprintf("%s<br/>%s", rc.Activity, truncate(rc.Error, 30))
	}
	return truncate(rc.Error, 40)
}

func getNodeStyle(node WorkflowChainNode) string {
	switch node.Status {
	case "Failed", "TimedOut":
		return ":::failed"
	case "Completed":
		return ":::success"
	case "Running":
		return ":::running"
	default:
		return ""
	}
}

func getStatusIcon(status string) string {
	switch status {
	case "Running":
		return "🔄"
	case "Completed":
		return "✅"
	case "Failed":
		return "❌"
	case "TimedOut":
		return "⏱️"
	case "Canceled":
		return "🚫"
	default:
		return "❓"
	}
}

func formatSequenceEvent(event TimelineEvent) string {
	participant := "Workflow"
	if event.Name != "" {
		switch event.Category {
		case "activity":
			participant = sanitizeParticipant(event.Name)
		case "child_workflow":
			participant = sanitizeParticipant(event.Name)
		case "nexus":
			participant = sanitizeParticipant("Nexus:" + event.Name)
		case "timer":
			participant = "Timer"
		}
	}

	switch {
	case strings.Contains(event.Type, "Scheduled") || strings.Contains(event.Type, "Started"):
		if event.Category == "workflow" {
			return fmt.Sprintf("Note over Workflow: %s started", event.Type)
		}
		return fmt.Sprintf("Workflow->>+%s: Start", participant)

	case strings.Contains(event.Type, "Completed"):
		if event.Category == "workflow" {
			return "Note over Workflow: ✅ Completed"
		}
		return fmt.Sprintf("%s-->>-Workflow: ✅ Done", participant)

	case strings.Contains(event.Type, "Failed"):
		errMsg := truncate(event.Error, 30)
		if errMsg == "" {
			errMsg = "failed"
		}
		return fmt.Sprintf("%s--x Workflow: ❌ %s", participant, errMsg)

	case strings.Contains(event.Type, "TimedOut"):
		return fmt.Sprintf("%s--x Workflow: ⏱️ Timeout", participant)

	case strings.Contains(event.Type, "Fired"):
		if event.Category == "timer" {
			return fmt.Sprintf("Timer-->>Workflow: ⏰ %s", event.Name)
		}
		return ""

	case strings.Contains(event.Type, "Signaled"):
		return fmt.Sprintf("Note over Workflow: 📨 Signal: %s", event.Name)

	default:
		return ""
	}
}

func sanitizeParticipant(name string) string {
	// Mermaid participant names can't have spaces or special chars
	result := strings.ReplaceAll(name, " ", "_")
	result = strings.ReplaceAll(result, "-", "_")
	result = strings.ReplaceAll(result, ".", "_")
	result = strings.ReplaceAll(result, ":", "_")
	if len(result) > 20 {
		result = result[:20]
	}
	return result
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
