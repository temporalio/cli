package temporalcli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/temporalio/cli/internal/agent"
	"github.com/temporalio/cli/internal/printer"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

// cliClientProvider implements agent.ClientProvider using the CLI's client options.
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

func (c *TemporalAgentCommand) run(cctx *CommandContext, args []string) error {
	// Parent command should show help
	return c.Command.Help()
}

func (c *TemporalAgentFailuresCommand) run(cctx *CommandContext, args []string) error {
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
			status := agent.ParseWorkflowStatus(part)
			if status == enums.WORKFLOW_EXECUTION_STATUS_UNSPECIFIED {
				return fmt.Errorf("invalid status: %s", part)
			}
			statuses = append(statuses, status)
		}
	}

	// Build options
	opts := agent.FailuresOptions{
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
	finder := agent.NewFailuresFinder(clientProvider, opts)
	result, err := finder.FindFailures(cctx, c.Parent.ClientOptions.Namespace)
	if err != nil {
		return fmt.Errorf("failed to find failures: %w", err)
	}

	// Output based on format
	if c.Format.Value == "mermaid" {
		cctx.Printer.Println(agent.FailuresToMermaid(result))
		return nil
	}
	return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
}

func (c *TemporalAgentTraceCommand) run(cctx *CommandContext, args []string) error {
	// Create client provider
	clientProvider, err := newCLIClientProvider(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer clientProvider.Close()

	// Build options
	opts := agent.TraverserOptions{
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
	traverser := agent.NewChainTraverser(clientProvider, opts)
	result, err := traverser.Trace(cctx, c.Parent.ClientOptions.Namespace, c.WorkflowId, c.RunId)
	if err != nil {
		return fmt.Errorf("failed to trace workflow: %w", err)
	}

	// Output based on format
	if c.Format.Value == "mermaid" {
		cctx.Printer.Println(agent.TraceToMermaid(result))
		return nil
	}
	return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
}

func (c *TemporalAgentTimelineCommand) run(cctx *CommandContext, args []string) error {
	// Create client
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	// Build options
	opts := agent.TimelineOptions{
		Compact:           c.Compact,
		IncludePayloads:   c.IncludePayloads,
		EventTypes:        c.EventTypes,
		ExcludeEventTypes: c.ExcludeEventTypes,
	}

	// Generate timeline
	generator := agent.NewTimelineGenerator(cl, opts)
	result, err := generator.Generate(cctx, c.Parent.ClientOptions.Namespace, c.WorkflowId, c.RunId)
	if err != nil {
		return fmt.Errorf("failed to generate timeline: %w", err)
	}

	// Output based on format
	if c.Format.Value == "mermaid" {
		cctx.Printer.Println(agent.TimelineToMermaid(result))
		return nil
	}
	return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
}

func (c *TemporalAgentStateCommand) run(cctx *CommandContext, args []string) error {
	// Create client
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	// Build options
	opts := agent.StateOptions{
		IncludeDetails: c.IncludeDetails,
	}

	// Extract state
	extractor := agent.NewStateExtractor(cl, opts)
	result, err := extractor.GetState(cctx, c.Parent.ClientOptions.Namespace, c.WorkflowId, c.RunId)
	if err != nil {
		return fmt.Errorf("failed to get workflow state: %w", err)
	}

	// Output based on format
	if c.Format.Value == "mermaid" {
		cctx.Printer.Println(agent.StateToMermaid(result))
		return nil
	}
	return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
}

func (c *TemporalAgentToolSpecCommand) run(cctx *CommandContext, _ []string) error {
	var output string
	var err error

	switch c.Format.Value {
	case "openai":
		output, err = agent.GetOpenAIToolSpecsJSON()
	case "claude":
		output, err = agent.GetClaudeToolSpecsJSON()
	case "langchain":
		output, err = agent.GetLangChainToolSpecsJSON()
	case "functions":
		output, err = agent.GetToolSpecsJSON()
	default:
		return fmt.Errorf("unknown format: %s", c.Format.Value)
	}

	if err != nil {
		return fmt.Errorf("failed to generate tool spec: %w", err)
	}

	cctx.Printer.Println(output)
	return nil
}
