package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
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
	findToolStart := time.Now()
	// Get all available tools
	availableTools, err := te.GetAllAvailableTools()
	if err != nil {
		return err
	}
	te.ctx.TimeFindToolMs = time.Since(findToolStart).Milliseconds()

	// Find the first occurrence of the tool (honoring tool_paths order)
	for _, tool := range availableTools {
		if tool.Name == toolName {
			// Found the tool, execute it
			if te.ctx.Verbose {
				ColorPrint(ColorGreen, fmt.Sprintf("Found tool '%s' in path '%s'\n", toolName, tool.Path))
				ColorPrint(ColorGreen, fmt.Sprintf("Executing with args: %v\n", args))
			}
			te.ctx.FoundToolPath = tool.Path

			// Execute the env setup script if it's defined
			envSetupStart := time.Now()
			env, err := te.executeEnvSetup()
			if err != nil {
				return fmt.Errorf("failed to execute env setup script: %w", err)
			}
			te.ctx.TimeEnvSetupMs = time.Since(envSetupStart).Milliseconds()

			// Construct the full path to the executable
			var fullPath string
			if !filepath.IsAbs(tool.Path) {
				fullPath = filepath.Join(te.ctx.Root, tool.Path)
			} else {
				fullPath = tool.Path
			}
			executablePath := filepath.Join(fullPath, toolName)

			execStart := time.Now()
			err = te.executeTool(executablePath, args, env)
			te.ctx.TimeExecToolMs = time.Since(execStart).Milliseconds()
			if err != nil {
				return err // Return original error
			}

			// After executing the tool, run the reporting command
			if reportErr := te.executeReportingCmd(); reportErr != nil {
				if te.ctx.Verbose {
					ColorPrint(ColorYellow, fmt.Sprintf("Warning: reporting command failed: %v\n", reportErr))
				}
				// Do not return this error, as the main tool succeeded
			}

			return nil
		}
	}

	return fmt.Errorf("tool '%s' not found in any configured tool path", toolName)
}

