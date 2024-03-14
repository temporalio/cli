package temporalcli_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/cli/temporalcli"
	"github.com/temporalio/cli/temporalcli/devserver"

	"google.golang.org/grpc"
)

type CommandHarness struct {
	*require.Assertions
	t       *testing.T
	Options temporalcli.CommandOptions
	// Defaults to a context closed on close or test complete
	Context context.Context
	// Can be used to cancel context given to commands (simulating interrupt)
	CancelContext context.CancelFunc
	Stdin         bytes.Buffer
}

func NewCommandHarness(t *testing.T) *CommandHarness {
	h := &CommandHarness{Assertions: require.New(t), t: t}
	h.Context, h.CancelContext = context.WithCancel(context.Background())
	t.Cleanup(h.Close)
	return h
}

// Reentrant, called after test by default, cancels context
func (h *CommandHarness) Close() {
	// Cancel context
	if h.CancelContext != nil {
		h.CancelContext()
	}
}

// Pieces must appear in order on the line and not overlap
func (h *CommandHarness) ContainsOnSameLine(text string, pieces ...string) {
	h.NoError(AssertContainsOnSameLine(text, pieces...))
}

func AssertContainsOnSameLine(text string, pieces ...string) error {
	// Build regex pattern based on pieces
	pattern := ""
	for _, piece := range pieces {
		if pattern != "" {
			pattern += ".*"
		}
		pattern += regexp.QuoteMeta(piece)
	}
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	// Split into lines, then check each piece is present
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if regex.MatchString(line) {
			return nil
		}
	}
	return fmt.Errorf("pieces not found in order on any line together")
}

func TestAssertContainsOnSameLine(t *testing.T) {
	require.Error(t, AssertContainsOnSameLine("a b c", "b", "a"))
	require.Error(t, AssertContainsOnSameLine("a\nb c", "a", "b"))
	require.NoError(t, AssertContainsOnSameLine("aba", "b", "a"))
	require.NoError(t, AssertContainsOnSameLine("a b a", "b", "a"))
	require.NoError(t, AssertContainsOnSameLine("axb", "a", "b"))
	require.NoError(t, AssertContainsOnSameLine("a a", "a", "a"))
}

func (h *CommandHarness) Eventually(
	condition func() bool,
	waitFor time.Duration,
	tick time.Duration,
	msgAndArgs ...interface{},
) {
	// We cannot use require.Eventually because it was poorly developed to run the
	// condition function in a goroutine which means it can run after complete or
	// have other race conditions. Don't even need a complicated ticker because it
	// doesn't need to be interruptible.
	for start := time.Now(); time.Since(start) < waitFor; {
		if condition() {
			return
		}
		time.Sleep(tick)
	}
	h.Fail("condition did not evaluate to true within timeout", msgAndArgs...)
}

func (h *CommandHarness) T() *testing.T {
	return h.t
}

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
	options.Stdin = &h.Stdin
	options.Stdout = &res.Stdout
	options.Stderr = &res.Stderr
	// Set args
	options.Args = args
	// Disable env if no env file and no --env-file arg
	options.DisableEnvConfig = options.EnvConfigFile == "" && !slices.Contains(args, "--env-file")
	// Capture error
	options.Fail = func(err error) {
		if res.Err != nil {
			panic("fail called twice")
		}
		res.Err = err
	}

	// Run
	ctx, cancel := context.WithCancel(h.Context)
	h.t.Cleanup(cancel)
	defer cancel()
	h.t.Logf("Calling: %v", strings.Join(args, " "))
	temporalcli.Execute(ctx, options)
	if res.Stdout.Len() > 0 {
		h.t.Logf("Stdout:\n-----\n%s\n-----", &res.Stdout)
	}
	if res.Stderr.Len() > 0 {
		h.t.Logf("Stderr:\n-----\n%s\n-----", &res.Stderr)
	}
	return res
}

// Run shared server suite
func TestSharedServerSuite(t *testing.T) {
	suite.Run(t, new(SharedServerSuite))
}

