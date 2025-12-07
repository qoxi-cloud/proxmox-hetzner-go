package exec

import "context"

// Executor defines the interface for running system commands.
// All methods support context.Context for cancellation and timeout.
//
// This interface enables testing of code that needs to execute system commands
// by allowing injection of mock implementations that simulate command behavior
// without actually running external processes.
type Executor interface {
	// Run executes a command and returns an error if it fails.
	// Stdout and stderr are discarded.
	// The command will be terminated if the context is canceled.
	Run(ctx context.Context, name string, args ...string) error

	// RunWithOutput executes a command and returns combined stdout/stderr.
	// Useful for commands where you need to capture the output.
	// The command will be terminated if the context is canceled.
	RunWithOutput(ctx context.Context, name string, args ...string) (string, error)

	// RunWithStdin executes a command with stdin input.
	// Useful for commands that read from stdin (e.g., piping data).
	// The command will be terminated if the context is canceled.
	RunWithStdin(ctx context.Context, stdin string, name string, args ...string) error
}
