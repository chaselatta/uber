package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	ctx, err := ParseArgs(os.Args[1:], nil)
	if err != nil {
		fmt.Println("Error:", err)
		flag.Usage()
		os.Exit(1)
	}

	fmt.Println("--- Parsed Arguments ---")
	fmt.Printf("Root Directory: %s\n", ctx.Root)
	fmt.Printf("Verbose Mode: %t\n", ctx.Verbose)
	fmt.Printf("Command: %s\n", ctx.Command)
	if len(ctx.RemainingArgs) > 0 {
		fmt.Printf("Remaining Args: %v\n", ctx.RemainingArgs)
	}
	fmt.Println("------------------------")
}
