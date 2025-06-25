package uber

import (
	"fmt"
	"os"
	"path/filepath"
)

// These variables will be set by the linker during build
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

// Run executes the main uber logic
func Run() error {
	// Get the absolute path to the uber binary
	binPath, err := filepath.Abs(os.Args[0])
	if err != nil {
		return fmt.Errorf("error getting binary path: %w", err)
	}

	ctx, err := ParseArgs(binPath, os.Args[1:], nil)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	// Handle version flag
	if ctx.ShowVersion {
		fmt.Printf("uber version %s\n", Version)
		fmt.Printf("commit: %s\n", Commit)
		fmt.Printf("date: %s\n", Date)
		return nil
	}

	// Create tool executor
	executor := NewToolExecutor(ctx)

	// Handle --list-tools flag
	if ctx.ListTools {
		if err := executor.ListAvailableTools(); err != nil {
			return fmt.Errorf("error: %w", err)
		}
		return nil
	}

	// Find and execute the tool
	if err := executor.FindAndExecuteTool(ctx.Command, ctx.RemainingArgs); err != nil {
		return fmt.Errorf("error: %w", err)
	}

	return nil
}