type SharedServerSuite struct {
	// Replaced each test
	*CommandHarness

	*DevServer
	Worker *DevWorker
	Suite  suite.Suite
}

func (s *SharedServerSuite) SetupSuite() {
	s.DevServer = StartDevServer(s.Suite.T(), DevServerOptions{
		StartOptions: devserver.StartOptions{
			// Enable for operator cluster commands
			EnableGlobalNamespace: true,
		},
	})
	// Stop server if we fail later
	success := false
	defer func() {
		if !success {
			s.Server.Stop()
		}
	}()
	s.Worker = s.DevServer.StartDevWorker(s.Suite.T(), DevWorkerOptions{})
	success = true
}

func (s *SharedServerSuite) TearDownSuite() {
	s.Stop()
}

func (s *SharedServerSuite) Stop() {
	s.Worker.Stop()
	s.DevServer.Stop()
}

func (s *SharedServerSuite) SetupTest() {
	// Clear log buffer
	s.ResetLogOutput()
	// Reset worker
	s.Worker.Reset()
	// Create new command harness
	s.CommandHarness = NewCommandHarness(s.Suite.T())
}

func (s *SharedServerSuite) TearDownTest() {
	// If there is log output, log it
	if b := s.LogOutput(); len(b) > 0 {
		s.t.Logf("Server/SDK Log Output:\n-----\n%s-----", b)
	}
	if s.CommandHarness != nil {
		s.CommandHarness.Close()
	}
	s.CommandHarness = nil
}

func (s *SharedServerSuite) T() *testing.T                 { return s.Suite.T() }
func (s *SharedServerSuite) SetT(t *testing.T)             { s.Suite.SetT(t) }
func (s *SharedServerSuite) SetS(suite suite.TestingSuite) { s.Suite.SetS(suite) }

type DevServer struct {
	Server *devserver.Server
	// With defaults populated
	Options DevServerOptions
	// For first namespace in options
	Client client.Client
	// For other services, like the WorkflowService

	logOutput     bytes.Buffer
	logOutputLock sync.RWMutex

	serverInterceptor     grpc.UnaryServerInterceptor
	serverInterceptorLock sync.RWMutex
}

type DevServerOptions struct {
	// Required options are set with reasonable defaults if not present
	devserver.StartOptions
	// HostPort and Namespace is overridden
	ClientOptions client.Options
}

