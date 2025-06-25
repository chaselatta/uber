package main

import (
	"fmt"
	"os"

	"github.com/chaselatta/uber/pkg/uber"
)

func main() {
	if err := uber.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
