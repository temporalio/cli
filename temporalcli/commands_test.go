package temporalcli_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/cli/temporalcli"
	"github.com/temporalio/cli/temporalcli/devserver"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

type CommandHarness struct {
	*require.Assertions
	T       *testing.T
	Options temporalcli.CommandOptions
	// Defaults to a context closed on close or test complete
	Context context.Context
	// Can be used to cancel context given to commands (simulating interrupt)
	CancelContext context.CancelFunc
}

func NewCommandHarness(t *testing.T) *CommandHarness {
	h := &CommandHarness{Assertions: require.New(t), T: t}
	h.Context, h.CancelContext = context.WithCancel(context.Background())
	t.Cleanup(h.Close)
	return h
}

// Reentrant, called after test by default, cancels context
func (h *CommandHarness) Close() { h.CancelContext() }

type CommandResult struct {
	Err    error
	Stdout bytes.Buffer
	Stderr bytes.Buffer
}

func (h *CommandHarness) Execute(args ...string) *CommandResult {
	// Copy options, update as needed
	res := &CommandResult{}
	options := h.Options
	// Set stdio
	options.Stdout, options.Stderr = &res.Stdout, &res.Stderr
	// Set args
	options.Args = args
	// Disable env if no env file
	options.DisableEnv = options.EnvFile == ""
	// Capture error
	options.Fail = func(err error) {
		if res.Err != nil {
			panic("fail called twice")
		}
		res.Err = err
	}

	// Run
	ctx, cancel := context.WithCancel(h.Context)
	h.T.Cleanup(cancel)
	defer cancel()
	h.T.Logf("Calling: %v", strings.Join(args, " "))
	temporalcli.Execute(ctx, options)
	if res.Stdout.Len() > 0 {
		h.T.Logf("Stdout:\n-----\n%s-----", &res.Stdout)
	}
	if res.Stderr.Len() > 0 {
		h.T.Logf("Stderr:\n-----\n%s-----", &res.Stderr)
	}
	return res
}

type WorkerCommandHarness struct {
	*CommandHarness
	Server *DevServer
	Worker *DevWorker
}

type WorkerCommandHarnessOptions struct {
	Server DevServerOptions
	Worker DevWorkerOptions
}

// NewCommandHarness + StartDevServer + StartDevWorker
func StartWorkerCommandHarness(t *testing.T, options WorkerCommandHarnessOptions) *WorkerCommandHarness {
	h := &WorkerCommandHarness{CommandHarness: NewCommandHarness(t)}
	success := false
	defer func() {
		if !success {
			h.Close()
		}
	}()
	h.Server = h.StartDevServer(options.Server)
	h.Worker = h.Server.StartDevWorker(options.Worker)
	success = true
	return h
}

type DevServer struct {
	Harness *CommandHarness
	// With defaults populated
	Options DevServerOptions
	// For first namespace in options
	Client client.Client

	logOutput     bytes.Buffer
	logOutputLock sync.RWMutex
}

type DevServerOptions struct {
	// Required options are set with reasonable defaults if not present
	devserver.StartOptions
	// HostPort and Namespace is overridden
	ClientOptions client.Options
}

// Automatically closed when harness context is (create new harness if needing
// to separate from another harness). Logs are dumped to test Logf on test
// cleanup.
func (h *CommandHarness) StartDevServer(options DevServerOptions) *DevServer {
	// Prepare options
	d := &DevServer{Harness: h, Options: options}
	if d.Options.FrontendIP == "" {
		d.Options.FrontendIP = "127.0.0.1"
	}
	if d.Options.FrontendPort == 0 {
		d.Options.FrontendPort = devserver.MustGetFreePort()
	}
	if len(d.Options.Namespaces) == 0 {
		d.Options.Namespaces = []string{"default"}
	}
	if d.Options.Logger == nil {
		d.Options.Logger = slog.New(slog.NewTextHandler(
			&concurrentWriter{w: &d.logOutput, wLock: &d.logOutputLock}, &slog.HandlerOptions{AddSource: true}))
	}
	if d.Options.ClientOptions.Logger == nil {
		d.Options.ClientOptions.Logger = d.Options.Logger
	}
	d.Options.ClientOptions.HostPort = fmt.Sprintf("%v:%v", d.Options.FrontendIP, d.Options.FrontendPort)
	d.Options.ClientOptions.Namespace = d.Options.Namespaces[0]

	// Create new context, cancel it on failure in this method
	serverCtx, cancel := context.WithCancel(h.Context)
	success := false
	defer func() {
		if !success {
			cancel()
		}
	}()

	// Start
	srv, err := devserver.Start(d.Options.StartOptions)
	// Dump logs on cleanup
	h.T.Cleanup(func() {
		h.T.Logf("Server/SDK Log Output:\n-----\n%s-----", d.LogOutput())
	})
	h.NoError(err)
	// Stop on context cancel
	go func() {
		<-serverCtx.Done()
		srv.Stop()
	}()

	// Dial client
	d.Client, err = client.Dial(d.Options.ClientOptions)
	h.NoError(err)
	success = true
	return d
}

