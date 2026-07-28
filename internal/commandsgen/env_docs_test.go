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
      - name: env
        type: string
        description: Active environment name.
        default: default
        implied-env: TEMPORAL_ENV
      - name: env-file
        type: string
        description: Path to environment settings file.
        implied-env: TEMPORAL_ENV_FILE
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
		"Setting a shell environment variable is not the same as storing a preset",
		"/cli/command-reference/env",
		"## Special cases",
		"`TEMPORAL_GRPC_META_*`",
		"| `TEMPORAL_TLS_CERT` | `TEMPORAL_TLS_CLIENT_CERT_PATH` |",
		"### Legacy `temporal env` variables",
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

	// Legacy `temporal env` variables belong in their own section, not the main
	// table, so they are not read as peers of the profile-based settings.
	specialCases := strings.Index(got, "## Special cases")
	for _, legacy := range []string{"`TEMPORAL_ENV`", "`TEMPORAL_ENV_FILE`"} {
		idx := strings.Index(got, legacy)
		if idx == -1 {
			t.Errorf("missing %q in:\n%s", legacy, got)
		} else if idx < specialCases {
			t.Errorf("%s appears in the main table at %d, want it after ## Special cases at %d", legacy, idx, specialCases)
		}
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
