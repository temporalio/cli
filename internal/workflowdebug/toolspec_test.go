package workflowdebug

import (
	"encoding/json"
	"testing"
)

func TestGetToolSpecs(t *testing.T) {
	specs := GetToolSpecs()

	if len(specs) != 4 {
		t.Errorf("expected 4 tool specs, got %d", len(specs))
	}

	// Verify each tool has required fields
	expectedNames := []string{"find_recent_failures", "trace_workflow_chain", "get_workflow_timeline", "get_workflow_state"}
	for i, spec := range specs {
		if spec.Name != expectedNames[i] {
			t.Errorf("expected tool name %s, got %s", expectedNames[i], spec.Name)
		}
		if spec.Description == "" {
			t.Errorf("tool %s has empty description", spec.Name)
		}
		if spec.Parameters.Type != "object" {
			t.Errorf("tool %s parameters type should be 'object', got %s", spec.Name, spec.Parameters.Type)
		}
		if len(spec.Parameters.Properties) == 0 {
			t.Errorf("tool %s has no properties", spec.Name)
		}
		if len(spec.Parameters.Required) == 0 {
			t.Errorf("tool %s has no required fields", spec.Name)
		}
	}
}

func TestGetToolSpecsJSON(t *testing.T) {
	jsonStr, err := GetToolSpecsJSON()
	if err != nil {
		t.Fatalf("failed to get tool specs JSON: %v", err)
	}

	// Verify it's valid JSON
	var specs []ToolSpec
	if err := json.Unmarshal([]byte(jsonStr), &specs); err != nil {
		t.Fatalf("failed to unmarshal tool specs JSON: %v", err)
	}

	if len(specs) != 4 {
		t.Errorf("expected 4 specs, got %d", len(specs))
	}
}

func TestGetOpenAIToolSpecsJSON(t *testing.T) {
	jsonStr, err := GetOpenAIToolSpecsJSON()
	if err != nil {
		t.Fatalf("failed to get OpenAI tool specs JSON: %v", err)
	}

	// Verify it's valid JSON with correct structure
	var specs []OpenAIToolWrapper
	if err := json.Unmarshal([]byte(jsonStr), &specs); err != nil {
		t.Fatalf("failed to unmarshal OpenAI tool specs JSON: %v", err)
	}

	if len(specs) != 4 {
		t.Errorf("expected 4 specs, got %d", len(specs))
	}

	for _, spec := range specs {
		if spec.Type != "function" {
			t.Errorf("expected type 'function', got %s", spec.Type)
		}
		if spec.Function.Name == "" {
			t.Error("function name is empty")
		}
	}
}

func TestGetLangChainToolSpecsJSON(t *testing.T) {
	jsonStr, err := GetLangChainToolSpecsJSON()
	if err != nil {
		t.Fatalf("failed to get LangChain tool specs JSON: %v", err)
	}

	// Verify it's valid JSON with correct structure
	var specs []LangChainToolSpec
	if err := json.Unmarshal([]byte(jsonStr), &specs); err != nil {
		t.Fatalf("failed to unmarshal LangChain tool specs JSON: %v", err)
	}

	if len(specs) != 4 {
		t.Errorf("expected 4 specs, got %d", len(specs))
	}

	for _, spec := range specs {
		if spec.Name == "" {
			t.Error("tool name is empty")
		}
		if spec.Args.Type != "object" {
			t.Errorf("expected args type 'object', got %s", spec.Args.Type)
		}
	}
}

func TestGetClaudeToolSpecsJSON(t *testing.T) {
	jsonStr, err := GetClaudeToolSpecsJSON()
	if err != nil {
		t.Fatalf("failed to get Claude tool specs JSON: %v", err)
	}

	// Verify it's valid JSON with correct structure
	var specs []ClaudeToolSpec
	if err := json.Unmarshal([]byte(jsonStr), &specs); err != nil {
		t.Fatalf("failed to unmarshal Claude tool specs JSON: %v", err)
	}

	if len(specs) != 4 {
		t.Errorf("expected 4 specs, got %d", len(specs))
	}

	for _, spec := range specs {
		if spec.Name == "" {
			t.Error("tool name is empty")
		}
		if spec.InputSchema.Type != "object" {
			t.Errorf("expected input_schema type 'object', got %s", spec.InputSchema.Type)
		}
	}
}

func TestToolSpecContainsNamespaceRequired(t *testing.T) {
	specs := GetToolSpecs()

	for _, spec := range specs {
		hasNamespace := false
		for _, req := range spec.Parameters.Required {
			if req == "namespace" {
				hasNamespace = true
				break
			}
		}
		if !hasNamespace {
			t.Errorf("tool %s should require 'namespace' parameter", spec.Name)
		}
	}
}
