package core

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestExecute_SimpleScript(t *testing.T) {
	scriptPath, err := filepath.Abs("testdata/echo.sh")
	if err != nil {
		t.Fatalf("failed to get script path: %v", err)
	}

	executor := NewExecutor()
	var stdout, stderr bytes.Buffer

	result, err := executor.Execute(context.Background(), ExecutionRequest{
		Script: Script{
			Name:  "echo-test",
			Path:  scriptPath,
			Shell: "bash",
		},
		Parameters: map[string]string{
			"message": "Hello World",
		},
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}

	output := stdout.String()
	if !strings.Contains(output, "Hello from echo-test") {
		t.Errorf("expected output to contain script name, got: %s", output)
	}
	if !strings.Contains(output, "Param MESSAGE: Hello World") {
		t.Errorf("expected output to contain parameter, got: %s", output)
	}
}

func TestExecute_ExitCode(t *testing.T) {
	scriptPath, err := filepath.Abs("testdata/exit_code.sh")
	if err != nil {
		t.Fatalf("failed to get script path: %v", err)
	}

	tests := []struct {
		name         string
		code         string
		expectedCode int
	}{
		{"exit 0", "0", 0},
		{"exit 1", "1", 1},
		{"exit 42", "42", 42},
		{"exit 255", "255", 255},
	}

	executor := NewExecutor()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			result, err := executor.Execute(context.Background(), ExecutionRequest{
				Script: Script{
					Name:  "exit-code-test",
					Path:  scriptPath,
					Shell: "bash",
				},
				Parameters: map[string]string{
					"code": tt.code,
				},
				Stdout: &stdout,
				Stderr: &stderr,
			})

			if err != nil {
				t.Fatalf("Execute returned error: %v", err)
			}

			if result.ExitCode != tt.expectedCode {
				t.Errorf("expected exit code %d, got %d", tt.expectedCode, result.ExitCode)
			}
		})
	}
}

func TestExecute_Stderr(t *testing.T) {
	scriptPath, err := filepath.Abs("testdata/stderr.sh")
	if err != nil {
		t.Fatalf("failed to get script path: %v", err)
	}

	executor := NewExecutor()
	var stdout, stderr bytes.Buffer

	result, err := executor.Execute(context.Background(), ExecutionRequest{
		Script: Script{
			Name:  "stderr-test",
			Path:  scriptPath,
			Shell: "bash",
		},
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}

	if !strings.Contains(stdout.String(), "stdout message") {
		t.Errorf("expected stdout to contain 'stdout message', got: %s", stdout.String())
	}

	if !strings.Contains(stderr.String(), "stderr message") {
		t.Errorf("expected stderr to contain 'stderr message', got: %s", stderr.String())
	}
}

func TestExecute_WorkingDirectory(t *testing.T) {
	scriptPath, err := filepath.Abs("testdata/workdir.sh")
	if err != nil {
		t.Fatalf("failed to get script path: %v", err)
	}

	executor := NewExecutor()

	t.Run("default to script directory", func(t *testing.T) {
		var stdout, stderr bytes.Buffer

		result, err := executor.Execute(context.Background(), ExecutionRequest{
			Script: Script{
				Name:  "workdir-test",
				Path:  scriptPath,
				Shell: "bash",
			},
			Stdout: &stdout,
			Stderr: &stderr,
		})

		if err != nil {
			t.Fatalf("Execute returned error: %v", err)
		}

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}

		expectedDir := filepath.Dir(scriptPath)
		output := strings.TrimSpace(stdout.String())
		if output != expectedDir {
			t.Errorf("expected working directory %s, got %s", expectedDir, output)
		}
	})

	t.Run("custom working directory", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		customDir := "/tmp"

		result, err := executor.Execute(context.Background(), ExecutionRequest{
			Script: Script{
				Name:  "workdir-test",
				Path:  scriptPath,
				Shell: "bash",
			},
			WorkDir: customDir,
			Stdout:  &stdout,
			Stderr:  &stderr,
		})

		if err != nil {
			t.Fatalf("Execute returned error: %v", err)
		}

		if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}

		// On macOS, /tmp is symlinked to /private/tmp
		output := strings.TrimSpace(stdout.String())
		if output != customDir && output != "/private/tmp" {
			t.Errorf("expected working directory %s or /private/tmp, got %s", customDir, output)
		}
	})
}

