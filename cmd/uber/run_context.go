package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/chaselatta/uber/config"
	"github.com/spf13/pflag"
)

// RunContext holds all parsed command-line arguments and flags.
type RunContext struct {
	Root              string
	UberBinPath       string
	Verbose           bool
	ListTools         bool
	ShowVersion       bool
	Command           string
	RemainingArgs     []string
	GlobalCommandArgs string
	Config            *config.Config
	FoundToolPath     string
	TimeFindToolMs    int64
	TimeEnvSetupMs    int64
	TimeExecToolMs    int64
}

// findProjectRoot walks up the directory tree starting from the current working directory
// to find a directory containing a .uber file, which indicates the project root.
// Returns the absolute path to the project root, or an error if not found.
func findProjectRoot() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Walk up the directory tree
	for {
		// Check if .uber file exists in current directory
		uberFile := filepath.Join(currentDir, ".uber")
		if _, err := os.Stat(uberFile); err == nil {
			return currentDir, nil
		}

		// Get parent directory
		parentDir := filepath.Dir(currentDir)

		// If we've reached the root of the filesystem, stop
		if parentDir == currentDir {
			break
		}

		currentDir = parentDir
	}

	return "", fmt.Errorf("no .uber file found in current directory or any parent directories")
}

// validateProjectRoot checks if the specified directory contains a .uber file.
// Returns an error if the directory doesn't contain a .uber file or if the path is invalid.
func validateProjectRoot(rootPath string) error {
	// Check if the directory exists
	if _, err := os.Stat(rootPath); err != nil {
		return fmt.Errorf("specified root directory does not exist: %w", err)
	}

	// Check if .uber file exists in the specified directory
	uberFile := filepath.Join(rootPath, ".uber")
	if _, err := os.Stat(uberFile); err != nil {
		return fmt.Errorf("specified root directory does not contain a .uber file")
	}

	return nil
}

// ParseArgs parses flags and positional arguments into a RunContext struct.
// It takes an explicit args slice (excluding the program name) for testability.
// If --root is specified, it validates that the directory contains a .uber file.
// If no --root is specified, it automatically finds the project root by walking up
// the directory tree to find a directory containing a .uber file.
func ParseArgs(binPath string, args []string, output io.Writer) (*RunContext, error) {
	fs := pflag.NewFlagSet("uber", pflag.ContinueOnError)
	fs.SetInterspersed(false) // Stop parsing at the first non-flag argument

	root := fs.String("root", "", "Specify the root directory (e.g., --root /path/to/dir)")
	verbose := fs.BoolP("verbose", "v", false, "Enable verbose output (-v or --verbose)")
	listTools := fs.Bool("list-tools", false, "List available tools")
	showVersion := fs.Bool("version", false, "Show version information")

	if output == nil {
		output = os.Stderr
	}
	fs.SetOutput(output)

	// We don't want pflag to error on unknown flags, because they are for the script
	fs.ParseErrorsWhitelist.UnknownFlags = true

	// Parse the known uber flags
	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// The remaining args are for the script and tool
	remainingArgsForTool := fs.Args()

	// Find the command. It's the first positional argument that isn't a value
	// for a preceding flag.
	commandIndex := -1
	for i, arg := range remainingArgsForTool {
		if !strings.HasPrefix(arg, "-") {
			// If the previous arg was a flag, this is its value, not the command
			if i > 0 && strings.HasPrefix(remainingArgsForTool[i-1], "-") {
				continue
			}
			commandIndex = i
			break
		}
	}

	var command string
	var toolArgs []string

	if commandIndex != -1 {
		command = remainingArgsForTool[commandIndex]
		toolArgs = remainingArgsForTool[commandIndex+1:]
	}

	// Reconstruct the full string of global arguments passed to the uber command
	var globalCommandArgs string
	commandFound := false
	for _, arg := range args {
		if arg == command {
			commandFound = true
			break
		}
	}
	if commandFound {
		globalArgsEnd := -1
		for i, arg := range args {
			if arg == command {
				globalArgsEnd = i
				break
			}
		}
		if globalArgsEnd != -1 {
			globalCommandArgs = strings.Join(args[:globalArgsEnd], " ")
		}
	} else {
		globalCommandArgs = strings.Join(args, " ")
	}

	// Validate command presence
	if !(*listTools || *showVersion) && command == "" {
		return nil, fmt.Errorf("missing required positional argument 'command'")
	}
	if *listTools && command != "" {
		return nil, fmt.Errorf("--list-tools does not accept additional arguments: %s", command)
	}
	if *showVersion && command != "" {
		return nil, fmt.Errorf("--version does not accept additional arguments: %s", command)
	}

	// Validate project root
	projectRoot := *root
	if projectRoot != "" {
		if err := validateProjectRoot(projectRoot); err != nil {
			return nil, fmt.Errorf("invalid --root flag: %w", err)
		}
	} else {
		foundRoot, err := findProjectRoot()
		if err != nil {
			return nil, fmt.Errorf("failed to find project root: %w", err)
		}
		projectRoot = foundRoot
	}

	// Normalize the path to handle symlinks (important on macOS)
	projectRoot, err := filepath.EvalSymlinks(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate symlinks for project root: %w", err)
	}

	// Load config
	config, err := config.LoadFromFile(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	return &RunContext{
		Root:              projectRoot,
		UberBinPath:       binPath,
		Verbose:           *verbose,
		ListTools:         *listTools,
		ShowVersion:       *showVersion,
		Command:           command,
		RemainingArgs:     toolArgs,
		GlobalCommandArgs: globalCommandArgs,
		Config:            config,
	}, nil
}
