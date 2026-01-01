package temporalcli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/temporalio/cli/internal/workflowdebug"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

// printWorkflowOutput outputs a result in the specified format.
// For mermaid, it outputs the mermaid diagram.
// For json (or unspecified), it outputs properly formatted JSON.
func printWorkflowOutput(cctx *CommandContext, format string, result any, toMermaid func() string) error {
	if format == "mermaid" {
		cctx.Printer.Println(toMermaid())
		return nil
	}
	// Default to JSON output - format it properly
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}
	cctx.Printer.Println(string(data))
	return nil
}

// cliClientProvider implements workflowdebug.ClientProvider using the CLI's client options.
type cliClientProvider struct {
	cctx          *CommandContext
	clientOptions *ClientOptions
	clients       map[string]client.Client
	primaryClient client.Client
	primaryNS     string
}

func newCLIClientProvider(cctx *CommandContext, clientOptions *ClientOptions) (*cliClientProvider, error) {
	// Create the primary client first
	primaryClient, err := clientOptions.dialClient(cctx)
	if err != nil {
		return nil, err
	}

	return &cliClientProvider{
		cctx:          cctx,
		clientOptions: clientOptions,
		clients:       make(map[string]client.Client),
		primaryClient: primaryClient,
		primaryNS:     clientOptions.Namespace,
	}, nil
}

func (p *cliClientProvider) GetClient(ctx context.Context, namespace string) (client.Client, error) {
	// If requesting the primary namespace, return the primary client
	if namespace == p.primaryNS {
		return p.primaryClient, nil
	}

	// Check if we already have a client for this namespace
	if cl, ok := p.clients[namespace]; ok {
		return cl, nil
	}

	// For other namespaces, create a new client with the same connection options
	// but different namespace. We copy the ClientOptions and use dialClient.
	opts := *p.clientOptions
	opts.Namespace = namespace

	// Check for namespace-specific API key in environment
	// Format: TEMPORAL_API_KEY_<NAMESPACE> where namespace is normalized
	// (dots/dashes replaced with underscores, uppercased)
	nsEnvKey := normalizeNamespaceForEnv(namespace)
	if apiKey := os.Getenv("TEMPORAL_API_KEY_" + nsEnvKey); apiKey != "" {
		opts.ApiKey = apiKey
	}

	cl, err := opts.dialClient(p.cctx)
	if err != nil {
		return nil, err
	}

	p.clients[namespace] = cl
	return cl, nil
}

// normalizeNamespaceForEnv converts a namespace to an environment variable suffix.
// e.g., "moedash-finance-ns.temporal-dev" -> "MOEDASH_FINANCE_NS_TEMPORAL_DEV"
func normalizeNamespaceForEnv(namespace string) string {
	result := strings.ToUpper(namespace)
	result = strings.ReplaceAll(result, ".", "_")
	result = strings.ReplaceAll(result, "-", "_")
	return result
}

func (p *cliClientProvider) Close() {
	if p.primaryClient != nil {
		p.primaryClient.Close()
	}
	for _, cl := range p.clients {
		cl.Close()
	}
}