func StartDevServer(t *testing.T, options DevServerOptions) *DevServer {
	success := false

	// Build options
	d := &DevServer{Options: options}
	if d.Options.FrontendIP == "" {
		d.Options.FrontendIP = "127.0.0.1"
	}
	if d.Options.FrontendPort == 0 {
		d.Options.FrontendPort = devserver.MustGetFreePort()
	}
	if len(d.Options.Namespaces) == 0 {
		d.Options.Namespaces = []string{
			"default",
			"batch-empty", // for test `TestBatchJob_List
		}
	}
	if d.Options.MasterClusterName == "" {
		d.Options.MasterClusterName = "active"
	}
	if d.Options.CurrentClusterName == "" {
		d.Options.CurrentClusterName = "active"
	}
	if d.Options.ClusterID == "" {
		d.Options.ClusterID = uuid.New().String()
	}
	if d.Options.InitialFailoverVersion == 0 {
		d.Options.InitialFailoverVersion = 1
	}

	if d.Options.Logger == nil {
		w := &concurrentWriter{w: &d.logOutput, wLock: &d.logOutputLock}
		d.Options.Logger = slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{AddSource: true}))
		// If this fails, we want to dump logs
		defer func() {
			if !success {
				if b := d.LogOutput(); len(b) > 0 {
					t.Logf("Server/SDK Log Output:\n-----\n%s-----", b)
				}
			}
		}()
	}
	if d.Options.ClientOptions.Logger == nil {
		d.Options.ClientOptions.Logger = d.Options.Logger
	}
	d.Options.ClientOptions.HostPort = fmt.Sprintf("%v:%v", d.Options.FrontendIP, d.Options.FrontendPort)
	d.Options.ClientOptions.Namespace = d.Options.Namespaces[0]
	if d.Options.ClientOptions.Identity == "" {
		d.Options.ClientOptions.Identity = "cli-test-client"
	}
	if d.Options.DynamicConfigValues == nil {
		d.Options.DynamicConfigValues = map[string]any{}
	}
	d.Options.DynamicConfigValues["system.forceSearchAttributesCacheRefreshOnRead"] = true
	d.Options.DynamicConfigValues["frontend.workerVersioningDataAPIs"] = true
	d.Options.DynamicConfigValues["frontend.workerVersioningWorkflowAPIs"] = true
	d.Options.DynamicConfigValues["worker.buildIdScavengerEnabled"] = true
	d.Options.DynamicConfigValues["frontend.enableUpdateWorkflowExecution"] = true
	d.Options.DynamicConfigValues["frontend.MaxConcurrentBatchOperationPerNamespace"] = 1000

	d.Options.GRPCInterceptors = append(
		d.Options.GRPCInterceptors,
		func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
			d.serverInterceptorLock.RLock()
			serverInterceptor := d.serverInterceptor
			d.serverInterceptorLock.RUnlock()
			if serverInterceptor != nil {
				return serverInterceptor(ctx, req, info, handler)
			}
			return handler(ctx, req)
		},
	)

	// Start
	var err error
	d.Server, err = devserver.Start(d.Options.StartOptions)
	require.NoError(t, err)
	defer func() {
		if !success {
			d.Server.Stop()
		}
	}()

	// Dial sdk client
	d.Client, err = client.Dial(d.Options.ClientOptions)
	require.NoError(t, err)
	defer func() {
		if !success {
			d.Client.Close()
		}
	}()

	// Create search attribute if not there
	ctx := context.Background()
	saResp, err := d.Client.OperatorService().ListSearchAttributes(ctx, &operatorservice.ListSearchAttributesRequest{
		Namespace: d.Options.ClientOptions.Namespace,
	})
	require.NoError(t, err)
	if _, ok := saResp.CustomAttributes["CustomKeywordField"]; !ok {
		_, err = d.Client.OperatorService().AddSearchAttributes(ctx, &operatorservice.AddSearchAttributesRequest{
			Namespace:        d.Options.ClientOptions.Namespace,
			SearchAttributes: map[string]enums.IndexedValueType{"CustomKeywordField": enums.INDEXED_VALUE_TYPE_KEYWORD},
		})
		require.NoError(t, err)
	}

	success = true
	return d
}

func (d *DevServer) Stop() {
	d.Client.Close()
	d.Server.Stop()
}

func (d *DevServer) LogOutput() []byte {
	d.logOutputLock.RLock()
	defer d.logOutputLock.RUnlock()
	// Copy bytes
	b := d.logOutput.Bytes()
	newB := make([]byte, len(b))
	copy(newB, b)
	return newB
}

func (d *DevServer) ResetLogOutput() {
	d.logOutputLock.Lock()
	defer d.logOutputLock.Unlock()
	d.logOutput.Reset()
}

// Shortcut for d.Options.ClientOptions.HostPort
func (d *DevServer) Address() string {
	return d.Options.ClientOptions.HostPort
}

// Shortcut for d.Options.ClientOptions.Namespace
func (d *DevServer) Namespace() string {
	return d.Options.ClientOptions.Namespace
}

func (d *DevServer) SetServerInterceptor(serverInterceptor grpc.UnaryServerInterceptor) {
	d.serverInterceptorLock.Lock()
	defer d.serverInterceptorLock.Unlock()
	d.serverInterceptor = serverInterceptor
}

type DevWorker struct {
	Worker worker.Worker
	// Has defaults populated
	Options DevWorkerOptions

	// Do not access these fields directly
	devOpsLock           sync.Mutex
	devWorkflowCallback  func(workflow.Context, any) (any, error)
	devWorkflowLastInput any
	devActivityCallback  func(context.Context, any) (any, error)
}