func (d *DevServer) LogOutput() []byte {
	d.logOutputLock.RLock()
	defer d.logOutputLock.RUnlock()
	return d.logOutput.Bytes()
}

// Shortcut for d.Options.ClientOptions.HostPort
func (d *DevServer) Address() string {
	return d.Options.ClientOptions.HostPort
}

// Shortcut for d.Options.ClientOptions.Namespace
func (d *DevServer) Namespace() string {
	return d.Options.ClientOptions.Namespace
}

type DevWorker struct {
	// Has defaults populated. Any DevWorkflowX options here can be mutated after
	// start.
	Options DevWorkerOptions
	Worker  worker.Worker
	// If mutating Options.DevWorkflowX fields after a workflow may have started,
	// a write lock needs to be obtained
	DevWorkflowLock sync.RWMutex
}

type DevWorkerOptions struct {
	Worker worker.Options
	// Default is random UUID
	TaskQueue string
	// Optional, no default, but DevWorkflow is always registered
	Workflows []any
	// Optional, no default
	Activities []any

	// If unset, default behavior is just to return DevWorkflowOutput
	DevWorkflowOnRun func(workflow.Context, any) (any, error)
	// Ignored if DevWorkflowOnRun is non-nil
	DevWorkflowOutput any
	// Ignored if DevWorkflowOnRun is non-nil
	DevWorkflowError error
}

// Simply a stub for client use
func DevWorkflow(workflow.Context, any) (any, error) { panic("Unreachable") }

// Stops when harness closes.
func (d *DevServer) StartDevWorker(options DevWorkerOptions) *DevWorker {
	// Prepare options
	w := &DevWorker{Options: options}
	if w.Options.TaskQueue == "" {
		w.Options.TaskQueue = uuid.NewString()
	}
	// Create worker and register workflows/activities
	w.Worker = worker.New(d.Client, w.Options.TaskQueue, w.Options.Worker)
	for _, wf := range w.Options.Workflows {
		w.Worker.RegisterWorkflow(wf)
	}
	w.Worker.RegisterWorkflowWithOptions(
		(&devWorkflow{opts: &w.Options, optsLock: &w.DevWorkflowLock}).DevWorkflow,
		workflow.RegisterOptions{Name: "DevWorkflow"},
	)
	for _, act := range w.Options.Activities {
		w.Worker.RegisterActivity(act)
	}
	// Start worker or fail
	d.Harness.NoError(w.Worker.Start(), "failed starting worker")

	// Stop on context cancel
	go func() {
		d.Harness.Context.Done()
		w.Worker.Stop()
	}()
	return w
}

type devWorkflow struct {
	opts     *DevWorkerOptions
	optsLock *sync.RWMutex
}

func (d *devWorkflow) DevWorkflow(ctx workflow.Context, input any) (any, error) {
	d.optsLock.RLock()
	run, out, err := d.opts.DevWorkflowOnRun, d.opts.DevWorkflowOutput, d.opts.DevWorkflowError
	d.optsLock.RUnlock()
	if run != nil {
		return run(ctx, input)
	}
	return out, err
}

type concurrentWriter struct {
	w     io.Writer
	wLock sync.Locker
}

func (w *concurrentWriter) Write(p []byte) (n int, err error) {
	w.wLock.Lock()
	defer w.wLock.Unlock()
	return w.w.Write(p)
}
