// Package exec provides an abstraction layer for system command execution,
// enabling testing of components that depend on running external commands.
//
// The package defines an Executor interface that abstracts os/exec operations,
// along with RealExecutor for production use and MockExecutor for testing.
//
// Usage:
//
//	// Production code
//	exec := exec.NewRealExecutor()
//	output, err := exec.RunWithOutput(ctx, "ls", "-la")
//
//	// Test code
//	mock := exec.NewMockExecutor()
//	mock.SetOutput("ls -la", "file1\nfile2")
//	output, err := mock.RunWithOutput(ctx, "ls", "-la")
package exec
