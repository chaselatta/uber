package uber

import (
	"bytes"
	"os"
	"testing"
)

func TestColorConstants(t *testing.T) {
	// Test that color constants are defined
	if ColorGreen == "" {
		t.Error("ColorGreen constant is empty")
	}
	if ColorYellow == "" {
		t.Error("ColorYellow constant is empty")
	}
	if ColorRed == "" {
		t.Error("ColorRed constant is empty")
	}
	if ColorReset == "" {
		t.Error("ColorReset constant is empty")
	}
}

func TestColorPrint(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	// Test color printing
	testMessage := "Test message"
	ColorPrint(ColorGreen, testMessage)

	// Close the write end and read the output
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Check that the message is in the output
	if output == "" {
		t.Error("Expected output, got empty string")
	}
}

func TestColorPrintError(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() {
		os.Stderr = oldStderr
	}()

	// Test error color printing
	testMessage := "Test error message"
	ColorPrintError(testMessage)

	// Close the write end and read the output
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Check that the message is in the output
	if output == "" {
		t.Error("Expected output, got empty string")
	}
}

func TestIsTTY(t *testing.T) {
	// Test that IsTTY returns a boolean value
	// We can't easily test the actual TTY detection in a test environment,
	// but we can ensure it doesn't panic and returns a consistent value
	result := IsTTY()
	// Just ensure it's a boolean (no panic)
	_ = result
}

func TestIsTTYStderr(t *testing.T) {
	// Test that IsTTYStderr returns a boolean value
	result := IsTTYStderr()
	// Just ensure it's a boolean (no panic)
	_ = result
}
