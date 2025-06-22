package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// Get the absolute path to the uber binary
	binPath, err := filepath.Abs(os.Args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting binary path: %v\n", err)
		os.Exit(1)
	}

	ctx, err := ParseArgs(binPath, os.Args[1:], nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	// Create tool executor
	executor := NewToolExecutor(ctx)

	// Find and execute the tool
	if err := executor.FindAndExecuteTool(ctx.Command, ctx.RemainingArgs); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
