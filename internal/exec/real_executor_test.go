package exec

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRealExecutorImplementsInterface(t *testing.T) {
	// Compile-time assertion
	var _ Executor = (*RealExecutor)(nil)

	// Runtime verification
	executor := NewRealExecutor()
	assert.NotNil(t, executor)
	assert.Zero(t, executor.Timeout)
}

func TestNewRealExecutorWithTimeout(t *testing.T) {
	timeout := 5 * time.Second
	executor := NewRealExecutorWithTimeout(timeout)
	assert.Equal(t, timeout, executor.Timeout)
}

func TestRealExecutorRun(t *testing.T) {
	executor := NewRealExecutor()

	err := executor.Run(t.Context(), "echo", "hello")
	assert.NoError(t, err)
}

func TestRealExecutorRunFailure(t *testing.T) {
	executor := NewRealExecutor()

	err := executor.Run(t.Context(), "nonexistent-command-12345")
	assert.Error(t, err)
}

func TestRealExecutorRunWithOutput(t *testing.T) {
	executor := NewRealExecutor()

	output, err := executor.RunWithOutput(t.Context(), "echo", "hello world")
	require.NoError(t, err)
	assert.Equal(t, "hello world\n", output)
}

func TestRealExecutorRunWithOutputMultipleArgs(t *testing.T) {
	executor := NewRealExecutor()

	output, err := executor.RunWithOutput(t.Context(), "echo", "arg1", "arg2", "arg3")
	require.NoError(t, err)
	assert.Equal(t, "arg1 arg2 arg3\n", output)
}

func TestRealExecutorRunWithOutputFailure(t *testing.T) {
	executor := NewRealExecutor()

	output, err := executor.RunWithOutput(t.Context(), "ls", "/nonexistent-path-12345")
	assert.Error(t, err)
	// Should contain error output
	assert.NotEmpty(t, output)
}

func TestRealExecutorRunWithStdin(t *testing.T) {
	executor := NewRealExecutor()

	// Use cat to echo stdin, but we're only checking no error occurs
	err := executor.RunWithStdin(t.Context(), "hello from stdin", "cat")
	assert.NoError(t, err)
}

func TestRealExecutorRunWithStdinGrep(t *testing.T) {
	executor := NewRealExecutor()

	// Test stdin with grep - should succeed when pattern matches
	err := executor.RunWithStdin(t.Context(), "line1\nmatch-me\nline3", "grep", "match-me")
	assert.NoError(t, err)
}

func TestRealExecutorRunWithStdinGrepNoMatch(t *testing.T) {
	executor := NewRealExecutor()

	// grep returns exit code 1 when no match found
	err := executor.RunWithStdin(t.Context(), "line1\nline2\nline3", "grep", "no-match")
	assert.Error(t, err)
}

func TestRealExecutorTimeout(t *testing.T) {
	// Create executor with 100ms timeout
	executor := NewRealExecutorWithTimeout(100 * time.Millisecond)

	// Sleep for longer than timeout
	err := executor.Run(t.Context(), "sleep", "5")
	require.Error(t, err)

	// Check that error indicates process was killed/signaled
	assert.True(t,
		strings.Contains(err.Error(), "killed") ||
			strings.Contains(err.Error(), "signal"),
		"expected timeout error, got: %v", err)
}

func TestRealExecutorContextCancellation(t *testing.T) {
	executor := NewRealExecutor()
	ctx, cancel := context.WithCancel(t.Context())

	// Cancel context immediately
	cancel()

	// Command should fail due to canceled context
	err := executor.Run(ctx, "sleep", "5")
	assert.Error(t, err)
}

func TestRealExecutorTimeoutWithOutput(t *testing.T) {
	executor := NewRealExecutorWithTimeout(100 * time.Millisecond)

	output, err := executor.RunWithOutput(t.Context(), "sleep", "5")
	require.Error(t, err)
	// Output may be empty or contain error message
	_ = output
}

func TestRealExecutorNoTimeoutSuccess(t *testing.T) {
	// Executor without timeout - command should complete normally
	executor := NewRealExecutor()

	// Quick command that completes well under any reasonable timeout
	output, err := executor.RunWithOutput(t.Context(), "echo", "fast")
	require.NoError(t, err)
	assert.Equal(t, "fast\n", output)
}

func TestRealExecutorApplyTimeoutWithZeroTimeout(t *testing.T) {
	executor := NewRealExecutor()
	ctx := t.Context()

	newCtx, cancel := executor.applyTimeout(ctx)
	defer cancel()

	// Context should be the same when no timeout
	assert.Equal(t, ctx, newCtx)
}

func TestRealExecutorApplyTimeoutWithNonZeroTimeout(t *testing.T) {
	executor := NewRealExecutorWithTimeout(1 * time.Second)
	ctx := t.Context()

	newCtx, cancel := executor.applyTimeout(ctx)
	defer cancel()

	// Context should be different (derived with timeout)
	assert.NotEqual(t, ctx, newCtx)

	// Should have a deadline
	deadline, ok := newCtx.Deadline()
	assert.True(t, ok)
	assert.False(t, deadline.IsZero())
}
