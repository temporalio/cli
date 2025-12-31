package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/cli/examples/ai-research-agent-impl/activity"
	"github.com/temporalio/cli/examples/ai-research-agent-impl/shared"
	"github.com/temporalio/cli/examples/ai-research-agent-impl/workflow"
)

func main() {
	// Create the Temporal client
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create Temporal client:", err)
	}
	defer c.Close()

	// Create a worker that listens on the task queue
	w := worker.New(c, shared.TaskQueue, worker.Options{})

	// Register workflow and activities
	w.RegisterWorkflow(workflow.ResearchWorkflow)
	w.RegisterActivity(activity.BreakdownQuestion)
	w.RegisterActivity(activity.ResearchSubQuestion)
	w.RegisterActivity(activity.SynthesizeAnswers)

	log.Println("Starting worker on task queue:", shared.TaskQueue)

	// Start listening to the task queue
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker:", err)
	}
}

