package exec

import (
	"context"
	"strings"
	"sync"
)

// MockExecutor is a test implementation of Executor that records commands
// and returns configured outputs/errors.
//
// MockExecutor is safe for concurrent use. All methods use mutex locking
// to ensure thread-safe access to internal state.
//
// Usage:
//
//	mock := NewMockExecutor()
//	mock.SetOutput("ls -la", "file1.txt\nfile2.txt")
//	mock.SetError("rm /protected", errors.New("permission denied"))
//
//	// Use mock in tests...
//	output, err := mock.RunWithOutput(ctx, "ls", "-la")
//
//	// Verify recorded commands
//	commands := mock.Commands()
type MockExecutor struct {
	mu       sync.Mutex
	commands []ExecutedCommand
	outputs  map[string]string
	errors   map[string]error
}

// Compile-time assertion that MockExecutor implements Executor.
var _ Executor = (*MockExecutor)(nil)

// NewMockExecutor creates a new MockExecutor with empty command history
// and response maps.
func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		outputs: make(map[string]string),
		errors:  make(map[string]error),
	}
}

// makeKey creates a lookup key from command name and args.
// The key format is "name arg1 arg2 ..." with single space separators.
func makeKey(name string, args ...string) string {
	if len(args) == 0 {
		return name
	}

	return name + " " + strings.Join(args, " ")
}

// SetOutput configures the output to return for a specific command.
// The cmd parameter should match the full command string (e.g., "ls -la").
//
// If both SetOutput and SetError are called for the same command,
// both values will be returned (output along with the error).
func (m *MockExecutor) SetOutput(cmd, output string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.outputs[cmd] = output
}

// SetError configures the error to return for a specific command.
// The cmd parameter should match the full command string (e.g., "rm /protected").
func (m *MockExecutor) SetError(cmd string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.errors[cmd] = err
}

// Commands returns all executed commands in order of execution.
// Returns a deep copy to prevent external modification of internal state.
func (m *MockExecutor) Commands() []ExecutedCommand {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Return a deep copy to prevent modification
	result := make([]ExecutedCommand, len(m.commands))

	for i, cmd := range m.commands {
		var argsCopy []string
		if cmd.Args != nil {
			argsCopy = make([]string, len(cmd.Args))
			copy(argsCopy, cmd.Args)
		}
		result[i] = ExecutedCommand{
			Name:  cmd.Name,
			Args:  argsCopy,
			Stdin: cmd.Stdin,
		}
	}

	return result
}

// Reset clears all recorded commands and configured responses.
// Useful for reusing a MockExecutor across multiple test cases.
func (m *MockExecutor) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.commands = nil
	m.outputs = make(map[string]string)
	m.errors = make(map[string]error)
}

// record adds a command to the execution history.
// Must be called while holding the mutex.
func (m *MockExecutor) record(name string, args []string, stdin string) {
	m.commands = append(m.commands, ExecutedCommand{
		Name:  name,
		Args:  args,
		Stdin: stdin,
	})
}

// response returns the configured output and error for a command key.
// Must be called while holding the mutex.
func (m *MockExecutor) response(key string) (string, error) {
	output := m.outputs[key]
	err := m.errors[key]

	return output, err
}

// Run executes a command and returns an error if configured.
// The command is recorded for later assertion.
func (m *MockExecutor) Run(_ context.Context, name string, args ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.record(name, args, "")
	key := makeKey(name, args...)
	_, err := m.response(key)

	return err
}

// RunWithOutput executes a command and returns the configured output/error.
// The command is recorded for later assertion.
func (m *MockExecutor) RunWithOutput(_ context.Context, name string, args ...string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.record(name, args, "")
	key := makeKey(name, args...)

	return m.response(key)
}

// RunWithStdin executes a command with stdin input.
// The command and stdin are recorded for later assertion.
func (m *MockExecutor) RunWithStdin(_ context.Context, stdin, name string, args ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.record(name, args, stdin)
	key := makeKey(name, args...)
	_, err := m.response(key)

	return err
}
