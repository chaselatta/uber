package main

import (
	"fmt"
	"os"
)

// ANSI color codes
const (
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorRed    = "\033[31m"
	ColorCyan   = "\033[36m"
	ColorReset  = "\033[0m"
)

// IsTTY checks if stdout is connected to a terminal
func IsTTY() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// IsTTYStderr checks if stderr is connected to a terminal
func IsTTYStderr() bool {
	fileInfo, _ := os.Stderr.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// ColorPrint prints colored text only if running in a TTY
func ColorPrint(color, message string) {
	if IsTTY() {
		fmt.Print(color + message + ColorReset)
	} else {
		fmt.Print(message)
	}
}

// ColorPrintError prints colored error text only if running in a TTY
func ColorPrintError(message string) {
	if IsTTYStderr() {
		fmt.Fprint(os.Stderr, ColorRed+message+ColorReset)
	} else {
		fmt.Fprint(os.Stderr, message)
	}
}