func TestExecute_ScriptNotFound(t *testing.T) {
	executor := NewExecutor()
	var stdout, stderr bytes.Buffer

	result, err := executor.Execute(context.Background(), ExecutionRequest{
		Script: Script{
			Name:  "nonexistent",
			Path:  "/nonexistent/script.sh",
			Shell: "bash",
		},
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if result.ExitCode != 127 {
		t.Errorf("expected exit code 127 for not found, got %d", result.ExitCode)
	}

	if result.Error == nil {
		t.Error("expected error to be set for not found script")
	}
}

func TestExecute_ContextCancellation(t *testing.T) {
	// Create a script that sleeps
	scriptContent := `#!/bin/bash
sleep 10
echo "should not reach here"
`
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "sleep.sh")
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	executor := NewExecutor()
	var stdout, stderr bytes.Buffer

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	result, err := executor.Execute(ctx, ExecutionRequest{
		Script: Script{
			Name:  "sleep-test",
			Path:  scriptPath,
			Shell: "bash",
		},
		Stdout: &stdout,
		Stderr: &stderr,
	})
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	// Should have been cancelled quickly, not waited 10 seconds
	if elapsed > 5*time.Second {
		t.Errorf("expected cancellation to be quick, took %v", elapsed)
	}

	if result.ExitCode != 130 {
		t.Errorf("expected exit code 130 for cancellation, got %d", result.ExitCode)
	}
}

func TestExecute_Timing(t *testing.T) {
	scriptPath, err := filepath.Abs("testdata/echo.sh")
	if err != nil {
		t.Fatalf("failed to get script path: %v", err)
	}

	executor := NewExecutor()
	var stdout, stderr bytes.Buffer

	result, err := executor.Execute(context.Background(), ExecutionRequest{
		Script: Script{
			Name:  "echo-test",
			Path:  scriptPath,
			Shell: "bash",
		},
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if result.StartTime.IsZero() {
		t.Error("expected StartTime to be set")
	}

	if result.EndTime.IsZero() {
		t.Error("expected EndTime to be set")
	}

	if result.EndTime.Before(result.StartTime) {
		t.Error("expected EndTime to be after StartTime")
	}
}

func TestGetInterpreter_Bash(t *testing.T) {
	e := &executor{}

	interp, args := e.getInterpreter(Script{
		Shell: "bash",
		Path:  "/nonexistent/script.sh", // Won't be read if Shell is set
	})

	// Since file doesn't exist, it will fall back to Shell
	if interp != "bash" {
		t.Errorf("expected interpreter 'bash', got '%s'", interp)
	}

	if len(args) != 0 {
		t.Errorf("expected no args, got %v", args)
	}
}

func TestGetInterpreter_Python(t *testing.T) {
	e := &executor{}

	interp, args := e.getInterpreter(Script{
		Shell: "python",
		Path:  "/nonexistent/script.py",
	})

	if interp != "python3" {
		t.Errorf("expected interpreter 'python3', got '%s'", interp)
	}

	if len(args) != 0 {
		t.Errorf("expected no args, got %v", args)
	}
}

func TestGetInterpreter_Shebang(t *testing.T) {
	// Create a script with a specific shebang
	scriptContent := `#!/usr/bin/env bash
echo "test"
`
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test.sh")
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	e := &executor{}
	interp, _ := e.getInterpreter(Script{
		Shell: "sh", // Should be overridden by shebang
		Path:  scriptPath,
	})

	if interp != "bash" {
		t.Errorf("expected interpreter 'bash' from shebang, got '%s'", interp)
	}
}

func TestReadShebang(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		expectedInterp  string
		expectedArgs    []string
	}{
		{
			name:            "simple bash",
			content:         "#!/bin/bash\necho test",
			expectedInterp:  "/bin/bash",
			expectedArgs:    []string{},
		},
		{
			name:            "env bash",
			content:         "#!/usr/bin/env bash\necho test",
			expectedInterp:  "bash",
			expectedArgs:    []string{},
		},
		{
			name:            "env python3",
			content:         "#!/usr/bin/env python3\nprint('test')",
			expectedInterp:  "python3",
			expectedArgs:    []string{},
		},
		{
			name:            "bash with args",
			content:         "#!/bin/bash -e\necho test",
			expectedInterp:  "/bin/bash",
			expectedArgs:    []string{"-e"},
		},
		{
			name:            "no shebang",
			content:         "echo test",
			expectedInterp:  "",
			expectedArgs:    nil,
		},
	}

	e := &executor{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			scriptPath := filepath.Join(tmpDir, "test.sh")
			if err := os.WriteFile(scriptPath, []byte(tt.content), 0755); err != nil {
				t.Fatalf("failed to write script: %v", err)
			}

			interp, args := e.readShebang(scriptPath)

			if interp != tt.expectedInterp {
				t.Errorf("expected interpreter '%s', got '%s'", tt.expectedInterp, interp)
			}

			if len(args) != len(tt.expectedArgs) {
				t.Errorf("expected %d args, got %d: %v", len(tt.expectedArgs), len(args), args)
			}
		})
	}
}

