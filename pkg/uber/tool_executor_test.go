package uber

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chaselatta/uber/config"
)

// createTempDirWithTool creates a temporary directory for tool execution tests
func createTempDirWithTool(t *testing.T, prefix string) (string, func()) {
	tempDir, err := os.MkdirTemp("", prefix)
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

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
	_, err = executor.findExecutableInPath(tempDir, "test-tool")
	if err == nil {
		t.Errorf("Expected error for non-executable file, got nil")
	}

	if !strings.Contains(err.Error(), "file exists but is not executable") {
		t.Errorf("Expected error message to contain 'file exists but is not executable', got: %v", err)
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
	// Create a temp project root
	tempDir, cleanup := createTempDirWithTool(t, "uber-test-not-found")
	defer cleanup()

	executor := &ToolExecutor{
		ctx: &RunContext{
			Root:    tempDir,
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
	expectedProjectRoot := tempDir // use the tempDir as root

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

		// Read the output and verify
		output, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}

		expectedContent := fmt.Sprintf("UBER_BIN_PATH=%s\nUBER_PROJECT_ROOT=%s\nUBER_VERBOSE=1\n",
			expectedBinPath, expectedProjectRoot)
		if string(output) != expectedContent {
			t.Errorf("Expected output:\n%s\nGot:\n%s", expectedContent, string(output))
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

		// Execute the tool
		err := executor.FindAndExecuteTool("env-writer-tool", []string{})
		if err != nil {
			t.Fatalf("Failed to execute test tool: %v", err)
		}

		// Read the output and verify
		output, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}

		expectedContent := fmt.Sprintf("UBER_BIN_PATH=%s\nUBER_PROJECT_ROOT=%s\n",
			expectedBinPath, expectedProjectRoot)
		if string(output) != expectedContent {
			t.Errorf("Expected output:\n%s\nGot:\n%s", expectedContent, string(output))
		}
	})
}

