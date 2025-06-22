package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

// RunContext holds all parsed command-line arguments and flags.
type RunContext struct {
	Root          string
	Verbose       bool
	Command       string
	RemainingArgs []string
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

	return &RunContext{
		Root:          *root,
		Verbose:       *verbose,
		Command:       remaining[0],
		RemainingArgs: remaining[1:],
	}, nil
}
