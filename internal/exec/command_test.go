package exec

import (
	"context"
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
