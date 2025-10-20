package cliext

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// createExecutable creates an executable file in the given directory.
// For Unix-like systems, it creates a bash script with the given content.
// For Windows, it creates a batch file.
//
// Example:
//
//	createExecutable(t, tempDir, "temporal-foo", "echo 'Hello from foo'")
func createExecutable(t *testing.T, dir, name, content string) string {
	t.Helper()

	var fullPath string
	var scriptContent string

	if runtime.GOOS == "windows" {
		// Windows batch file
		fullPath = filepath.Join(dir, name+".bat")
		scriptContent = "@echo off\r\n" + content
	} else {
		// Unix-like shell script
		fullPath = filepath.Join(dir, name)
		scriptContent = "#!/bin/bash\n" + content
	}

	// Write the file
	err := os.WriteFile(fullPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create executable %s: %v", fullPath, err)
	}

	return fullPath
}

// createNonExecutable creates a non-executable file (for negative testing).
func createNonExecutable(t *testing.T, dir, name string) string {
	t.Helper()

	fullPath := filepath.Join(dir, name)
	err := os.WriteFile(fullPath, []byte("not executable"), 0644)
	if err != nil {
		t.Fatalf("Failed to create non-executable file %s: %v", fullPath, err)
	}

	return fullPath
}

// setupTestPATH sets the PATH environment variable to the given directory
// and returns a cleanup function that restores the original PATH.
//
// Example:
//
//	cleanup := setupTestPATH(t, tempDir)
//	defer cleanup()
func setupTestPATH(t *testing.T, paths ...string) func() {
	t.Helper()

	oldPath := os.Getenv("PATH")

	separator := ":"
	if runtime.GOOS == "windows" {
		separator = ";"
	}

	newPath := strings.Join(paths, separator)
	err := os.Setenv("PATH", newPath)
	if err != nil {
		t.Fatalf("Failed to set PATH: %v", err)
	}

	return func() {
		os.Setenv("PATH", oldPath)
	}
}

// appendToTestPATH appends the given directories to the PATH and returns a cleanup function.
func appendToTestPATH(t *testing.T, paths ...string) func() {
	t.Helper()

	oldPath := os.Getenv("PATH")

	separator := ":"
	if runtime.GOOS == "windows" {
		separator = ";"
	}

	allPaths := append([]string{oldPath}, paths...)
	newPath := strings.Join(allPaths, separator)

	err := os.Setenv("PATH", newPath)
	if err != nil {
		t.Fatalf("Failed to set PATH: %v", err)
	}

	return func() {
		os.Setenv("PATH", oldPath)
	}
}

// runCommand runs a command and returns its combined output (stdout + stderr).
func runCommand(t *testing.T, name string, args ...string) (string, error) {
	t.Helper()

	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()

	return string(output), err
}

// runCommandWithEnv runs a command with custom environment variables.
func runCommandWithEnv(t *testing.T, env []string, name string, args ...string) (string, error) {
	t.Helper()

	cmd := exec.Command(name, args...)
	cmd.Env = env
	output, err := cmd.CombinedOutput()

	return string(output), err
}

// mustRunCommand runs a command and fails the test if it returns an error.
func mustRunCommand(t *testing.T, name string, args ...string) string {
	t.Helper()

	output, err := runCommand(t, name, args...)
	if err != nil {
		t.Fatalf("Command %s %v failed: %v\nOutput: %s", name, args, err, output)
	}

	return output
}

// assertCommandFails asserts that a command fails with a non-zero exit code.
func assertCommandFails(t *testing.T, name string, args ...string) string {
	t.Helper()

	output, err := runCommand(t, name, args...)
	if err == nil {
		t.Fatalf("Command %s %v should have failed but succeeded", name, args)
	}

	return output
}

// getExitCode extracts the exit code from a command error.
func getExitCode(err error) int {
	if err == nil {
		return 0
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode()
	}

	return -1
}

// tryExecuteExtension attempts to find and execute an extension for the given command tokens.
//
// This is a test helper that combines discovery, matching, and execution.
// It returns the result of execution if an extension is found, or an error if no
// matching extension exists.
//
// Example:
//
//	result := tryExecuteExtension(ctx, []string{"foo", "bar", "--help"}, 0)
//	if result.Error != nil {
//	    // Handle error
//	}
//	os.Exit(result.ExitCode)
func tryExecuteExtension(ctx context.Context, commandTokens []string, timeout time.Duration) *ExecuteExtensionResult {
	// Discover extensions
	extensions := DiscoverExtensions()
	if len(extensions) == 0 {
		return &ExecuteExtensionResult{
			ExitCode: 1,
			Error:    fmt.Errorf("no extensions found"),
		}
	}

	// Find matching extension
	ext, remainingArgs := FindExtension(extensions, commandTokens)
	if ext == nil {
		return &ExecuteExtensionResult{
			ExitCode: 1,
			Error:    fmt.Errorf("no extension found for command: %v", commandTokens),
		}
	}

	// Execute the extension with remaining args using standard I/O
	io := NewStdIOConfig()
	exitCode, err := ExecuteExtension(ctx, ext, remainingArgs, io, timeout)
	return &ExecuteExtensionResult{
		Extension: ext,
		ExitCode:  exitCode,
		Error:     err,
	}
}
