package exec

import (
	"context"
	"testing"
)

// TestExecutorInterfaceDefinition verifies the Executor interface is properly defined
// and can be used for type assertions.
func TestExecutorInterfaceDefinition(t *testing.T) {
	// This test verifies that the Executor interface is properly defined
	// by checking that we can use it as a type.
	var executor Executor
	if executor != nil {
		t.Error("expected nil executor")
	}
}

// TestExecutorInterfaceMethodSignatures verifies the interface method signatures
// are correct by creating a minimal mock implementation.
func TestExecutorInterfaceMethodSignatures(t *testing.T) {
	// Create a minimal implementation to verify method signatures compile
	var executor Executor = &testExecutor{}

	ctx := t.Context()

	// Verify Run method signature
	err := executor.Run(ctx, "echo", "hello")
	if err != nil {
		t.Errorf("Run() unexpected error: %v", err)
	}

	// Verify RunWithOutput method signature
	output, err := executor.RunWithOutput(ctx, "echo", "hello")
	if err != nil {
		t.Errorf("RunWithOutput() unexpected error: %v", err)
	}
	if output != "test output" {
		t.Errorf("RunWithOutput() = %q, want %q", output, "test output")
	}

	// Verify RunWithStdin method signature
	err = executor.RunWithStdin(ctx, "input data", "cat")
	if err != nil {
		t.Errorf("RunWithStdin() unexpected error: %v", err)
	}
}

// testExecutor is a minimal implementation used to verify interface compliance.
type testExecutor struct{}

// Compile-time assertion that testExecutor implements Executor.
var _ Executor = (*testExecutor)(nil)

func (e *testExecutor) Run(_ context.Context, _ string, _ ...string) error {
	return nil
}

func (e *testExecutor) RunWithOutput(_ context.Context, _ string, _ ...string) (string, error) {
	return "test output", nil
}

func (e *testExecutor) RunWithStdin(_ context.Context, _, _ string, _ ...string) error {
	return nil
}
