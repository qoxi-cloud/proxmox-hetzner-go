// Package exec provides an abstraction layer for system command execution,
// enabling testing of components that depend on running external commands.
//
// # Interface
//
// The Executor interface defines three methods for running commands:
//   - Run: Execute command, return error only
//   - RunWithOutput: Execute command, return stdout/stderr and error
//   - RunWithStdin: Execute command with stdin input, return error
//
// All methods accept context.Context as the first parameter for cancellation
// and timeout support.
//
// # RealExecutor
//
// RealExecutor implements Executor using os/exec. Use this in production:
//
//	exec := exec.NewRealExecutor()
//	err := exec.Run(ctx, "git", "status")
//
//	// With timeout
//	exec := exec.NewRealExecutorWithTimeout(30 * time.Second)
//	output, err := exec.RunWithOutput(ctx, "slow-command")
//
// # MockExecutor
//
// MockExecutor implements Executor for testing. It records all commands
// and returns configured outputs/errors:
//
//	mock := exec.NewMockExecutor()
//	mock.SetOutput("ip link show", "eth0: state UP")
//	mock.SetError("rm /important", errors.New("permission denied"))
//
//	// Use in tests
//	step := &MyStep{executor: mock}
//	step.Execute(ctx)
//
//	// Verify commands were called
//	assert.True(t, mock.WasCalledWith("ip", "link", "show"))
//	assert.Equal(t, 2, mock.CommandCount())
//
// See CLAUDE.md section "Mock Executor: Use for testing system commands"
// for more examples.
package exec
