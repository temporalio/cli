// Package agent provides structured, agent-optimized views of Temporal workflow
// execution data, designed for AI agents and automated tooling.
package agent

import "encoding/json"

// ToolSpec represents an OpenAI-compatible function/tool definition.
type ToolSpec struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  ToolParameters `json:"parameters"`
}

// ToolParameters defines the parameters for a tool.
type ToolParameters struct {
	Type       string                  `json:"type"`
	Properties map[string]ToolProperty `json:"properties"`
	Required   []string                `json:"required"`
}

// ToolProperty defines a single parameter property.
type ToolProperty struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
	Default     any      `json:"default,omitempty"`
}

// GetToolSpecs returns the OpenAI-compatible tool specifications for all agent tools.
func GetToolSpecs() []ToolSpec {
	return []ToolSpec{
		{
			Name: "find_recent_failures",
			Description: `Find recent workflow failures in a Temporal namespace.
Automatically traverses workflow chains to find the deepest failure (leaf failure) and root cause.
Returns structured JSON with failure reports including root_workflow, leaf_failure, depth, chain, and root_cause.
CLI command: temporal workflow failures --namespace <ns> --since 1h`,
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"namespace": {
						Type:        "string",
						Description: "The Temporal namespace to search for failures.",
					},
					"since": {
						Type:        "string",
						Description: "Time window to search for failures. Examples: '1h', '24h', '7d'. Default: '1h'.",
						Default:     "1h",
					},
					"status": {
						Type:        "string",
						Description: "Comma-separated list of statuses to filter by. Accepted values: Failed, TimedOut, Canceled, Terminated. Default: 'Failed,TimedOut'.",
						Default:     "Failed,TimedOut",
					},
					"follow_children": {
						Type:        "boolean",
						Description: "Whether to traverse child workflows to find leaf failures and root causes. Default: true.",
						Default:     true,
					},
					"error_contains": {
						Type:        "string",
						Description: "Filter failures to only those containing this substring in the error message (case-insensitive).",
					},
					"leaf_only": {
						Type:        "boolean",
						Description: "Show only leaf failures (workflows with no failing children). Filters out parent workflows that failed due to child failures, de-duplicating failure chains. Default: false.",
						Default:     false,
					},
					"compact_errors": {
						Type:        "boolean",
						Description: "Extract the core error message and strip wrapper context. Removes verbose details like workflow IDs, run IDs, and event IDs. Default: false.",
						Default:     false,
					},
					"group_by": {
						Type:        "string",
						Description: "Group failures by a field instead of listing individually. Returns aggregated counts and summaries per group.",
						Enum:        []string{"none", "type", "namespace", "status", "error"},
						Default:     "none",
					},
					"format": {
						Type:        "string",
						Description: "Output format. Use 'mermaid' to generate a visual diagram (pie chart when grouped, flowchart of failure chains otherwise). Default: 'json'.",
						Enum:        []string{"json", "mermaid"},
						Default:     "json",
					},
					"limit": {
						Type:        "integer",
						Description: "Maximum number of failures to return. Default: 50.",
						Default:     50,
					},
				},
				Required: []string{"namespace"},
			},
		},
		{
			Name: "trace_workflow_chain",
			Description: `Trace a workflow execution through its entire child workflow chain to find the deepest failure.
Identifies the leaf failure point and extracts root cause information.
Automates the manual process of: finding the workflow, inspecting children, following failing children until reaching the leaf, and extracting failure info.
CLI command: temporal workflow diagnose --workflow-id <id>`,
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"namespace": {
						Type:        "string",
						Description: "The Temporal namespace containing the workflow.",
					},
					"workflow_id": {
						Type:        "string",
						Description: "The workflow ID to trace.",
					},
					"run_id": {
						Type:        "string",
						Description: "Optional run ID. If not specified, uses the latest run.",
					},
					"max_depth": {
						Type:        "integer",
						Description: "Maximum depth to traverse when following child workflows. 0 means unlimited. Default: 0.",
						Default:     0,
					},
					"format": {
						Type:        "string",
						Description: "Output format. Use 'mermaid' to generate a visual flowchart showing the workflow chain. Default: 'json'.",
						Enum:        []string{"json", "mermaid"},
						Default:     "json",
					},
				},
				Required: []string{"namespace", "workflow_id"},
			},
		},
		{
			Name: "get_workflow_timeline",
			Description: `Get a compact event timeline for a workflow execution.
Returns structured JSON with workflow metadata and a list of events including timestamps, types, categories, and relevant details.
Categories include: workflow, activity, timer, child_workflow, signal, update, and other.
CLI command: temporal workflow show --workflow-id <id> --compact`,
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"namespace": {
						Type:        "string",
						Description: "The Temporal namespace containing the workflow.",
					},
					"workflow_id": {
						Type:        "string",
						Description: "The workflow ID to get timeline for.",
					},
					"run_id": {
						Type:        "string",
						Description: "Optional run ID. If not specified, uses the latest run.",
					},
					"compact": {
						Type:        "boolean",
						Description: "Whether to use compact mode (fewer events, key milestones only). Default: false.",
						Default:     false,
					},
					"include_payloads": {
						Type:        "boolean",
						Description: "Whether to include input/output payloads in events. Default: false.",
						Default:     false,
					},
					"format": {
						Type:        "string",
						Description: "Output format. Use 'mermaid' to generate a visual sequence diagram showing event flow. Default: 'json'.",
						Enum:        []string{"json", "mermaid"},
						Default:     "json",
					},
				},
				Required: []string{"namespace", "workflow_id"},
			},
		},
		{
			Name: "get_workflow_state",
			Description: `Get the current state of a workflow execution.
Returns information about pending activities, pending child workflows, pending Nexus operations, and workflow status.
Useful for understanding what a running workflow is currently doing or waiting for, including cross-namespace Nexus calls.
CLI command: temporal workflow describe --workflow-id <id> --pending`,
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"namespace": {
						Type:        "string",
						Description: "The Temporal namespace containing the workflow.",
					},
					"workflow_id": {
						Type:        "string",
						Description: "The workflow ID to get state for.",
					},
					"run_id": {
						Type:        "string",
						Description: "Optional run ID. If not specified, uses the latest run.",
					},
					"include_details": {
						Type:        "boolean",
						Description: "Include detailed information about pending items like memo and search attributes. Default: false.",
						Default:     false,
					},
					"format": {
						Type:        "string",
						Description: "Output format. Use 'mermaid' to generate a visual state diagram showing pending activities and child workflows. Default: 'json'.",
						Enum:        []string{"json", "mermaid"},
						Default:     "json",
					},
				},
				Required: []string{"namespace", "workflow_id"},
			},
		},
	}
}