type DevWorkerOptions struct {
	Worker worker.Options
	// Default is random UUID
	TaskQueue string
	// Optional, no default, but DevWorkflow is always registered
	Workflows []any
	// Optional, no default, but DevActivity is always registered
	Activities []any
}

// Simply a stub for client use
func DevWorkflow(workflow.Context, any) (any, error) { panic("Unreachable") }

// Simply a stub for client use
func DevActivity(context.Context, any) (any, error) { panic("Unreachable") }

// Stops when harness closes.
func (d *DevServer) StartDevWorker(t *testing.T, options DevWorkerOptions) *DevWorker {
	// Prepare options
	w := &DevWorker{Options: options}
	if w.Options.TaskQueue == "" {
		w.Options.TaskQueue = uuid.NewString()
	}
	if w.Options.Worker.OnFatalError == nil {
		w.Options.Worker.OnFatalError = func(err error) {
			t.Logf("Worker fatal error: %v", err)
		}
	}
	// Create worker and register workflows/activities
	w.Worker = worker.New(d.Client, w.Options.TaskQueue, w.Options.Worker)
	for _, wf := range w.Options.Workflows {
		w.Worker.RegisterWorkflow(wf)
	}
	for _, act := range w.Options.Activities {
		w.Worker.RegisterActivity(act)
	}
	ops := &devOperations{w}
	w.Worker.RegisterWorkflowWithOptions(ops.DevWorkflow, workflow.RegisterOptions{Name: "DevWorkflow"})
	w.Worker.RegisterActivity(ops.DevActivity)
	// Start worker or fail
	require.NoError(t, w.Worker.Start(), "failed starting worker")
	return w
}

func (d *DevWorker) Stop() {
	d.Worker.Stop()
}

// Default is just to return DevActivity result
func (d *DevWorker) OnDevWorkflow(fn func(workflow.Context, any) (any, error)) {
	d.devOpsLock.Lock()
	defer d.devOpsLock.Unlock()
	d.devWorkflowCallback = fn
}

func (d *DevWorker) DevWorkflowLastInput() any {
	d.devOpsLock.Lock()
	defer d.devOpsLock.Unlock()
	return d.devWorkflowLastInput
}

// Default is just to return result
func (d *DevWorker) OnDevActivity(fn func(context.Context, any) (any, error)) {
	d.devOpsLock.Lock()
	defer d.devOpsLock.Unlock()
	d.devActivityCallback = fn
}

func (d *DevWorker) Reset() {
	d.devOpsLock.Lock()
	defer d.devOpsLock.Unlock()
	d.devWorkflowCallback = nil
	d.devWorkflowLastInput = nil
	d.devActivityCallback = nil
}

type devOperations struct{ worker *DevWorker }

func (d *devOperations) DevWorkflow(ctx workflow.Context, input any) (any, error) {
	d.worker.devOpsLock.Lock()
	d.worker.devWorkflowLastInput = input
	callback := d.worker.devWorkflowCallback
	d.worker.devOpsLock.Unlock()
	// Set a default retry policy so that logical errors in your test don't hang forever
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ActivityID:          "dev-activity-id",
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	})
	if callback != nil {
		return callback(ctx, input)
	}
	var res any
	err := workflow.ExecuteActivity(ctx, DevActivity, input).Get(ctx, &res)
	return res, err
}

func (d *devOperations) DevActivity(ctx context.Context, input any) (any, error) {
	d.worker.devOpsLock.Lock()
	callback := d.worker.devActivityCallback
	d.worker.devOpsLock.Unlock()
	if callback != nil {
		return callback(ctx, input)
	}
	return input, nil
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

func TestUnknownCommandExitsNonzero(t *testing.T) {
	commandHarness := NewCommandHarness(t)
	res := commandHarness.Execute("blerkflow")
	assert.Contains(t, res.Err.Error(), "unknown command")
}
