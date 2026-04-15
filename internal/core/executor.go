package core

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/mattn/go-isatty"
)

// ExecutionRequest contains all inputs needed to execute a script.
type ExecutionRequest struct {
	Script     Script
	Parameters map[string]string
	WorkDir    string            // Working directory (default: caller's cwd)
	Env        map[string]string // Additional env vars
	Stdin      io.Reader         // Input (default: os.Stdin)
	Stdout     io.Writer         // Output (default: os.Stdout)
	Stderr     io.Writer         // Errors (default: os.Stderr)
}

// ExecutionResult contains the outcome of a script execution.
type ExecutionResult struct {
	Script    Script
	ExitCode  int
	StartTime time.Time
	EndTime   time.Time
	Error     error
}

// Executor executes scripts.
type Executor interface {
	// Execute runs a script and returns when complete.
	Execute(ctx context.Context, req ExecutionRequest) (*ExecutionResult, error)
}

// executor is the default implementation of Executor.
type executor struct{}

// NewExecutor creates a new Executor.
func NewExecutor() Executor {
	return &executor{}
}

// Execute runs a script with the given request parameters.
func (e *executor) Execute(ctx context.Context, req ExecutionRequest) (*ExecutionResult, error) {
	result := &ExecutionResult{
		Script:    req.Script,
		StartTime: time.Now(),
	}

	// Validate script path exists
	if _, err := os.Stat(req.Script.Path); os.IsNotExist(err) {
		result.EndTime = time.Now()
		result.ExitCode = 127 // Command not found
		result.Error = fmt.Errorf("script not found: %s", req.Script.Path)
		return result, nil
	}

	// Determine interpreter
	interpreter, args := e.getInterpreter(req.Script)

	// Build command args: interpreter args + script path
	cmdArgs := append(args, req.Script.Path)
	cmd := exec.CommandContext(ctx, interpreter, cmdArgs...)

	// Set working directory (default: caller's working directory)
	if req.WorkDir != "" {
		cmd.Dir = req.WorkDir
	}

	// Set environment
	cmd.Env = e.buildEnv(req)

	// Connect I/O (use defaults if not specified)
	if req.Stdin != nil {
		cmd.Stdin = req.Stdin
	} else {
		cmd.Stdin = os.Stdin
	}
	if req.Stdout != nil {
		cmd.Stdout = req.Stdout
	} else {
		cmd.Stdout = os.Stdout
	}
	if req.Stderr != nil {
		cmd.Stderr = req.Stderr
	} else {
		cmd.Stderr = os.Stderr
	}

	// Set up process group for signal handling.
	// For interactive scripts running in a TTY, use Foreground mode so the
	// child becomes the foreground process group and can read from the terminal.
	// Without this, interactive tools (gum, kubectl exec -it, etc.) receive
	// SIGTTIN and hang because they're in a background process group.
	if req.Script.Interactive && isatty.IsTerminal(os.Stdin.Fd()) {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Foreground: true,
			Ctty:       int(os.Stdin.Fd()),
		}
	} else {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		result.EndTime = time.Now()
		if os.IsPermission(err) {
			result.ExitCode = 126 // Permission denied
			result.Error = fmt.Errorf("permission denied: %s", req.Script.Path)
		} else {
			result.ExitCode = 1
			result.Error = fmt.Errorf("failed to start script: %w", err)
		}
		return result, nil
	}

	// Wait for completion in a goroutine
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// Handle completion or context cancellation
	select {
	case <-ctx.Done():
		// Context cancelled - terminate gracefully
		if cmd.Process != nil {
			// Send SIGTERM first for graceful shutdown
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
			select {
			case <-done:
				// Process exited
			case <-time.After(2 * time.Second):
				// Force kill after timeout
				_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
				<-done
			}
		}
		result.EndTime = time.Now()
		result.ExitCode = 130 // 128 + SIGINT
		result.Error = ctx.Err()
		return result, nil

	case err := <-done:
		// Normal completion (or context cancellation handled by CommandContext)
		result.EndTime = time.Now()

		// Check if this was due to context cancellation
		if ctx.Err() != nil {
			result.ExitCode = 130
			result.Error = ctx.Err()
			return result, nil
		}

		result.ExitCode = e.extractExitCode(err)
		if err != nil && result.ExitCode == 0 {
			// Non-exit error (shouldn't happen, but handle it)
			result.ExitCode = 1
			result.Error = err
		}
		return result, nil
	}
}

// getInterpreter determines the interpreter and args for a script.
func (e *executor) getInterpreter(script Script) (string, []string) {
	// First, try to read shebang from the script
	if interp, args := e.readShebang(script.Path); interp != "" {
		return interp, args
	}

	// Fall back to extension/shell-based detection
	switch script.Shell {
	case "python":
		return "python3", nil
	case "bash":
		return "bash", nil
	default:
		return "sh", nil
	}
}

// readShebang reads and parses the shebang line from a script file.
func (e *executor) readShebang(path string) (string, []string) {
	file, err := os.Open(path)
	if err != nil {
		return "", nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#!") {
			shebang := strings.TrimPrefix(line, "#!")
			shebang = strings.TrimSpace(shebang)

			// Handle /usr/bin/env style shebangs
			if strings.HasPrefix(shebang, "/usr/bin/env ") {
				parts := strings.Fields(shebang)
				if len(parts) >= 2 {
					return parts[1], parts[2:]
				}
			}

			// Direct interpreter path
			parts := strings.Fields(shebang)
			if len(parts) >= 1 {
				return parts[0], parts[1:]
			}
		}
	}

	return "", nil
}

// buildEnv constructs the environment variables for script execution.
func (e *executor) buildEnv(req ExecutionRequest) []string {
	// Start with current environment
	env := os.Environ()

	// Add tap metadata
	env = append(env, fmt.Sprintf("TAP_SCRIPT_NAME=%s", req.Script.Name))
	env = append(env, fmt.Sprintf("TAP_SCRIPT_PATH=%s", req.Script.Path))

	// Add parameters as TAP_PARAM_<NAME>
	for name, value := range req.Parameters {
		envName := fmt.Sprintf("TAP_PARAM_%s", strings.ToUpper(name))
		env = append(env, fmt.Sprintf("%s=%s", envName, value))
	}

	// Add custom env vars
	for k, v := range req.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return env
}

// extractExitCode gets the exit code from an exec error.
func (e *executor) extractExitCode(err error) int {
	if err == nil {
		return 0
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode()
	}
	return 1
}
