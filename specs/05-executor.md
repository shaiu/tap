# Executor Spec

> Script execution engine

## Overview

The executor runs scripts with their parameters, streaming output in real-time. After execution, tap exits with the script's exit code.

## Core Behavior

1. **Real-time streaming** — stdout/stderr displayed as they occur
2. **Preserve PTY** — Interactive scripts work correctly (colors, prompts)
3. **Pass-through exit** — tap exits with script's exit code
4. **Parameter injection** — Pass parameters as environment variables or arguments

## Execution Flow

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Parameters  │────►│   Executor   │────►│   Script     │
│  (map)       │     │              │     │   Process    │
└──────────────┘     └──────┬───────┘     └──────┬───────┘
                            │                     │
                            │ stream              │ stdout/stderr
                            ▼                     ▼
                     ┌──────────────┐     ┌──────────────┐
                     │   Terminal   │◄────│   Real-time  │
                     │              │     │   Output     │
                     └──────────────┘     └──────────────┘
                            │
                            │ exit code
                            ▼
                     ┌──────────────┐
                     │   os.Exit()  │
                     └──────────────┘
```

## Data Structures

### ExecutionRequest

```go
type ExecutionRequest struct {
    Script     core.Script
    Parameters map[string]string
    WorkDir    string              // Working directory (default: script's dir)
    Env        map[string]string   // Additional env vars
    Stdin      io.Reader           // Input (default: os.Stdin)
    Stdout     io.Writer           // Output (default: os.Stdout)
    Stderr     io.Writer           // Errors (default: os.Stderr)
}
```

### ExecutionResult

```go
type ExecutionResult struct {
    Script    core.Script
    ExitCode  int
    StartTime time.Time
    EndTime   time.Time
    Error     error
}
```

## Executor Interface

```go
type Executor interface {
    // Execute runs a script and returns when complete
    Execute(ctx context.Context, req ExecutionRequest) (*ExecutionResult, error)
}
```

## Implementation

### TUI Exit Pattern

The key pattern: **exit TUI before executing script**. This ensures:
- Script gets full control of the terminal
- Interactive scripts work (prompts, passwords)
- Colors and formatting display correctly

```go
// In the app model
func (m AppModel) executeScript(script core.Script, params map[string]string) tea.Cmd {
    return tea.ExitAltScreen(func() tea.Msg {
        // Now we're out of the TUI
        result, err := m.executor.Execute(context.Background(), ExecutionRequest{
            Script:     script,
            Parameters: params,
            Stdout:     os.Stdout,
            Stderr:     os.Stderr,
            Stdin:      os.Stdin,
        })
        
        if err != nil {
            return ExecutionErrorMsg{Err: err}
        }
        return ExecutionCompleteMsg{Result: result}
    })
}

// After execution completes
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case ExecutionCompleteMsg:
        // Exit with script's exit code
        os.Exit(msg.Result.ExitCode)
        
    case ExecutionErrorMsg:
        fmt.Fprintf(os.Stderr, "tap: execution error: %v\n", msg.Err)
        os.Exit(1)
    }
    // ...
}
```

### Core Execution

```go
type executor struct{}

func NewExecutor() Executor {
    return &executor{}
}

func (e *executor) Execute(ctx context.Context, req ExecutionRequest) (*ExecutionResult, error) {
    result := &ExecutionResult{
        Script:    req.Script,
        StartTime: time.Now(),
    }
    
    // Determine interpreter
    interpreter, args := e.getInterpreter(req.Script)
    
    // Build command
    cmd := exec.CommandContext(ctx, interpreter, append(args, req.Script.Path)...)
    
    // Set working directory
    if req.WorkDir != "" {
        cmd.Dir = req.WorkDir
    } else {
        cmd.Dir = filepath.Dir(req.Script.Path)
    }
    
    // Set environment
    cmd.Env = e.buildEnv(req)
    
    // Connect I/O
    cmd.Stdin = req.Stdin
    cmd.Stdout = req.Stdout
    cmd.Stderr = req.Stderr
    
    // Run
    err := cmd.Run()
    result.EndTime = time.Now()
    
    // Extract exit code
    if err != nil {
        if exitErr, ok := err.(*exec.ExitError); ok {
            result.ExitCode = exitErr.ExitCode()
        } else {
            result.Error = err
            result.ExitCode = 1
        }
    }
    
    return result, nil
}
```

### Interpreter Detection

```go
func (e *executor) getInterpreter(script core.Script) (string, []string) {
    // First, try to read shebang
    if interp, args := e.readShebang(script.Path); interp != "" {
        return interp, args
    }
    
    // Fall back to extension-based detection
    switch script.Shell {
    case "python":
        return "python3", nil
    case "bash":
        return "bash", nil
    default:
        return "sh", nil
    }
}

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
            
            // Handle /usr/bin/env
            if strings.HasPrefix(shebang, "/usr/bin/env ") {
                parts := strings.Fields(shebang)
                if len(parts) >= 2 {
                    return parts[1], parts[2:]
                }
            }
            
            parts := strings.Fields(shebang)
            if len(parts) >= 1 {
                return parts[0], parts[1:]
            }
        }
    }
    
    return "", nil
}
```

### Environment Building

Parameters are passed as environment variables prefixed with `TAP_PARAM_`:

```go
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
```

### Script Access to Parameters

Scripts can access parameters via environment variables:

```bash
#!/bin/bash
# ---
# name: deploy
# parameters:
#   - name: environment
#     type: string
# ---

