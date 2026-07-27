package commandsgen

import (
	"strings"
	"testing"
)

const envVarFixture = `
option-sets:
  - name: client
    options:
      - name: address
        type: string
        description: Temporal Service gRPC endpoint.
        default: localhost:7233
        implied-env: TEMPORAL_ADDRESS
      - name: namespace
        type: string
        description: Temporal Service Namespace.
        default: default
        implied-env: TEMPORAL_NAMESPACE
      - name: grpc-meta
        type: string[]
        description: |
          HTTP headers for requests.
commands:
  - name: workflow
    summary: Workflow
    description: Manage Workflows.
  - name: workflow list
    summary: List
    description: List Workflows.
    option-sets:
      - client
    options:
      - name: query
        type: string
        description: Visibility query.
        implied-env: TEMPORAL_QUERY
    docs:
      keywords:
        - workflow
      description-header: Manage Workflows
      tags:
        - Workflows
`

func TestGenerateEnvVarDocsFile(t *testing.T) {
	cmds, err := ParseCommands([]byte(envVarFixture))
	if err != nil {
		t.Fatalf("ParseCommands: %v", err)
	}

	docs := GenerateEnvVarDocsFile(cmds)
	if docs == nil {
		t.Fatal("expected env var docs, got nil")
	}
	got := string(docs)

	for _, want := range []string{
		"id: environment-variables",
		"| Environment variable | Flag | Description |",
		"| `TEMPORAL_ADDRESS` | `--address` |",
		"| `TEMPORAL_NAMESPACE` | `--namespace` |",
		"| `TEMPORAL_QUERY` | `--query` |",
		"Do not confuse these shell environment variables",
		"/cli/command-reference/env",
		"## Special cases",
		"`TEMPORAL_GRPC_META_*`",
		"| `TEMPORAL_TLS_CERT` | `TEMPORAL_TLS_CLIENT_CERT_PATH` |",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}

	// Alphabetized: ADDRESS before NAMESPACE before QUERY
	addr := strings.Index(got, "`TEMPORAL_ADDRESS`")
	ns := strings.Index(got, "`TEMPORAL_NAMESPACE`")
	query := strings.Index(got, "`TEMPORAL_QUERY`")
	if !(addr < ns && ns < query) {
		t.Errorf("expected alphabetical order ADDRESS < NAMESPACE < QUERY; got %d, %d, %d", addr, ns, query)
	}
}

func TestGenerateEnvVarDocsFileEmpty(t *testing.T) {
	cmds, err := ParseCommands([]byte(splitFixture))
	if err != nil {
		t.Fatalf("ParseCommands: %v", err)
	}
	if docs := GenerateEnvVarDocsFile(cmds); docs != nil {
		t.Fatalf("expected nil when no implied-env options, got:\n%s", docs)
	}
}

func TestGenerateDocsFilesIncludesEnvVarsWithoutSubdir(t *testing.T) {
	cmds, err := ParseCommands([]byte(envVarFixture))
	if err != nil {
		t.Fatalf("ParseCommands: %v", err)
	}

	docs, err := GenerateDocsFiles(cmds, nil)
	if err != nil {
		t.Fatalf("GenerateDocsFiles: %v", err)
	}
	if _, ok := docs["environment-variables"]; !ok {
		t.Fatalf("expected environment-variables page, got keys: %v", keys(docs))
	}

	docsSub, err := GenerateDocsFiles(cmds, []string{"workflow"})
	if err != nil {
		t.Fatalf("GenerateDocsFiles with subdir: %v", err)
	}
	if _, ok := docsSub["environment-variables"]; ok {
		t.Fatal("environment-variables must not be emitted when -subdir is set")
	}
}
