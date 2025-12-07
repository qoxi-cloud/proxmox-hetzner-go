package exec

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"
)

// ExecutedCommand represents a command that was executed.
// Used by MockExecutor to record command invocations for test assertions.
type ExecutedCommand struct {
	// Name is the command name (e.g., "ls", "grep").
	Name string

	// Args contains the command arguments.
	Args []string

	// Stdin contains the stdin input provided to the command, if any.
	Stdin string
}

// String returns a string representation of the executed command
// in the format "name arg1 arg2 ...".
func (c ExecutedCommand) String() string {
	if len(c.Args) == 0 {
		return c.Name
	}

	return c.Name + " " + strings.Join(c.Args, " ")
}

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

// RealExecutor executes actual system commands using os/exec.
//
// Security note: RealExecutor intentionally accepts dynamic command names and arguments.
// This is by design as it serves as the production implementation of the Executor interface.
// Callers are responsible for validating and sanitizing command inputs before invoking
// the executor methods. Use MockExecutor for testing to avoid executing real commands.
type RealExecutor struct {
	// Timeout is the default timeout for command execution.
	// If zero, commands run with the context's deadline only.
	Timeout time.Duration
}

// Compile-time assertion that RealExecutor implements Executor.
var _ Executor = (*RealExecutor)(nil)

// NewRealExecutor creates a new RealExecutor without a default timeout.
// Commands will run with the context's deadline only.
func NewRealExecutor() *RealExecutor {
	return &RealExecutor{}
}

// NewRealExecutorWithTimeout creates a new RealExecutor with the specified
// default timeout applied to all command executions.
func NewRealExecutorWithTimeout(timeout time.Duration) *RealExecutor {
	return &RealExecutor{Timeout: timeout}
}

// applyTimeout creates a derived context with timeout if Timeout > 0.
// Returns the original context and a no-op cancel func if no timeout is set.
func (e *RealExecutor) applyTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if e.Timeout > 0 {
		return context.WithTimeout(ctx, e.Timeout)
	}

	return ctx, func() {
		// no-op cancel: safe to call even when no timeout is configured
	}
}

// Run executes a command and returns an error if it fails.
// Stdout and stderr are discarded.
func (e *RealExecutor) Run(ctx context.Context, name string, args ...string) error {
	ctx, cancel := e.applyTimeout(ctx)
	defer cancel()

	// nosemgrep: go.lang.security.audit.dangerous-exec-command -- intentional dynamic command execution
	return exec.CommandContext(ctx, name, args...).Run()
}

// RunWithOutput executes a command and returns combined stdout/stderr.
func (e *RealExecutor) RunWithOutput(ctx context.Context, name string, args ...string) (string, error) {
	ctx, cancel := e.applyTimeout(ctx)
	defer cancel()

	// nosemgrep: go.lang.security.audit.dangerous-exec-command -- intentional dynamic command execution
	out, err := exec.CommandContext(ctx, name, args...).CombinedOutput()

	return string(out), err
}

// RunWithStdin executes a command with stdin input.
func (e *RealExecutor) RunWithStdin(ctx context.Context, stdin, name string, args ...string) error {
	ctx, cancel := e.applyTimeout(ctx)
	defer cancel()

	// nosemgrep: go.lang.security.audit.dangerous-exec-command -- intentional dynamic command execution
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = bytes.NewBufferString(stdin)

	return cmd.Run()
}
