// Package exec provides an abstraction layer for system command execution,
// enabling testing of components that depend on running external commands.
//
// The package defines an Executor interface that abstracts os/exec operations,
// along with RealExecutor for production use and MockExecutor for testing.
//
// Usage:
//
//	// Production code
//	executor := exec.NewRealExecutor()
//	output, err := executor.RunWithOutput(ctx, "ls", "-la")
//
//	// Test code
//	// Note: SetOutput key is constructed by joining command and args with spaces.
//	// For RunWithOutput(ctx, "ls", "-la"), the key is "ls -la".
//	mock := exec.NewMockExecutor()
//	mock.SetOutput("ls -la", "file1\nfile2")
//	output, err := mock.RunWithOutput(ctx, "ls", "-la")
package exec