// TemporalWorkflowFailuresCommand - list recent workflow failures
func (c *TemporalWorkflowFailuresCommand) run(cctx *CommandContext, args []string) error {
	// Create client provider
	clientProvider, err := newCLIClientProvider(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer clientProvider.Close()

	// Parse statuses (supports both comma-separated and multiple flags)
	var statuses []enums.WorkflowExecutionStatus
	for _, s := range c.Status {
		// Split by comma in case user passes "Failed,TimedOut"
		for _, part := range strings.Split(s, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			status := workflowdebug.ParseWorkflowStatus(part)
			if status == enums.WORKFLOW_EXECUTION_STATUS_UNSPECIFIED {
				return fmt.Errorf("invalid status: %s", part)
			}
			statuses = append(statuses, status)
		}
	}

	// Build options
	opts := workflowdebug.FailuresOptions{
		Since:            c.Since.Duration(),
		Statuses:         statuses,
		FollowChildren:   c.FollowChildren,
		FollowNamespaces: c.FollowNamespaces,
		MaxDepth:         c.Depth,
		Limit:            c.Limit,
		ErrorContains:    c.ErrorContains,
		LeafOnly:         c.LeafOnly,
		CompactErrors:    c.CompactErrors,
		GroupBy:          c.GroupBy.Value,
	}

	// Add the main namespace to follow namespaces if following children
	if opts.FollowChildren && len(opts.FollowNamespaces) == 0 {
		opts.FollowNamespaces = []string{c.Parent.ClientOptions.Namespace}
	} else if opts.FollowChildren {
		// Ensure main namespace is included
		found := false
		for _, ns := range opts.FollowNamespaces {
			if ns == c.Parent.ClientOptions.Namespace {
				found = true
				break
			}
		}
		if !found {
			opts.FollowNamespaces = append([]string{c.Parent.ClientOptions.Namespace}, opts.FollowNamespaces...)
		}
	}

	// Find failures
	finder := workflowdebug.NewFailuresFinder(clientProvider, opts)
	result, err := finder.FindFailures(cctx, c.Parent.ClientOptions.Namespace)
	if err != nil {
		return fmt.Errorf("failed to find failures: %w", err)
	}

	// Output based on format
	return printWorkflowOutput(cctx, c.Format.Value, result, func() string {
		return workflowdebug.FailuresToMermaid(result)
	})
}

// TemporalWorkflowDiagnoseCommand - trace workflow to deepest failure
func (c *TemporalWorkflowDiagnoseCommand) run(cctx *CommandContext, args []string) error {
	// Create client provider
	clientProvider, err := newCLIClientProvider(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer clientProvider.Close()

	// Build options
	opts := workflowdebug.TraverserOptions{
		FollowNamespaces: c.FollowNamespaces,
		MaxDepth:         c.Depth,
	}

	// Add the main namespace to follow namespaces
	if len(opts.FollowNamespaces) == 0 {
		opts.FollowNamespaces = []string{c.Parent.ClientOptions.Namespace}
	} else {
		// Ensure main namespace is included
		found := false
		for _, ns := range opts.FollowNamespaces {
			if ns == c.Parent.ClientOptions.Namespace {
				found = true
				break
			}
		}
		if !found {
			opts.FollowNamespaces = append([]string{c.Parent.ClientOptions.Namespace}, opts.FollowNamespaces...)
		}
	}

	// Trace workflow
	traverser := workflowdebug.NewChainTraverser(clientProvider, opts)
	result, err := traverser.Trace(cctx, c.Parent.ClientOptions.Namespace, c.WorkflowId, c.RunId)
	if err != nil {
		return fmt.Errorf("failed to diagnose workflow: %w", err)
	}

	// Output based on format
	return printWorkflowOutput(cctx, c.Format.Value, result, func() string {
		return workflowdebug.TraceToMermaid(result)
	})
}

// TemporalToolSpecCommand - output tool specifications for AI frameworks
func (c *TemporalToolSpecCommand) run(cctx *CommandContext, _ []string) error {
	var output string
	var err error

	switch c.Format.Value {
	case "openai":
		output, err = workflowdebug.GetOpenAIToolSpecsJSON()
	case "claude":
		output, err = workflowdebug.GetClaudeToolSpecsJSON()
	case "langchain":
		output, err = workflowdebug.GetLangChainToolSpecsJSON()
	case "functions":
		output, err = workflowdebug.GetToolSpecsJSON()
	default:
		return fmt.Errorf("unknown format: %s", c.Format.Value)
	}

	if err != nil {
		return fmt.Errorf("failed to generate tool spec: %w", err)
	}

	cctx.Printer.Println(output)
	return nil
}
