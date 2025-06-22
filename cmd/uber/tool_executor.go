package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ToolExecutor handles finding and executing tools based on the configuration
type ToolExecutor struct {
	ctx *RunContext
}

// NewToolExecutor creates a new ToolExecutor instance
func NewToolExecutor(ctx *RunContext) *ToolExecutor {
	return &ToolExecutor{
		ctx: ctx,
	}
}

// FindAndExecuteTool searches for the specified tool in the configured tool paths
// and executes it with the given arguments
func (te *ToolExecutor) FindAndExecuteTool(toolName string, args []string) error {
	// If no tool paths configured, return error
	if te.ctx.Config.ToolPaths == nil || len(te.ctx.Config.ToolPaths) == 0 {
		return fmt.Errorf("no tool paths configured in .uber file")
	}

	// Search for the tool in each configured path
	for _, toolPath := range te.ctx.Config.ToolPaths {
		executablePath, err := te.findExecutableInPath(toolPath, toolName)
		if err != nil {
			if te.ctx.Verbose {
				fmt.Printf("Tool '%s' not found in path '%s': %v\n", toolName, toolPath, err)
			}
			continue
		}

		// Found the executable, execute it
		if te.ctx.Verbose {
			fmt.Printf("Found tool '%s' at: %s\n", toolName, executablePath)
			fmt.Printf("Executing with args: %v\n", args)
		}

		return te.executeTool(executablePath, args)
	}

	return fmt.Errorf("tool '%s' not found in any configured tool path", toolName)
}

// findExecutableInPath looks for an executable with the given name in the specified path
func (te *ToolExecutor) findExecutableInPath(toolPath, toolName string) (string, error) {
	var fullPath string

	// If the path is relative, make it relative to the project root
	if !filepath.IsAbs(toolPath) {
		fullPath = filepath.Join(te.ctx.Root, toolPath)
	} else {
		fullPath = toolPath
	}

	// Construct the full path to the executable
	executablePath := filepath.Join(fullPath, toolName)

	// Check if the file exists and is executable
	fileInfo, err := os.Stat(executablePath)
	if err != nil {
		return "", fmt.Errorf("file not found: %w", err)
	}

	// Check if the file is executable
	if fileInfo.Mode()&0111 == 0 {
		return "", fmt.Errorf("file is not executable")
	}

	return executablePath, nil
}

// executeTool executes the tool with the given arguments
func (te *ToolExecutor) executeTool(executablePath string, args []string) error {
	// Create the command
	cmd := exec.Command(executablePath, args...)

	// Set up stdin, stdout, and stderr to be the same as the parent process
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Execute the command
	if te.ctx.Verbose {
		fmt.Printf("Executing: %s %v\n", executablePath, args)
	}

	return cmd.Run()
}
