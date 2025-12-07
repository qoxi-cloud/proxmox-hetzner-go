package exec

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test constants for commonly used literals.
const (
	testFileListOutput   = "file1.txt\nfile2.txt"
	testPermissionDenied = "permission denied"
	testCommandNotFound  = "command not found"
	testInputData        = "input data"
)

func TestMockExecutorImplementsInterface(t *testing.T) {
	// Compile-time assertion is in mock.go
	// Runtime verification
	mock := NewMockExecutor()
	assert.NotNil(t, mock)

	// Verify it's usable as Executor
	var executor Executor = mock
	assert.NotNil(t, executor)
}

func TestNewMockExecutor(t *testing.T) {
	mock := NewMockExecutor()

	assert.NotNil(t, mock)
	assert.NotNil(t, mock.outputs)
	assert.NotNil(t, mock.errors)
	assert.Empty(t, mock.commands)
}

func TestMockExecutorSetOutput(t *testing.T) {
	tests := []struct {
		name           string
		cmd            string
		output         string
		expectedOutput string
	}{
		{
			name:           "simple command",
			cmd:            "ls",
			output:         testFileListOutput,
			expectedOutput: testFileListOutput,
		},
		{
			name:           "command with args",
			cmd:            "ls -la /tmp",
			output:         "total 0\ndrwxr-xr-x 2 root root 40 Jan 1 00:00 .",
			expectedOutput: "total 0\ndrwxr-xr-x 2 root root 40 Jan 1 00:00 .",
		},
		{
			name:           "empty output",
			cmd:            "true",
			output:         "",
			expectedOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockExecutor()
			mock.SetOutput(tt.cmd, tt.output)

			// Verify output is stored
			mock.mu.Lock()
			storedOutput := mock.outputs[tt.cmd]
			mock.mu.Unlock()

			assert.Equal(t, tt.expectedOutput, storedOutput)
		})
	}
}

func TestMockExecutorSetError(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		err         error
		expectedErr error
	}{
		{
			name:        "simple error",
			cmd:         "rm /protected",
			err:         errors.New(testPermissionDenied),
			expectedErr: errors.New(testPermissionDenied),
		},
		{
			name:        "command not found",
			cmd:         "nonexistent",
			err:         errors.New(testCommandNotFound),
			expectedErr: errors.New(testCommandNotFound),
		},
		{
			name:        "nil error",
			cmd:         "valid-cmd",
			err:         nil,
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockExecutor()
			mock.SetError(tt.cmd, tt.err)

			// Verify error is stored
			mock.mu.Lock()
			storedErr := mock.errors[tt.cmd]
			mock.mu.Unlock()

			if tt.expectedErr == nil {
				assert.Nil(t, storedErr)
			} else {
				assert.Equal(t, tt.expectedErr.Error(), storedErr.Error())
			}
		})
	}
}

func TestMockExecutorCommands(t *testing.T) {
	mock := NewMockExecutor()
	ctx := t.Context()

	// Execute some commands
	require.NoError(t, mock.Run(ctx, "echo", "hello"))
	_, err := mock.RunWithOutput(ctx, "ls", "-la")
	require.NoError(t, err)
	require.NoError(t, mock.RunWithStdin(ctx, testInputData, "cat"))

	commands := mock.Commands()

	require.Len(t, commands, 3)

	// Verify first command
	assert.Equal(t, "echo", commands[0].Name)
	assert.Equal(t, []string{"hello"}, commands[0].Args)
	assert.Empty(t, commands[0].Stdin)

	// Verify second command
	assert.Equal(t, "ls", commands[1].Name)
	assert.Equal(t, []string{"-la"}, commands[1].Args)
	assert.Empty(t, commands[1].Stdin)

	// Verify third command
	assert.Equal(t, "cat", commands[2].Name)
	assert.Empty(t, commands[2].Args)
	assert.Equal(t, testInputData, commands[2].Stdin)
}

func TestMockExecutorCommandsReturnsCopy(t *testing.T) {
	mock := NewMockExecutor()
	ctx := t.Context()

	require.NoError(t, mock.Run(ctx, "echo", "hello", "world"))

	// Get commands and modify the returned slice (Name and Args)
	commands := mock.Commands()
	commands[0].Name = "modified"
	commands[0].Args[0] = "modified-arg"

	// Get commands again and verify originals are unchanged (deep copy)
	commandsAgain := mock.Commands()
	assert.Equal(t, "echo", commandsAgain[0].Name, "Name should be unchanged")
	assert.Equal(t, []string{"hello", "world"}, commandsAgain[0].Args, "Args should be unchanged")
}

func TestMockExecutorCommandsEmpty(t *testing.T) {
	mock := NewMockExecutor()

	commands := mock.Commands()

	assert.Empty(t, commands)
	assert.NotNil(t, commands)
}