# Access via TAP_PARAM_* environment variables
echo "Deploying to $TAP_PARAM_ENVIRONMENT"

# Or use the variables directly if passed as args
```

Or for more complex cases, pass as arguments:

```go
func (e *executor) buildArgs(script core.Script, params map[string]string) []string {
    var args []string
    
    for _, param := range script.Parameters {
        if value, ok := params[param.Name]; ok {
            if param.Short != "" {
                args = append(args, fmt.Sprintf("-%s", param.Short), value)
            } else {
                args = append(args, fmt.Sprintf("--%s=%s", param.Name, value))
            }
        }
    }
    
    return args
}
```

## Headless Execution

For `tap run` in non-interactive mode:

```go
func RunHeadless(script core.Script, params map[string]string) int {
    executor := NewExecutor()
    
    result, err := executor.Execute(context.Background(), ExecutionRequest{
        Script:     script,
        Parameters: params,
        Stdout:     os.Stdout,
        Stderr:     os.Stderr,
        Stdin:      os.Stdin,
    })
    
    if err != nil {
        fmt.Fprintf(os.Stderr, "tap: %v\n", err)
        return 1
    }
    
    return result.ExitCode
}
```

## Context Cancellation

Support graceful shutdown on Ctrl+C:

```go
func (e *executor) Execute(ctx context.Context, req ExecutionRequest) (*ExecutionResult, error) {
    // ... setup ...
    
    // Handle signals
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    
    // Start process
    if err := cmd.Start(); err != nil {
        return nil, err
    }
    
    // Wait for completion or cancellation
    done := make(chan error, 1)
    go func() {
        done <- cmd.Wait()
    }()
    
    select {
    case <-ctx.Done():
        // Context cancelled - kill process
        cmd.Process.Signal(syscall.SIGTERM)
        select {
        case <-done:
            // Process exited
        case <-time.After(5 * time.Second):
            cmd.Process.Kill()
        }
        return nil, ctx.Err()
        
    case <-sigChan:
        // Forward signal to child
        cmd.Process.Signal(syscall.SIGTERM)
        err := <-done
        // ...
        
    case err := <-done:
        // Normal completion
        // ...
    }
}
```

## PTY Support (Optional Enhancement)

For full terminal emulation with interactive scripts:

```go
import "github.com/creack/pty"

func (e *executor) ExecuteWithPTY(ctx context.Context, req ExecutionRequest) (*ExecutionResult, error) {
    // ... build cmd ...
    
    // Start with PTY
    ptmx, err := pty.Start(cmd)
    if err != nil {
        // Fallback to regular execution
        return e.Execute(ctx, req)
    }
    defer ptmx.Close()
    
    // Handle terminal resize
    ch := make(chan os.Signal, 1)
    signal.Notify(ch, syscall.SIGWINCH)
    go func() {
        for range ch {
            if ws, err := pty.GetsizeFull(os.Stdin); err == nil {
                pty.Setsize(ptmx, ws)
            }
        }
    }()
    ch <- syscall.SIGWINCH // Initial size
    
    // Copy I/O
    go io.Copy(ptmx, os.Stdin)
    io.Copy(os.Stdout, ptmx)
    
    return // ...
}
```

## Error Handling

| Error | Behavior |
|-------|----------|
| Script not found | Exit with error message, code 127 |
| Permission denied | Exit with error message, code 126 |
| Invalid interpreter | Exit with error message, code 1 |
| Script error | Pass through exit code |
| Context cancelled | Exit code 130 (128 + SIGINT) |

## Messages

```go
type ExecutionCompleteMsg struct {
    Result *ExecutionResult
}

type ExecutionErrorMsg struct {
    Err error
}
```

## Testing

### Unit Tests

```go
func TestGetInterpreter_Bash(t *testing.T)
func TestGetInterpreter_Python(t *testing.T)
func TestGetInterpreter_Shebang(t *testing.T)
func TestBuildEnv_Parameters(t *testing.T)
func TestBuildEnv_Metadata(t *testing.T)
func TestReadShebang(t *testing.T)
```

### Integration Tests

```go
func TestExecute_SimpleScript(t *testing.T)
func TestExecute_WithParameters(t *testing.T)
func TestExecute_ExitCode(t *testing.T)
func TestExecute_Stderr(t *testing.T)
func TestExecute_WorkingDirectory(t *testing.T)
func TestExecute_ContextCancellation(t *testing.T)
```

### Test Scripts

```bash
# testdata/echo.sh
#!/bin/bash
echo "Hello from $TAP_SCRIPT_NAME"
echo "Param: $TAP_PARAM_MESSAGE"
exit 0

# testdata/exit_code.sh
#!/bin/bash
exit ${TAP_PARAM_CODE:-0}

# testdata/interactive.sh
#!/bin/bash
read -p "Enter name: " name
echo "Hello, $name"
```