// GetToolSpecsJSON returns the tool specifications as a JSON string.
func GetToolSpecsJSON() (string, error) {
	specs := GetToolSpecs()
	data, err := json.MarshalIndent(specs, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// OpenAIToolWrapper wraps a ToolSpec in the OpenAI tools format.
type OpenAIToolWrapper struct {
	Type     string   `json:"type"`
	Function ToolSpec `json:"function"`
}

// GetOpenAIToolSpecs returns the tool specifications in OpenAI's tools array format.
func GetOpenAIToolSpecs() []OpenAIToolWrapper {
	specs := GetToolSpecs()
	wrapped := make([]OpenAIToolWrapper, len(specs))
	for i, spec := range specs {
		wrapped[i] = OpenAIToolWrapper{
			Type:     "function",
			Function: spec,
		}
	}
	return wrapped
}

// GetOpenAIToolSpecsJSON returns the tool specifications as OpenAI-compatible JSON.
func GetOpenAIToolSpecsJSON() (string, error) {
	specs := GetOpenAIToolSpecs()
	data, err := json.MarshalIndent(specs, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// LangChainToolSpec represents a LangChain-compatible tool definition.
type LangChainToolSpec struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Args        ToolParameters `json:"args_schema"`
}

// GetLangChainToolSpecs returns the tool specifications in LangChain format.
func GetLangChainToolSpecs() []LangChainToolSpec {
	specs := GetToolSpecs()
	langchain := make([]LangChainToolSpec, len(specs))
	for i, spec := range specs {
		langchain[i] = LangChainToolSpec{
			Name:        spec.Name,
			Description: spec.Description,
			Args:        spec.Parameters,
		}
	}
	return langchain
}

// GetLangChainToolSpecsJSON returns the tool specifications as LangChain-compatible JSON.
func GetLangChainToolSpecsJSON() (string, error) {
	specs := GetLangChainToolSpecs()
	data, err := json.MarshalIndent(specs, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ClaudeToolSpec represents an Anthropic Claude-compatible tool definition.
type ClaudeToolSpec struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema ToolParameters `json:"input_schema"`
}

// GetClaudeToolSpecs returns the tool specifications in Anthropic Claude format.
func GetClaudeToolSpecs() []ClaudeToolSpec {
	specs := GetToolSpecs()
	claude := make([]ClaudeToolSpec, len(specs))
	for i, spec := range specs {
		claude[i] = ClaudeToolSpec{
			Name:        spec.Name,
			Description: spec.Description,
			InputSchema: spec.Parameters,
		}
	}
	return claude
}

// GetClaudeToolSpecsJSON returns the tool specifications as Anthropic Claude-compatible JSON.
func GetClaudeToolSpecsJSON() (string, error) {
	specs := GetClaudeToolSpecs()
	data, err := json.MarshalIndent(specs, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