func TestMockExecutorReset(t *testing.T) {
	mock := NewMockExecutor()
	ctx := t.Context()

	// Set up some state
	mock.SetOutput("ls", "file.txt")
	mock.SetError("rm", errors.New("error"))
	require.NoError(t, mock.Run(ctx, "echo", "hello"))

	// Verify state exists
	assert.NotEmpty(t, mock.Commands())

	// Reset
	mock.Reset()

	// Verify all state is cleared
	commands := mock.Commands()
	assert.Empty(t, commands)

	// Verify outputs and errors are cleared
	mock.mu.Lock()
	assert.Empty(t, mock.outputs)
	assert.Empty(t, mock.errors)
	mock.mu.Unlock()
}

func TestMockExecutorRun(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		args        []string
		configErr   error
		expectErr   bool
		expectedCmd ExecutedCommand
	}{
		{
			name:      "successful command",
			cmd:       "echo",
			args:      []string{"hello", "world"},
			configErr: nil,
			expectErr: false,
			expectedCmd: ExecutedCommand{
				Name:  "echo",
				Args:  []string{"hello", "world"},
				Stdin: "",
			},
		},
		{
			name:      "command with error",
			cmd:       "rm",
			args:      []string{"-rf", "/protected"},
			configErr: errors.New(testPermissionDenied),
			expectErr: true,
			expectedCmd: ExecutedCommand{
				Name:  "rm",
				Args:  []string{"-rf", "/protected"},
				Stdin: "",
			},
		},
		{
			name:      "command without args",
			cmd:       "pwd",
			args:      nil,
			configErr: nil,
			expectErr: false,
			expectedCmd: ExecutedCommand{
				Name:  "pwd",
				Args:  nil,
				Stdin: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockExecutor()
			ctx := t.Context()

			key := makeKey(tt.cmd, tt.args...)
			if tt.configErr != nil {
				mock.SetError(key, tt.configErr)
			}

			err := mock.Run(ctx, tt.cmd, tt.args...)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Equal(t, tt.configErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			commands := mock.Commands()
			require.Len(t, commands, 1)
			assert.Equal(t, tt.expectedCmd.Name, commands[0].Name)
			assert.Equal(t, tt.expectedCmd.Args, commands[0].Args)
			assert.Equal(t, tt.expectedCmd.Stdin, commands[0].Stdin)
		})
	}
}

func TestMockExecutorRunWithOutput(t *testing.T) {
	tests := []struct {
		name           string
		cmd            string
		args           []string
		configOutput   string
		configErr      error
		expectErr      bool
		expectedOutput string
	}{
		{
			name:           "successful command with output",
			cmd:            "ls",
			args:           []string{"-la"},
			configOutput:   testFileListOutput,
			configErr:      nil,
			expectErr:      false,
			expectedOutput: testFileListOutput,
		},
		{
			name:           "command with error and output",
			cmd:            "ls",
			args:           []string{"/nonexistent"},
			configOutput:   "ls: cannot access '/nonexistent': No such file or directory",
			configErr:      errors.New("exit status 2"),
			expectErr:      true,
			expectedOutput: "ls: cannot access '/nonexistent': No such file or directory",
		},
		{
			name:           "unconfigured command returns empty",
			cmd:            "unknown",
			args:           nil,
			configOutput:   "",
			configErr:      nil,
			expectErr:      false,
			expectedOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockExecutor()
			ctx := t.Context()

			key := makeKey(tt.cmd, tt.args...)
			if tt.configOutput != "" {
				mock.SetOutput(key, tt.configOutput)
			}
			if tt.configErr != nil {
				mock.SetError(key, tt.configErr)
			}

			output, err := mock.RunWithOutput(ctx, tt.cmd, tt.args...)

			assert.Equal(t, tt.expectedOutput, output)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify command was recorded
			commands := mock.Commands()
			require.Len(t, commands, 1)
			assert.Equal(t, tt.cmd, commands[0].Name)
		})
	}
}

func TestMockExecutorRunWithStdin(t *testing.T) {
	tests := []struct {
		name          string
		stdin         string
		cmd           string
		args          []string
		configErr     error
		expectErr     bool
		expectedStdin string
	}{
		{
			name:          "command with stdin",
			stdin:         testInputData,
			cmd:           "cat",
			args:          nil,
			configErr:     nil,
			expectErr:     false,
			expectedStdin: testInputData,
		},
		{
			name:          "grep with stdin and pattern",
			stdin:         "line1\nmatch-me\nline3",
			cmd:           "grep",
			args:          []string{"match-me"},
			configErr:     nil,
			expectErr:     false,
			expectedStdin: "line1\nmatch-me\nline3",
		},
		{
			name:          "command with stdin and error",
			stdin:         "no match here",
			cmd:           "grep",
			args:          []string{"pattern"},
			configErr:     errors.New("exit status 1"),
			expectErr:     true,
			expectedStdin: "no match here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockExecutor()
			ctx := t.Context()

			key := makeKey(tt.cmd, tt.args...)
			if tt.configErr != nil {
				mock.SetError(key, tt.configErr)
			}

			err := mock.RunWithStdin(ctx, tt.stdin, tt.cmd, tt.args...)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify command was recorded with stdin
			commands := mock.Commands()
			require.Len(t, commands, 1)
			assert.Equal(t, tt.cmd, commands[0].Name)
			assert.Equal(t, tt.expectedStdin, commands[0].Stdin)
		})
	}
}

