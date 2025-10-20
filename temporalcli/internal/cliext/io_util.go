package cliext

import (
	"bytes"
	"io"
	"os"
)

// IOConfig holds the I/O configuration for extension execution.
type IOConfig struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// NewStdIOConfig creates an IOConfig that uses standard I/O streams.
// This is the default for interactive extension execution.
func NewStdIOConfig() *IOConfig {
	return &IOConfig{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

// NewCaptureIOConfig creates an IOConfig that captures stdout and stderr.
// This is useful for testing or when you need to process extension output.
//
// Returns the IOConfig and buffers for stdout and stderr.
func NewCaptureIOConfig(stdin io.Reader) (*IOConfig, *bytes.Buffer, *bytes.Buffer) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	if stdin == nil {
		stdin = os.Stdin
	}

	return &IOConfig{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	}, stdout, stderr
}

// NewCustomIOConfig creates an IOConfig with custom readers/writers.
func NewCustomIOConfig(stdin io.Reader, stdout, stderr io.Writer) *IOConfig {
	return &IOConfig{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	}
}
