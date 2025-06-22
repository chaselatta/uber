package main

import (
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
	_, err = executor.findExecutableInPath(tempDir, "test-tool")
	if err == nil {
		t.Errorf("Expected error for non-executable file, got nil")
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
