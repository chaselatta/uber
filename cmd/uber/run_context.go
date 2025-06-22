package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/chaselatta/uber/config"
)

// RunContext holds all parsed command-line arguments and flags.
type RunContext struct {
	Root          string
	Verbose       bool
	Command       string
	RemainingArgs []string
	Config        *config.Config
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
func ParseArgs(args []string, output io.Writer) (*RunContext, error) {
	fs := flag.NewFlagSet("uber", flag.ContinueOnError)
	root := fs.String("root", "", "Specify the root directory (e.g., --root /path/to/dir)")
	verbose := fs.Bool("verbose", false, "Enable verbose output (-v or --verbose)")
	fs.BoolVar(verbose, "v", false, "Enable verbose output (shorthand for --verbose)")

	if output == nil {
		output = os.Stderr
	}
	fs.SetOutput(output)

	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	remaining := fs.Args()
	if len(remaining) == 0 {
		return nil, fmt.Errorf("missing required positional argument 'command'")
	}

	// If root is specified, validate that it contains a .uber file
	projectRoot := *root
	if projectRoot != "" {
		if err := validateProjectRoot(projectRoot); err != nil {
			return nil, fmt.Errorf("invalid --root flag: %w", err)
		}
	} else {
		// If no root is specified, try to find the project root automatically
		foundRoot, err := findProjectRoot()
		if err != nil {
			return nil, fmt.Errorf("failed to find project root: %w", err)
		}
		projectRoot = foundRoot
	}

	// Load the TOML configuration
	config, err := config.LoadFromFile(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	return &RunContext{
		Root:          projectRoot,
		Verbose:       *verbose,
		Command:       remaining[0],
		RemainingArgs: remaining[1:],
		Config:        config,
	}, nil
}
