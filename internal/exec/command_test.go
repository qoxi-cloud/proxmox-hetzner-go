package exec

import (
	"context"
	"fmt"
	"testing"
)

// testOutputValue is a constant used by testExecutor to verify interface compliance.
const testOutputValue = "test output"

// TestExecutorInterfaceMethodSignatures ensures testExecutor satisfies the
// Executor interface and that its methods can be invoked with the expected
// signatures without asserting any specific behavior.
func TestExecutorInterfaceMethodSignatures(t *testing.T) {
	// Compile-time assertion that *testExecutor implements Executor.
	var _ Executor = (*testExecutor)(nil)

	// Runtime calls to verify that the method signatures are usable.
	ctx := t.Context()
	executor := &testExecutor{}

	//nolint:errcheck // Testing method signatures, not behavior
	executor.Run(ctx, "echo", "hello")
	//nolint:errcheck // Testing method signatures, not behavior
	executor.RunWithOutput(ctx, "echo", "hello")
	//nolint:errcheck // Testing method signatures, not behavior
	executor.RunWithStdin(ctx, "input data", "cat")
}

// testExecutor is a minimal implementation used to verify interface compliance.
type testExecutor struct{}

// Compile-time assertion that testExecutor implements Executor.
var _ Executor = (*testExecutor)(nil)

func (e *testExecutor) Run(_ context.Context, _ string, _ ...string) error {
	return nil
}

func (e *testExecutor) RunWithOutput(_ context.Context, _ string, _ ...string) (string, error) {
	return testOutputValue, nil
}

func (e *testExecutor) RunWithStdin(_ context.Context, _, _ string, _ ...string) error {
	return nil
}

// TestExecutedCommandString tests the String() method of ExecutedCommand.
func TestExecutedCommandString(t *testing.T) {
	tests := []struct {
		name     string
		cmd      ExecutedCommand
		expected string
	}{
		{
			name: "command with no args",
			cmd: ExecutedCommand{
				Name: "ls",
				Args: nil,
			},
			expected: "ls",
		},
		{
			name: "command with empty args slice",
			cmd: ExecutedCommand{
				Name: "pwd",
				Args: []string{},
			},
			expected: "pwd",
		},
		{
			name: "command with single arg",
			cmd: ExecutedCommand{
				Name: "ls",
				Args: []string{"-la"},
			},
			expected: "ls -la",
		},
		{
			name: "command with multiple args",
			cmd: ExecutedCommand{
				Name: "git",
				Args: []string{"add", "file1.go", "file2.go"},
			},
			expected: "git add file1.go file2.go",
		},
		{
			name: "empty command name with no args",
			cmd: ExecutedCommand{
				Name: "",
				Args: nil,
			},
			expected: "",
		},
		{
			name: "empty command name with args",
			cmd: ExecutedCommand{
				Name: "",
				Args: []string{"arg1", "arg2"},
			},
			expected: " arg1 arg2",
		},
		{
			name: "command with args containing spaces",
			cmd: ExecutedCommand{
				Name: "echo",
				Args: []string{"hello world", "foo bar"},
			},
			expected: "echo hello world foo bar",
		},
		{
			name: "stdin field does not affect output",
			cmd: ExecutedCommand{
				Name:  "cat",
				Args:  []string{"-n"},
				Stdin: "some input data",
			},
			expected: "cat -n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cmd.String()
			if result != tt.expected {
				t.Errorf("ExecutedCommand.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestExecutedCommandStringImplementsStringer verifies that ExecutedCommand
// implements the fmt.Stringer interface.
func TestExecutedCommandStringImplementsStringer(t *testing.T) {
	// Compile-time assertion that ExecutedCommand implements fmt.Stringer.
	var _ fmt.Stringer = ExecutedCommand{}

	// Runtime verification that fmt functions use the String() method.
	cmd := ExecutedCommand{
		Name: "go",
		Args: []string{"test", "./..."},
	}

	//nolint:gocritic // intentionally testing fmt.Stringer interface integration via fmt package
	formatted := fmt.Sprintf("%s", cmd)
	expected := "go test ./..."

	if formatted != expected {
		t.Errorf("fmt.Sprintf(\"%%s\", cmd) = %q, want %q", formatted, expected)
	}
}
