package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// AvailableTool represents a tool that can be executed
type AvailableTool struct {
	Name string
	Path string
}

// GetAllAvailableTools scans all configured tool paths and returns all executable tools
// in the order they appear in the tool_paths configuration
func (te *ToolExecutor) GetAllAvailableTools() ([]AvailableTool, error) {
	// If no tool paths configured, return error
	if te.ctx.Config.ToolPaths == nil || len(te.ctx.Config.ToolPaths) == 0 {
		return nil, fmt.Errorf("no tool paths configured in .uber file")
	}

	var allTools []AvailableTool

	// Search for tools in each configured path in order
	for _, toolPath := range te.ctx.Config.ToolPaths {
		tools, err := te.listExecutablesInPath(toolPath)
		if err != nil {
			if te.ctx.Verbose {
				ColorPrint(ColorYellow, fmt.Sprintf("Error scanning path '%s': %v\n", toolPath, err))
			}
			continue
		}

		// Add tools from this path to the list
		for _, toolName := range tools {
			allTools = append(allTools, AvailableTool{
				Name: toolName,
				Path: toolPath,
			})
		}
	}

	return allTools, nil
}

// FindAndExecuteTool searches for the specified tool in the configured tool paths
// and executes it with the given arguments
func (te *ToolExecutor) FindAndExecuteTool(toolName string, args []string) error {
	// Get all available tools
	availableTools, err := te.GetAllAvailableTools()
	if err != nil {
		return err
	}

	// Find the first occurrence of the tool (honoring tool_paths order)
	for _, tool := range availableTools {
		if tool.Name == toolName {
			// Found the tool, execute it
			if te.ctx.Verbose {
				ColorPrint(ColorGreen, fmt.Sprintf("Found tool '%s' in path '%s'\n", toolName, tool.Path))
				ColorPrint(ColorGreen, fmt.Sprintf("Executing with args: %v\n", args))
			}

			// Execute the env setup script if it's defined
			env, err := te.executeEnvSetupScript()
			if err != nil {
				return fmt.Errorf("failed to execute env setup script: %w", err)
			}

			// Construct the full path to the executable
			var fullPath string
			if !filepath.IsAbs(tool.Path) {
				fullPath = filepath.Join(te.ctx.Root, tool.Path)
			} else {
				fullPath = tool.Path
			}
			executablePath := filepath.Join(fullPath, toolName)

			return te.executeTool(executablePath, args, env)
		}
	}

	return fmt.Errorf("tool '%s' not found in any configured tool path", toolName)
}

// executeEnvSetupScript executes the environment setup script if it is defined
// in the .uber configuration file and returns the resulting environment.
func (te *ToolExecutor) executeEnvSetupScript() ([]string, error) {
	if te.ctx.Config.EnvSetupScript == "" {
		return nil, nil // No script defined
	}

	// Resolve the script path
	scriptPath := te.ctx.Config.EnvSetupScript
	if !filepath.IsAbs(scriptPath) {
		scriptPath = filepath.Join(te.ctx.Root, scriptPath)
	}

	// Check if the script exists and is executable
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("script '%s' not found", scriptPath)
	}
	if !te.isExecutable(scriptPath) {
		return nil, fmt.Errorf("script '%s' is not executable", scriptPath)
	}

	// Create the command to execute the script and capture its environment.
	// We source the script and then run `env` to get all environment variables.
	commandStr := fmt.Sprintf(". %s && env", scriptPath)
	cmd := exec.Command("sh", "-c", commandStr)

	cmd.Env = te.prepareEnvironment()

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if te.ctx.Verbose {
		ColorPrint(ColorCyan, fmt.Sprintf("Executing env setup script: %s\n", scriptPath))
	}

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("error executing env setup script '%s': %w", scriptPath, err)
	}

	// Parse the output of `env` into a slice of strings, filtering for valid env vars
	var validEnvVars []string
	scanner := bufio.NewScanner(&stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "=") {
			validEnvVars = append(validEnvVars, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading env setup script output: %w", err)
	}

	return validEnvVars, nil
}