// executeEnvSetup executes the environment setup script if it is defined
// in the .uber configuration file and returns the resulting environment.
func (te *ToolExecutor) executeEnvSetup() ([]string, error) {
	if te.ctx.Config.EnvSetup == "" {
		return nil, nil // No script defined
	}

	// Resolve the script path
	scriptPath := te.ctx.Config.EnvSetup
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

	// Execute the script directly. It is expected to print environment variables
	// to stdout, one per line, in KEY=VALUE format.
	cmd := exec.Command(scriptPath)
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

	// The current environment
	currentEnv := te.prepareEnvironment()
	envMap := make(map[string]string)
	for _, v := range currentEnv {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	// Parse the output of the script and update the environment
	scanner := bufio.NewScanner(&stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key, value := parts[0], parts[1]
			envMap[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading env setup script output: %w", err)
	}

	// Convert the map back to a slice of strings
	var newEnv []string
	for key, value := range envMap {
		newEnv = append(newEnv, fmt.Sprintf("%s=%s", key, value))
	}

	return newEnv, nil
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

// executeReportingCmd runs the reporting command if it's defined in the .uber configuration
func (te *ToolExecutor) executeReportingCmd() error {
	if te.ctx.Config.ReportingCmd == "" {
		return nil // No reporting command defined
	}

	// Resolve the reporting command path
	executablePath := te.ctx.Config.ReportingCmd
	if !filepath.IsAbs(executablePath) {
		executablePath = filepath.Join(te.ctx.Root, executablePath)
	}

	// Check if the command exists and is executable
	if _, err := os.Stat(executablePath); os.IsNotExist(err) {
		return fmt.Errorf("reporting command '%s' not found", executablePath)
	}
	if !te.isExecutable(executablePath) {
		return fmt.Errorf("reporting command '%s' is not executable", executablePath)
	}

	// The reporting command doesn't take arguments from the command line
	cmd := exec.Command(executablePath)

	// The environment is prepared with additional reporting variables
	cmd.Env = te.prepareReportingEnvironment()

	// For reporting, we capture stdout and stderr to show in verbose mode,
	// but we don't want to pollute the main command's output.
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if te.ctx.Verbose {
		ColorPrint(ColorCyan, fmt.Sprintf("Executing reporting command: %s\n", executablePath))
		for _, envVar := range cmd.Env {
			if strings.HasPrefix(envVar, "UBER_TIMING") || strings.HasPrefix(envVar, "UBER_EXECUTED_") || strings.HasPrefix(envVar, "UBER_ARGS") {
				ColorPrint(ColorCyan, fmt.Sprintf("  %s\n", envVar))
			}
		}
	}

	err := cmd.Run()
	if te.ctx.Verbose && err != nil {
		ColorPrint(ColorYellow, fmt.Sprintf("Reporting command STDOUT: %s\n", stdout.String()))
		ColorPrint(ColorYellow, fmt.Sprintf("Reporting command STDERR: %s\n", stderr.String()))
	}

	if err != nil {
		return fmt.Errorf("error executing reporting command '%s': %w", executablePath, err)
	}

	return nil
}

// prepareReportingEnvironment creates the environment for the reporting command
func (te *ToolExecutor) prepareReportingEnvironment() []string {
	// Start with the base environment
	env := te.prepareEnvironment()

	totalTime := te.ctx.TimeFindToolMs + te.ctx.TimeEnvSetupMs + te.ctx.TimeExecToolMs

	// Add timing variables
	env = append(env,
		fmt.Sprintf("UBER_EXECUTED_COMMAND=%s", te.ctx.Command),
		fmt.Sprintf("UBER_EXECUTED_TOOL_PATH=%s", te.ctx.FoundToolPath),
		fmt.Sprintf("UBER_ARGS=%s", strings.Join(te.ctx.RemainingArgs, " ")),
		fmt.Sprintf("UBER_TIMING_FIND_TOOL_MS=%d", te.ctx.TimeFindToolMs),
		fmt.Sprintf("UBER_TIMING_ENV_SETUP_MS=%d", te.ctx.TimeEnvSetupMs),
		fmt.Sprintf("UBER_TIMING_EXECUTION_MS=%d", te.ctx.TimeExecToolMs),
		fmt.Sprintf("UBER_TOTAL_TIME_MS=%d", totalTime),
	)

	return env
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

func (te *ToolExecutor) resolveToolFullPath(toolPath, toolName string) string {
	if filepath.IsAbs(toolPath) {
		return filepath.Join(toolPath, toolName)
	}
	return filepath.Join(te.ctx.Root, toolPath, toolName)
}

func (te *ToolExecutor) findExecutableInPath(toolPath, toolName string) (string, error) {
	fullPath := te.resolveToolFullPath(toolPath, toolName)
	if te.isExecutable(fullPath) {
		return fullPath, nil
	}
	return "", fmt.Errorf("tool '%s' not found or not executable in '%s'", toolName, toolPath)
}

// findExecutable finds the executable for a given tool name in the configured tool paths
func (te *ToolExecutor) findExecutable(toolName string) (string, error) {
	// Handle absolute path case
	if filepath.IsAbs(toolName) {
		if te.isExecutable(toolName) {
			return toolName, nil
		}
		return "", fmt.Errorf("executable '%s' is not a valid executable file", toolName)
	}

	// Search in tool paths
	for _, toolPath := range te.ctx.Config.ToolPaths {
		var fullPath string
		if !filepath.IsAbs(toolPath) {
			fullPath = filepath.Join(te.ctx.Root, toolPath)
		} else {
			fullPath = toolPath
		}
		executablePath := filepath.Join(fullPath, toolName)

		if te.isExecutable(executablePath) {
			return executablePath, nil
		}
	}

	return "", fmt.Errorf("executable '%s' not found in any tool path", toolName)
}

// listExecutablesInPath scans a directory and returns a list of all executable files
func (te *ToolExecutor) listExecutablesInPath(toolPath string) ([]string, error) {
	var fullPath string
	if filepath.IsAbs(toolPath) {
		fullPath = toolPath
	} else {
		fullPath = filepath.Join(te.ctx.Root, toolPath)
	}

	files, err := os.ReadDir(fullPath)
	if err != nil {
		// Suppress error if path does not exist, as it's a common scenario
		if os.IsNotExist(err) {
			return nil, nil // Return empty list, don't propagate error
		}
		return nil, fmt.Errorf("failed to read directory '%s': %w", fullPath, err)
	}

	var executables []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		// Check if the file is executable
		if te.isExecutable(filepath.Join(fullPath, file.Name())) {
			executables = append(executables, file.Name())
		}
	}

	return executables, nil
}

// isExecutable checks if a file at the given path is an executable.
func (te *ToolExecutor) isExecutable(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	mode := info.Mode()
	return mode.IsRegular() && (mode&0111 != 0)
}
