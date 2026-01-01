package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"

	"github.com/temporalio/cli/examples/ai-research-agent-impl/shared"
	"github.com/temporalio/cli/examples/ai-research-agent-impl/workflow"
)

func main() {
	// Parse command line arguments
	question := flag.String("question", "What is the meaning of life?", "The research question to answer")
	flag.Parse()

	// Create the Temporal client
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create Temporal client:", err)
	}
	defer c.Close()

	// Set up the workflow options
	// WorkflowIDConflictPolicy terminates any existing workflow with the same ID
	// and starts a new run
	options := client.StartWorkflowOptions{
		ID:                       "research-workflow",
		TaskQueue:                shared.TaskQueue,
		WorkflowIDConflictPolicy: enums.WORKFLOW_ID_CONFLICT_POLICY_TERMINATE_EXISTING,
	}

	// Create the request
	req := shared.ResearchRequest{
		Question: *question,
	}

	// Start the workflow
	log.Println("Starting research workflow for question:", req.Question)
	we, err := c.ExecuteWorkflow(context.Background(), options, workflow.ResearchWorkflow, req)
	if err != nil {
		log.Fatalln("Unable to execute workflow:", err)
	}

	log.Println("Workflow started:", we.GetID(), we.GetRunID())

	// Wait for the workflow to complete
	var result shared.ResearchResult
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Workflow failed:", err)
	}

	fmt.Println("\n--- Research Result ---")
	fmt.Println("Question:", result.Question)
	fmt.Println("Answer:", result.Answer)
}
