package trace

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

type WorkflowTracer struct {
	client client.Client
	update *WorkflowExecutionUpdate
	opts   WorkflowTraceOptions
	writer *TermWriter

	interruptSignals []os.Signal

	doneChan chan bool
	sigChan  chan os.Signal
	errChan  chan error
}

func NewWorkflowTracer(client client.Client, options ...func(tracer *WorkflowTracer)) (*WorkflowTracer, error) {
	writer, err := NewTermWriter().WithTerminalSize()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize terminal writer: %w", err)
	}
	tracer := &WorkflowTracer{
		client:   client,
		doneChan: make(chan bool),
		errChan:  make(chan error),
		sigChan:  make(chan os.Signal, 1),
		writer:   writer,
	}
	for _, opt := range options {
		opt(tracer)
	}
	signal.Notify(tracer.sigChan, tracer.interruptSignals...)

	return tracer, nil
}

func WithWriter(writer *TermWriter) func(*WorkflowTracer) {
	return func(t *WorkflowTracer) {
		t.writer = writer
	}
}

// WithInterrupts sets the signals that will interrupt the tracer
func WithInterrupts(signals ...os.Signal) func(*WorkflowTracer) {
	return func(t *WorkflowTracer) {
		t.interruptSignals = signals
	}
}

// WithOptions sets the view options for the tracer
func WithOptions(opts WorkflowTraceOptions) func(*WorkflowTracer) {
	return func(t *WorkflowTracer) {
		t.opts = opts
	}
}

// GetExecutionUpdates gets workflow execution updates for a particular workflow
func (t *WorkflowTracer) GetExecutionUpdates(ctx context.Context, wid, rid string) error {
	iter, err := GetWorkflowExecutionUpdates(ctx, t.client, wid, rid, t.opts.NoFold, t.opts.FoldStatus, t.opts.Depth, t.opts.Concurrency)
	if err != nil {
		return err
	}

	// Start a goroutine to receive updates
	go func() {
		for iter.HasNext() {
			if t.update, err = iter.Next(); err != nil {
				t.errChan <- err
			}
		}
		t.doneChan <- true
	}()
	return nil
}

func (t *WorkflowTracer) PrintUpdates(tmpl *ExecutionTemplate, updatePeriod time.Duration) (int, error) {
	var currentEvents int64
	var totalEvents int64
	var isUpToDate bool

	ticker := time.NewTicker(updatePeriod).C

	for {
		select {
		case <-ticker:
			state := t.update.GetState()
			if state == nil {
				continue
			}

			if !isUpToDate {
				currentEvents, totalEvents = state.GetNumberOfEvents()
				// TODO: This will sometime leave the watch hanging on "Processing events" (usually when there's more childs that workers and they're not closing)
				// We could maybe set isUpToDate = true if we've seen the same number of events for a number of loops.
				isUpToDate = totalEvents > 0 && currentEvents >= totalEvents && !state.IsArchived
				_, _ = t.writer.WriteLine(ProgressString(currentEvents, totalEvents))
			} else {
				err := tmpl.Execute(t.writer, t.update.GetState(), 0)
				if err != nil {
					return 1, err
				}
			}
			if err := t.writer.Flush(true); err != nil {
				return 1, err
			}
		case <-t.doneChan:
			return PrintAndExit(t.writer, tmpl, t.update)
		case <-t.sigChan:
			return PrintAndExit(t.writer, tmpl, t.update)
		case err := <-t.errChan:
			return 1, err
		}
	}
}

func ProgressString(currentEvents int64, totalEvents int64) string {
	if totalEvents == 0 {
		if currentEvents == 0 {
			return "Processing HistoryEvents"
		}
		return fmt.Sprintf("Processing HistoryEvents (%d)", currentEvents)
	} else {
		return fmt.Sprintf("Processing HistoryEvents (%d/%d)", currentEvents, totalEvents)
	}
}

func PrintAndExit(writer *TermWriter, tmpl *ExecutionTemplate, update *WorkflowExecutionUpdate) (int, error) {
	state := update.GetState()
	if state == nil {
		return 0, nil
	}
	if err := tmpl.Execute(writer, update.GetState(), 0); err != nil {
		return 1, err
	}
	if err := writer.Flush(false); err != nil {
		return 1, err
	}
	return GetExitCode(update.GetState()), nil
}

// GetExitCode returns the exit code for a given workflow execution status.
func GetExitCode(exec *WorkflowExecutionState) int {
	if exec == nil {
		// Don't panic if the state is missing.
		return 0
	}
	switch exec.Status {
	case enums.WORKFLOW_EXECUTION_STATUS_FAILED:
		return 2
	case enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT:
		return 3
	case enums.WORKFLOW_EXECUTION_STATUS_UNSPECIFIED:
		return 4
	}
	return 0
}