func TestExecuteWithEnvSetup(t *testing.T) {
	// Create a temp project root
	tempDir, cleanup := createTempDirWithTool(t, "uber-test-env-setup")
	defer cleanup()

	// Create an env setup script that prints KEY=VALUE to stdout
	setupScript := filepath.Join(tempDir, "setup.sh")
	setupScriptContent := `#!/bin/sh
echo 'MY_VAR=hello from script'
`
	err := os.WriteFile(setupScript, []byte(setupScriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create setup script: %v", err)
	}

	// Create a tool that will print the env var to a file
	outputFile := filepath.Join(tempDir, "output.txt")
	toolPath := filepath.Join(tempDir, "print_env_tool")
	toolContent := fmt.Sprintf(`#!/bin/sh
echo "MY_VAR is: $MY_VAR" > %s
`, outputFile)
	err = os.WriteFile(toolPath, []byte(toolContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	// Create RunContext with EnvSetup configured
	ctx := &RunContext{
		Root:    tempDir,
		Verbose: false, // set to false to not clutter test output
		Config: &config.Config{
			ToolPaths: []string{tempDir},
			EnvSetup:  setupScript,
		},
	}

	executor := NewToolExecutor(ctx)
	err = executor.FindAndExecuteTool("print_env_tool", []string{})
	if err != nil {
		t.Fatalf("FindAndExecuteTool failed: %v", err)
	}

	// Check the output file
	output, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	expectedOutput := "MY_VAR is: hello from script\n"
	if string(output) != expectedOutput {
		t.Errorf("Expected output '%s', got '%s'", expectedOutput, string(output))
	}
}

func TestExecuteWithPythonEnvSetup(t *testing.T) {
	// Create a temp project root
	tempDir, cleanup := createTempDirWithTool(t, "uber-test-python-env-setup")
	defer cleanup()

	// Create an env setup script that prints KEY=VALUE to stdout
	setupScript := filepath.Join(tempDir, "setup.py")
	setupScriptContent := `#!/usr/bin/env python3
print('MY_VAR=hello from python script')
`
	err := os.WriteFile(setupScript, []byte(setupScriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create setup script: %v", err)
	}

	// Create a tool that will print the env var to a file
	outputFile := filepath.Join(tempDir, "output.txt")
	toolPath := filepath.Join(tempDir, "print_env_tool")
	toolContent := fmt.Sprintf(`#!/bin/sh
echo "MY_VAR is: $MY_VAR" > %s
`, outputFile)
	err = os.WriteFile(toolPath, []byte(toolContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	// Create RunContext with EnvSetup configured
	ctx := &RunContext{
		Root:    tempDir,
		Verbose: false, // set to false to not clutter test output
		Config: &config.Config{
			ToolPaths: []string{tempDir},
			EnvSetup:  setupScript,
		},
	}

	executor := NewToolExecutor(ctx)
	err = executor.FindAndExecuteTool("print_env_tool", []string{})
	if err != nil {
		t.Fatalf("FindAndExecuteTool failed: %v", err)
	}

	// Check the output file
	output, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	expectedOutput := "MY_VAR is: hello from python script\n"
	if string(output) != expectedOutput {
		t.Errorf("Expected output '%s', got '%s'", expectedOutput, string(output))
	}
}

func TestResolveToolNameWithExtension(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "uber-test-extension-resolution")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files with different extensions
	testFiles := []string{
		"foo.sh",
		"foo.py",
		"bar.sh",
		"baz", // No extension
	}

	for _, fileName := range testFiles {
		filePath := filepath.Join(tempDir, fileName)
		if err := os.WriteFile(filePath, []byte("#!/bin/bash\necho 'test'"), 0755); err != nil {
			t.Fatalf("Failed to create test file %s: %v", fileName, err)
		}
	}

	executor := &ToolExecutor{
		ctx: &RunContext{
			Root:    "/test/project",
			Verbose: false,
			Config:  &config.Config{},
		},
	}

	// Test cases
	testCases := []struct {
		name        string
		requested   string
		expected    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "exact match with extension",
			requested:   "foo.sh",
			expected:    "foo.sh",
			expectError: false,
		},
		{
			name:        "exact match without extension",
			requested:   "baz",
			expected:    "baz",
			expectError: false,
		},
		{
			name:        "ambiguous - multiple extensions",
			requested:   "foo",
			expectError: true,
			errorMsg:    "ambiguous tool name",
		},
		{
			name:        "not found",
			requested:   "nonexistent",
			expectError: true,
			errorMsg:    "not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := executor.resolveToolName(tempDir, tc.requested)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
					return
				}
				if tc.errorMsg != "" && !strings.Contains(err.Error(), tc.errorMsg) {
					t.Errorf("Expected error to contain '%s', got: %v", tc.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
					return
				}
				if result != tc.expected {
					t.Errorf("Expected '%s', got '%s'", tc.expected, result)
				}
			}
		})
	}
}

func TestResolveToolNamePriority(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "uber-test-extension-priority")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files: one without extension, one with extension
	files := []string{"foo", "foo.sh"}

	for _, fileName := range files {
		filePath := filepath.Join(tempDir, fileName)
		if err := os.WriteFile(filePath, []byte("#!/bin/bash\necho 'test'"), 0755); err != nil {
			t.Fatalf("Failed to create test file %s: %v", fileName, err)
		}
	}

	executor := &ToolExecutor{
		ctx: &RunContext{
			Root:    "/test/project",
			Verbose: false,
			Config:  &config.Config{},
		},
	}

	// When requesting "foo", it should prefer the file without extension
	result, err := executor.resolveToolName(tempDir, "foo")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if result != "foo" {
		t.Errorf("Expected 'foo' (no extension), got '%s'", result)
	}
}

func TestGetAllAvailableToolsWithExtensions(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "uber-test-available-tools")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files with extensions
	testFiles := []string{
		"tool1.sh",
		"tool1.py",
		"tool2.sh",
		"tool3", // No extension
	}

	for _, fileName := range testFiles {
		filePath := filepath.Join(tempDir, fileName)
		if err := os.WriteFile(filePath, []byte("#!/bin/bash\necho 'test'"), 0755); err != nil {
			t.Fatalf("Failed to create test file %s: %v", fileName, err)
		}
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

	tools, err := executor.GetAllAvailableTools()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should return all 4 tools since users can run any of them directly
	expectedCount := 4
	if len(tools) != expectedCount {
		t.Errorf("Expected %d tools, got %d", expectedCount, len(tools))
	}

	// Check that all tools are present
	expectedTools := map[string]bool{
		"tool1.sh": false,
		"tool1.py": false,
		"tool2.sh": false,
		"tool3":    false,
	}

	for _, tool := range tools {
		expectedTools[tool.Name] = true
	}

	for toolName, found := range expectedTools {
		if !found {
			t.Errorf("Expected tool '%s' to be in available tools list", toolName)
		}
	}
}
