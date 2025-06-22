package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// RunContext holds all parsed command-line arguments and flags.
type RunContext struct {
	Root          string
	Verbose       bool
	Command       string
	RemainingArgs []string
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

// ParseArgs parses flags and positional arguments into a RunContext struct.
// It takes an explicit args slice (excluding the program name) for testability.
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

	// If no root is specified, try to find the project root automatically
	projectRoot := *root
	if projectRoot == "" {
		foundRoot, err := findProjectRoot()
		if err != nil {
			return nil, fmt.Errorf("failed to find project root: %w", err)
		}
		projectRoot = foundRoot
	}

	return &RunContext{
		Root:          projectRoot,
		Verbose:       *verbose,
		Command:       remaining[0],
		RemainingArgs: remaining[1:],
	}, nil
}
