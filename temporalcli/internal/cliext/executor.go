package cliext

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

// ExecuteExtension executes an extension binary with the provided I/O configuration.
//
// It runs the extension as a subprocess with the given arguments, using the
// provided IOConfig for stdin, stdout, and stderr. The exit code is returned
// along with any execution errors.
//
// If timeout is greater than zero, the extension will be killed if it
// runs longer than the specified duration.
//
// Example with standard I/O:
//
//	io := cliext.NewStdIOConfig()
//	exitCode, err := ExecuteExtension(ctx, ext, []string{"--help"}, io, 0)
//
// Example with captured output:
//
//	io, stdout, stderr := cliext.NewCaptureIOConfig(nil)
//	exitCode, err := ExecuteExtension(ctx, ext, args, io, 0)
//	fmt.Println("Output:", stdout.String())
func ExecuteExtension(ctx context.Context, extension *Extension, args []string, ioConfig *IOConfig, timeout time.Duration) (int, error) {
	if extension == nil {
		return 1, fmt.Errorf("extension cannot be nil")
	}

	if ioConfig == nil {
		ioConfig = NewStdIOConfig()
	}

	// Create command context with timeout if specified
	var cmd *exec.Cmd
	if timeout > 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		cmd = exec.CommandContext(timeoutCtx, extension.Path, args...)
	} else {
		cmd = exec.CommandContext(ctx, extension.Path, args...)
	}

	// Wire up I/O
	cmd.Stdin = ioConfig.Stdin
	cmd.Stdout = ioConfig.Stdout
	cmd.Stderr = ioConfig.Stderr

	// Run the command
	err := cmd.Run()
	if err != nil {
		// Extract exit code from error
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), nil
		}
		// Other errors (e.g., command not found, permission denied)
		return 1, fmt.Errorf("failed to execute extension %s: %w", extension.Name, err)
	}

	return 0, nil
}

// ExecuteExtensionResult holds the result of an extension execution.
type ExecuteExtensionResult struct {
	Extension *Extension // nil if no extension was found
	ExitCode  int
	Error     error
}

// TryExecuteExtension attempts to find and execute an extension for the given command tokens.
// It returns nil if no extension is found, or an ExecuteExtensionResult if an extension was attempted.
func TryExecuteExtension(ctx context.Context, commandTokens []string, timeout time.Duration, ioConfig *IOConfig) *ExecuteExtensionResult {
	// Discover extensions
	extensions := DiscoverExtensions()
	if len(extensions) == 0 {
		return nil
	}

	// Find matching extension
	ext, remainingArgs := FindExtension(extensions, commandTokens)
	if ext == nil {
		return nil
	}

	// Execute the extension with remaining args
	if ioConfig == nil {
		ioConfig = NewStdIOConfig()
	}
	exitCode, err := ExecuteExtension(ctx, ext, remainingArgs, ioConfig, timeout)
	return &ExecuteExtensionResult{
		Extension: ext,
		ExitCode:  exitCode,
		Error:     err,
	}
}