func TestBuildEnv_Parameters(t *testing.T) {
	e := &executor{}

	env := e.buildEnv(ExecutionRequest{
		Script: Script{
			Name: "test-script",
			Path: "/path/to/script.sh",
		},
		Parameters: map[string]string{
			"name":   "Alice",
			"count":  "42",
			"enable": "true",
		},
	})

	envMap := make(map[string]string)
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	// Check tap metadata
	if envMap["TAP_SCRIPT_NAME"] != "test-script" {
		t.Errorf("expected TAP_SCRIPT_NAME='test-script', got '%s'", envMap["TAP_SCRIPT_NAME"])
	}

	if envMap["TAP_SCRIPT_PATH"] != "/path/to/script.sh" {
		t.Errorf("expected TAP_SCRIPT_PATH='/path/to/script.sh', got '%s'", envMap["TAP_SCRIPT_PATH"])
	}

	// Check parameters
	if envMap["TAP_PARAM_NAME"] != "Alice" {
		t.Errorf("expected TAP_PARAM_NAME='Alice', got '%s'", envMap["TAP_PARAM_NAME"])
	}

	if envMap["TAP_PARAM_COUNT"] != "42" {
		t.Errorf("expected TAP_PARAM_COUNT='42', got '%s'", envMap["TAP_PARAM_COUNT"])
	}

	if envMap["TAP_PARAM_ENABLE"] != "true" {
		t.Errorf("expected TAP_PARAM_ENABLE='true', got '%s'", envMap["TAP_PARAM_ENABLE"])
	}
}

func TestBuildEnv_CustomEnv(t *testing.T) {
	e := &executor{}

	env := e.buildEnv(ExecutionRequest{
		Script: Script{
			Name: "test-script",
			Path: "/path/to/script.sh",
		},
		Env: map[string]string{
			"CUSTOM_VAR": "custom_value",
			"ANOTHER":    "another_value",
		},
	})

	envMap := make(map[string]string)
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	if envMap["CUSTOM_VAR"] != "custom_value" {
		t.Errorf("expected CUSTOM_VAR='custom_value', got '%s'", envMap["CUSTOM_VAR"])
	}

	if envMap["ANOTHER"] != "another_value" {
		t.Errorf("expected ANOTHER='another_value', got '%s'", envMap["ANOTHER"])
	}
}
