package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// These variables will be set by the linker during build
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
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

	// Handle version flag
	if ctx.ShowVersion {
		fmt.Printf("uber version %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("date: %s\n", date)
		return
	}

	// Create tool executor
	executor := NewToolExecutor(ctx)

	// Handle --list-tools flag
	if ctx.ListTools {
		if err := executor.ListAvailableTools(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Find and execute the tool
	if err := executor.FindAndExecuteTool(ctx.Command, ctx.RemainingArgs); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
