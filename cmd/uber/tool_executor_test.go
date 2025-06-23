package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/chaselatta/uber/config"
)

func TestNewToolExecutor(t *testing.T) {
	ctx := &RunContext{
		Root:    "/test/project",
		Verbose: true,
		Config: &config.Config{
			ToolPaths: []string{"/usr/local/bin", "bin"},
		},
	}

	executor := NewToolExecutor(ctx)

	if executor.ctx != ctx {
		t.Errorf("Expected ctx to be set, got %v", executor.ctx)
	}
}

func TestFindExecutableInPath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "uber-test-executable")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test executable
	testExecutable := filepath.Join(tempDir, "test-tool")
	if err := os.WriteFile(testExecutable, []byte("#!/bin/bash\necho 'test'"), 0755); err != nil {
		t.Fatalf("Failed to create test executable: %v", err)
	}

	executor := &ToolExecutor{
		ctx: &RunContext{
			Root:    "/test/project",
			Verbose: false,
			Config:  &config.Config{},
		},
	}

	// Test finding executable with absolute path
	executablePath, err := executor.findExecutableInPath(tempDir, "test-tool")
	if err != nil {
		t.Errorf("Expected to find executable, got error: %v", err)
	}
	if executablePath != filepath.Join(tempDir, "test-tool") {
		t.Errorf("Expected executable path to be %s, got %s",
			filepath.Join(tempDir, "test-tool"), executablePath)
	}

	// Test finding executable with relative path
	relativePath := "bin"
	executablePath, err = executor.findExecutableInPath(relativePath, "test-tool")
	if err == nil {
		t.Errorf("Expected error when executable doesn't exist, got nil")
	}
}