// executeTool executes the tool with the given arguments
func (te *ToolExecutor) executeTool(executablePath string, args []string, env []string) error {
	// Create the command
	cmd := exec.Command(executablePath, args...)

	// Set up stdin, stdout, and stderr to be the same as the parent process
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set environment variables for context
	if env != nil {
		cmd.Env = env
	} else {
		cmd.Env = te.prepareEnvironment()
	}

	// Execute the command
	if te.ctx.Verbose {
		ColorPrint(ColorGreen, fmt.Sprintf("Executing: %s %v\n", executablePath, args))
		ColorPrint(ColorGreen, fmt.Sprintf("UBER_BIN_PATH=%s\n", te.ctx.UberBinPath))
		ColorPrint(ColorGreen, fmt.Sprintf("UBER_PROJECT_ROOT=%s\n", te.ctx.Root))
	}

	return cmd.Run()
}

// prepareEnvironment creates the environment variables for tool execution
func (te *ToolExecutor) prepareEnvironment() []string {
	env := append(os.Environ(),
		fmt.Sprintf("UBER_BIN_PATH=%s", te.ctx.UberBinPath),
		fmt.Sprintf("UBER_PROJECT_ROOT=%s", te.ctx.Root),
	)

	// Only set UBER_VERBOSE if verbose is true
	if te.ctx.Verbose {
		env = append(env, "UBER_VERBOSE=1")
	}

	// Add global command arguments if they exist
	if te.ctx.GlobalCommandArgs != "" {
		env = append(env, fmt.Sprintf("UBER_GLOBAL_COMMAND_ARGS=%s", te.ctx.GlobalCommandArgs))
	}

	return env
}

// ListAvailableTools scans all configured tool paths and lists all executable tools
func (te *ToolExecutor) ListAvailableTools() error {
	// Get all available tools
	availableTools, err := te.GetAllAvailableTools()
	if err != nil {
		return err
	}

	fmt.Println("Available tools:")
	fmt.Println()

	// Group tools by path for display
	currentPath := ""
	for _, tool := range availableTools {
		if tool.Path != currentPath {
			if currentPath != "" {
				fmt.Println()
			}
			ColorPrint(ColorCyan, fmt.Sprintf("From %s:\n", tool.Path))
			currentPath = tool.Path
		}
		fmt.Printf("  %s\n", tool.Name)
	}

	return nil
}

// resolveToolFullPath resolves the full path to a tool given a toolPath and toolName
func (te *ToolExecutor) resolveToolFullPath(toolPath, toolName string) string {
	var fullPath string
	if !filepath.IsAbs(toolPath) {
		fullPath = filepath.Join(te.ctx.Root, toolPath)
	} else {
		fullPath = toolPath
	}
	return filepath.Join(fullPath, toolName)
}

// findExecutableInPath checks if a specific executable exists in the given path
func (te *ToolExecutor) findExecutableInPath(toolPath, toolName string) (string, error) {
	executablePath := te.resolveToolFullPath(toolPath, toolName)

	// Check if the file exists
	if _, err := os.Stat(executablePath); err != nil {
		return "", fmt.Errorf("executable not found: %w", err)
	}

	// Check if the file is executable
	if !te.isExecutable(executablePath) {
		return "", fmt.Errorf("file exists but is not executable")
	}

	return executablePath, nil
}

// listExecutablesInPath lists all executable files in the specified path
func (te *ToolExecutor) listExecutablesInPath(toolPath string) ([]string, error) {
	var fullPath string
	if !filepath.IsAbs(toolPath) {
		fullPath = filepath.Join(te.ctx.Root, toolPath)
	} else {
		fullPath = toolPath
	}

	// Check if the directory exists
	if _, err := os.Stat(fullPath); err != nil {
		return nil, fmt.Errorf("directory not found: %w", err)
	}

	// Read the directory
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var executables []string
	for _, entry := range entries {
		if !entry.IsDir() {
			candidatePath := te.resolveToolFullPath(toolPath, entry.Name())
			if te.isExecutable(candidatePath) {
				executables = append(executables, entry.Name())
			}
		}
	}

	return executables, nil
}

// isExecutable checks if a file is executable
func (te *ToolExecutor) isExecutable(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	// Check if the file has executable permissions
	mode := info.Mode()
	return mode.IsRegular() && (mode&0111 != 0)
}
