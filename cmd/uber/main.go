package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	ctx, err := ParseArgs(os.Args[1:], nil)
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