func TestFindExecutableInPathNonExecutable(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "uber-test-non-executable")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a non-executable file
	testFile := filepath.Join(tempDir, "test-tool")
	if err := os.WriteFile(testFile, []byte("not executable"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	executor := &ToolExecutor{
		ctx: &RunContext{
			Root:    "/test/project",
			Verbose: false,
			Config:  &config.Config{},
		},
	}

	// Test finding non-executable file
	// Note: findExecutableInPath now only checks for file existence, not executability
	// The executability check will happen later when exec.Command is called
	executablePath, err := executor.findExecutableInPath(tempDir, "test-tool")
	if err != nil {
		t.Errorf("Expected to find file (executability check happens later), got error: %v", err)
	}
	if executablePath != filepath.Join(tempDir, "test-tool") {
		t.Errorf("Expected executable path to be %s, got %s",
			filepath.Join(tempDir, "test-tool"), executablePath)
	}
}

func TestFindAndExecuteToolNoToolPaths(t *testing.T) {
	executor := &ToolExecutor{
		ctx: &RunContext{
			Root:    "/test/project",
			Verbose: false,
			Config:  &config.Config{ToolPaths: nil},
		},
	}

	err := executor.FindAndExecuteTool("test-tool", []string{})
	if err == nil {
		t.Errorf("Expected error when no tool paths configured, got nil")
	}
}

func TestFindAndExecuteToolEmptyToolPaths(t *testing.T) {
	executor := &ToolExecutor{
		ctx: &RunContext{
			Root:    "/test/project",
			Verbose: false,
			Config:  &config.Config{ToolPaths: []string{}},
		},
	}

	err := executor.FindAndExecuteTool("test-tool", []string{})
	if err == nil {
		t.Errorf("Expected error when tool paths is empty, got nil")
	}
}

func TestFindAndExecuteToolNotFound(t *testing.T) {
	executor := &ToolExecutor{
		ctx: &RunContext{
			Root:    "/test/project",
			Verbose: false,
			Config: &config.Config{
				ToolPaths: []string{"/nonexistent/path", "/another/nonexistent"},
			},
		},
	}

	err := executor.FindAndExecuteTool("nonexistent-tool", []string{})
	if err == nil {
		t.Errorf("Expected error when tool not found, got nil")
	}
}

func TestExecuteNonExecutableFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "uber-test-non-executable-execution")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a non-executable file
	testFile := filepath.Join(tempDir, "test-tool")
	if err := os.WriteFile(testFile, []byte("not executable"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	executor := &ToolExecutor{
		ctx: &RunContext{
			Root:    "/test/project",
			Verbose: false,
			Config: &config.Config{
				ToolPaths: []string{tempDir},
			},
		},
	}

	// Test that execution fails when trying to run a non-executable file
	err = executor.FindAndExecuteTool("test-tool", []string{})
	if err == nil {
		t.Errorf("Expected error when trying to execute non-executable file, got nil")
	}
}

func TestExecuteToolWithEnvironmentVariables(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "uber-test-env-vars")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a temporary file to capture output
	outputFile := filepath.Join(tempDir, "output.txt")

	// Create a test executable that writes environment variables to a file
	envTestExecutable := filepath.Join(tempDir, "env-writer-tool")
	envWriterScript := fmt.Sprintf(`#!/bin/bash
echo "UBER_BIN_PATH=$UBER_BIN_PATH" > %s
echo "UBER_PROJECT_ROOT=$UBER_PROJECT_ROOT" >> %s
if [ -n "$UBER_VERBOSE" ]; then
    echo "UBER_VERBOSE=$UBER_VERBOSE" >> %s
fi
`, outputFile, outputFile, outputFile)

	if err := os.WriteFile(envTestExecutable, []byte(envWriterScript), 0755); err != nil {
		t.Fatalf("Failed to create env writer executable: %v", err)
	}

	expectedBinPath := "/usr/local/bin/uber"
	expectedProjectRoot := "/test/project"

	// Test case 1: Verbose is true
	t.Run("VerboseTrue", func(t *testing.T) {
		// Clean up output file
		os.Remove(outputFile)

		executor := &ToolExecutor{
			ctx: &RunContext{
				Root:        expectedProjectRoot,
				UberBinPath: expectedBinPath,
				Verbose:     true,
				Config: &config.Config{
					ToolPaths: []string{tempDir},
				},
			},
		}

		// Execute the tool that writes environment variables to a file
		err := executor.FindAndExecuteTool("env-writer-tool", []string{})
		if err != nil {
			t.Fatalf("Failed to execute test tool: %v", err)
		}

		// Read the output file to verify environment variables
		output, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}

		outputStr := string(output)

		// Check that all expected environment variables are present
		expectedVars := map[string]string{
			"UBER_BIN_PATH":     expectedBinPath,
			"UBER_PROJECT_ROOT": expectedProjectRoot,
			"UBER_VERBOSE":      "1",
		}

		for varName, expectedValue := range expectedVars {
			expectedLine := fmt.Sprintf("%s=%s", varName, expectedValue)
			if !contains(outputStr, expectedLine) {
				t.Errorf("Expected environment variable line '%s' not found in output:\n%s", expectedLine, outputStr)
			}
		}
	})

	// Test case 2: Verbose is false
	t.Run("VerboseFalse", func(t *testing.T) {
		// Clean up output file
		os.Remove(outputFile)

		executor := &ToolExecutor{
			ctx: &RunContext{
				Root:        expectedProjectRoot,
				UberBinPath: expectedBinPath,
				Verbose:     false,
				Config: &config.Config{
					ToolPaths: []string{tempDir},
				},
			},
		}

		// Execute the tool that writes environment variables to a file
		err := executor.FindAndExecuteTool("env-writer-tool", []string{})
		if err != nil {
			t.Fatalf("Failed to execute test tool: %v", err)
		}

		// Read the output file to verify environment variables
		output, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}

		outputStr := string(output)

		// Check that UBER_VERBOSE is NOT present when verbose is false
		if contains(outputStr, "UBER_VERBOSE=") {
			t.Errorf("UBER_VERBOSE should not be set when verbose is false, but found in output:\n%s", outputStr)
		}

		// Check that other environment variables are still present
		expectedVars := map[string]string{
			"UBER_BIN_PATH":     expectedBinPath,
			"UBER_PROJECT_ROOT": expectedProjectRoot,
		}

		for varName, expectedValue := range expectedVars {
			expectedLine := fmt.Sprintf("%s=%s", varName, expectedValue)
			if !contains(outputStr, expectedLine) {
				t.Errorf("Expected environment variable line '%s' not found in output:\n%s", expectedLine, outputStr)
			}
		}
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
