package main

import (
	"log"
	"os"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func oneHourTimerWorkflow(ctx workflow.Context) error {
	return workflow.Sleep(ctx, time.Hour)
}

func main() {
	address := envOrDefault("TEMPORAL_ADDRESS", "127.0.0.1:7233")
	namespace := envOrDefault("TEMPORAL_NAMESPACE", "fx-test")
	taskQueue := envOrDefault("TEMPORAL_TASK_QUEUE", "fx-test-task-queue")

	c, err := client.Dial(client.Options{
		HostPort:  address,
		Namespace: namespace,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	w := worker.New(c, taskQueue, worker.Options{})
	w.RegisterWorkflowWithOptions(oneHourTimerWorkflow, workflow.RegisterOptions{
		Name: "OneHourTimerWorkflow",
	})
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatal(err)
	}
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