func TestMockExecutorThreadSafety(t *testing.T) {
	mock := NewMockExecutor()
	ctx := t.Context()
	const numGoroutines = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 4)

	// Concurrent SetOutput calls
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			mock.SetOutput("cmd"+string(rune('a'+i%26)), "output")
		}(i)
	}

	// Concurrent SetError calls
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			mock.SetError("err"+string(rune('a'+i%26)), errors.New("error"))
		}(i)
	}

	// Concurrent Run calls
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			// Error intentionally ignored in concurrent test - we only verify no panics/races
			_ = mock.Run(ctx, "echo", "hello") //nolint:errcheck // error irrelevant in concurrent stress test
		}()
	}

	// Concurrent Commands calls
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_ = mock.Commands()
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify no panic occurred and state is consistent
	commands := mock.Commands()
	assert.Len(t, commands, numGoroutines)
}

func TestMockExecutorThreadSafetyReset(t *testing.T) {
	mock := NewMockExecutor()
	ctx := t.Context()
	const numGoroutines = 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 3)

	// Concurrent operations while resetting
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			// Error intentionally ignored in concurrent test - we only verify no panics/races
			_ = mock.Run(ctx, "echo", "hello") //nolint:errcheck // error irrelevant in concurrent stress test
		}()
		go func() {
			defer wg.Done()
			mock.Reset()
		}()
		go func() {
			defer wg.Done()
			_ = mock.Commands()
		}()
	}

	wg.Wait()

	// Verify mock is still functional after concurrent operations
	commands := mock.Commands()
	assert.NotNil(t, commands, "Commands() should return non-nil slice after concurrent resets")

	// Verify subsequent Run calls still record commands
	require.NoError(t, mock.Run(ctx, "test", "arg"))
	commandsAfter := mock.Commands()
	assert.GreaterOrEqual(t, len(commandsAfter), 1, "Run should record commands after concurrent resets")
}

func TestMakeKey(t *testing.T) {
	tests := []struct {
		name     string
		cmdName  string
		args     []string
		expected string
	}{
		{
			name:     "command only",
			cmdName:  "ls",
			args:     nil,
			expected: "ls",
		},
		{
			name:     "command with empty args",
			cmdName:  "pwd",
			args:     []string{},
			expected: "pwd",
		},
		{
			name:     "command with one arg",
			cmdName:  "ls",
			args:     []string{"-la"},
			expected: "ls -la",
		},
		{
			name:     "command with multiple args",
			cmdName:  "docker",
			args:     []string{"run", "-d", "--name", "test", "nginx"},
			expected: "docker run -d --name test nginx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeKey(tt.cmdName, tt.args...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMockExecutorMultipleCommands(t *testing.T) {
	mock := NewMockExecutor()
	ctx := t.Context()

	// Configure different responses for different commands
	mock.SetOutput("echo hello", "hello\n")
	mock.SetOutput("ls -la", "file1.txt\nfile2.txt\n")
	mock.SetError("rm -rf /", errors.New("operation not permitted"))

	// Execute commands
	output1, err1 := mock.RunWithOutput(ctx, "echo", "hello")
	output2, err2 := mock.RunWithOutput(ctx, "ls", "-la")
	err3 := mock.Run(ctx, "rm", "-rf", "/")

	// Verify responses
	assert.Equal(t, "hello\n", output1)
	assert.NoError(t, err1)

	assert.Equal(t, "file1.txt\nfile2.txt\n", output2)
	assert.NoError(t, err2)

	assert.Error(t, err3)
	assert.Equal(t, "operation not permitted", err3.Error())

	// Verify all commands were recorded in order
	commands := mock.Commands()
	require.Len(t, commands, 3)
	assert.Equal(t, "echo", commands[0].Name)
	assert.Equal(t, "ls", commands[1].Name)
	assert.Equal(t, "rm", commands[2].Name)
}

func TestMockExecutorOverwriteConfiguration(t *testing.T) {
	mock := NewMockExecutor()

	// Set initial output
	mock.SetOutput("echo test", "first output")

	// Overwrite with new output
	mock.SetOutput("echo test", "second output")

	output, err := mock.RunWithOutput(t.Context(), "echo", "test")

	assert.NoError(t, err)
	assert.Equal(t, "second output", output)
}
