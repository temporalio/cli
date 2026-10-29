package main

import (
	"context"
	"log"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func TimerWorkflow(ctx workflow.Context) error {
	return workflow.Sleep(ctx, 10*time.Hour)
}

func FiveSecondActivity(ctx context.Context) error {
	select {
	case <-time.After(5 * time.Second):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func ActivityTimerActivityWorkflow(ctx workflow.Context) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	})

	if err := workflow.ExecuteActivity(ctx, FiveSecondActivity).Get(ctx, nil); err != nil {
		return err
	}
	if err := workflow.Sleep(ctx, 24*time.Hour); err != nil {
		return err
	}
	return workflow.ExecuteActivity(ctx, FiveSecondActivity).Get(ctx, nil)
}

func main() {
	c, err := client.Dial(client.Options{
		HostPort:  "127.0.0.1:7233",
		Namespace: "timeskipping-test",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	w := worker.New(c, "vts-test-queue", worker.Options{})
	w.RegisterWorkflow(TimerWorkflow)
	w.RegisterWorkflow(ActivityTimerActivityWorkflow)
	w.RegisterActivity(FiveSecondActivity)

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatal(err)
	}
}
