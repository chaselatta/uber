package main

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/chaselatta/uber/config"
)

// createTempDirWithUberFile creates a temporary directory with a .uber TOML file
// and returns the directory path and a cleanup function
func createTempDirWithUberFile(t *testing.T, prefix string) (string, func()) {
	tempDir, err := os.MkdirTemp("", prefix)
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create .uber TOML file in temp directory
	uberFile := filepath.Join(tempDir, ".uber")
	tomlContent := `tool_paths = ["/usr/local/bin", "bin", "tools", "/opt/tools", "./scripts"]`
	if err := os.WriteFile(uberFile, []byte(tomlContent), 0644); err != nil {
		t.Fatalf("Failed to create .uber file: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    *RunContext
		wantErr bool
		setup   func() (string, func()) // setup function returns temp dir and cleanup function
	}{
		{
			name: "all flags and command with valid root",
			args: []string{"--root", "/tmp", "--verbose", "start", "foo", "bar"},
			want: &RunContext{
				Root:          "/tmp",
				Verbose:       true,
				Command:       "start",
				RemainingArgs: []string{"foo", "bar"},
				Config: &config.Config{
					ToolPaths: []string{"/usr/local/bin", "bin", "tools", "/opt/tools", "./scripts"},
				},
			},
			wantErr: false,
			setup: func() (string, func()) {
				return createTempDirWithUberFile(t, "uber-test-valid-root")
			},
		},
		{
			name:    "empty",
			args:    []string{},
			want:    nil,
			wantErr: true,
			setup:   nil,
		},
		{
			name: "root not parsed twice",
			args: []string{"--root", "/tmp", "--verbose", "start", "--root", "foo"},
			want: &RunContext{
				Root:          "/tmp",
				Verbose:       true,
				Command:       "start",
				RemainingArgs: []string{"--root", "foo"},
				Config: &config.Config{
					ToolPaths: []string{"/usr/local/bin", "bin", "tools", "/opt/tools", "./scripts"},
				},
			},
			wantErr: false,
			setup: func() (string, func()) {
				return createTempDirWithUberFile(t, "uber-test-valid-root")
			},
		},
		{
			name:    "missing command",
			args:    []string{"--root", "/tmp"},
			want:    nil,
			wantErr: true,
			setup:   nil,
		},
		{
			name:    "unknown flag",
			args:    []string{"--unknown", "start"},
			want:    nil,
			wantErr: true,
			setup:   nil,
		},
		{
			name:    "invalid root directory does not exist",
			args:    []string{"--root", "/nonexistent/directory", "start"},
			want:    nil,
			wantErr: true,
			setup:   nil,
		},
		{
			name:    "invalid root directory missing .uber file",
			args:    []string{"--root", "/tmp", "start"},
			want:    nil,
			wantErr: true,
			setup:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tempDir string
			var cleanup func()

			if tt.setup != nil {
				tempDir, cleanup = tt.setup()
				defer cleanup()

				// Update the test args to use the actual temp directory path
				for i, arg := range tt.args {
					if arg == "/tmp" {
						tt.args[i] = tempDir
					}
				}

				// Update the expected result to use the actual temp directory path
				if tt.want != nil && tt.want.Root == "/tmp" {
					tt.want.Root = tempDir
				}
			}

			got, err := ParseArgs(tt.args, io.Discard)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseArgs() = %+v, want %+v", got, tt.want)
			}

			// Check specific error messages for validation failures
			if tt.wantErr && err != nil {
				if tt.name == "invalid root directory does not exist" {
					if err.Error()[:len("invalid --root flag: specified root directory does not exist")] != "invalid --root flag: specified root directory does not exist" {
						t.Errorf("Expected error about directory not existing, got: %v", err)
					}
				} else if tt.name == "invalid root directory missing .uber file" {
					if err.Error()[:len("invalid --root flag: specified root directory does not contain a .uber file")] != "invalid --root flag: specified root directory does not contain a .uber file" {
						t.Errorf("Expected error about missing .uber file, got: %v", err)
					}
				}
			}
		})
	}
}

func TestParseArgsWithAutoRoot(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "uber-test-parse")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create nested directory structure
	subDir1 := filepath.Join(tempDir, "subdir1")
	subDir2 := filepath.Join(subDir1, "subdir2")

	if err := os.MkdirAll(subDir2, 0755); err != nil {
		t.Fatalf("Failed to create nested directories: %v", err)
	}

	// Create .uber TOML file in subdir1 (project root)
	uberFile := filepath.Join(subDir1, ".uber")
	tomlContent := `tool_paths = ["/usr/local/bin", "bin", "tools", "/opt/tools", "./scripts"]`
	if err := os.WriteFile(uberFile, []byte(tomlContent), 0644); err != nil {
		t.Fatalf("Failed to create .uber file: %v", err)
	}

	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get original working directory: %v", err)
	}

	// Change to subdir2
	if err := os.Chdir(subDir2); err != nil {
		t.Fatalf("Failed to change to subdir2: %v", err)
	}
	defer os.Chdir(originalWd)

	// Test ParseArgs without --root flag
	args := []string{"test-command", "arg1", "arg2"}
	ctx, err := ParseArgs(args, nil)
	if err != nil {
		t.Fatalf("ParseArgs failed: %v", err)
	}

	// Verify the root was found automatically
	expectedRoot, err := filepath.Abs(subDir1)
	if err != nil {
		t.Fatalf("Failed to get absolute path of expected root: %v", err)
	}
	// Normalize the path to handle symlinks (important on macOS)
	expectedRoot, err = filepath.EvalSymlinks(expectedRoot)
	if err != nil {
		t.Fatalf("Failed to evaluate symlinks for expected root: %v", err)
	}

	if ctx.Root != expectedRoot {
		t.Errorf("Expected root %s, got %s", expectedRoot, ctx.Root)
	}

	if ctx.Command != "test-command" {
		t.Errorf("Expected command 'test-command', got '%s'", ctx.Command)
	}

	expectedRemainingArgs := []string{"arg1", "arg2"}
	if !reflect.DeepEqual(ctx.RemainingArgs, expectedRemainingArgs) {
		t.Errorf("Expected remaining args %v, got %v", expectedRemainingArgs, ctx.RemainingArgs)
	}

	// Verify the configuration was loaded
	expectedConfig := &config.Config{
		ToolPaths: []string{"/usr/local/bin", "bin", "tools", "/opt/tools", "./scripts"},
	}
	if !reflect.DeepEqual(ctx.Config, expectedConfig) {
		t.Errorf("Expected config %+v, got %+v", expectedConfig, ctx.Config)
	}
}

func TestParseArgsWithoutAutoRoot(t *testing.T) {
	// Test that ParseArgs fails when no root is specified and no .uber file exists
	args := []string{"test-command"}
	_, err := ParseArgs(args, nil)
	if err == nil {
		t.Error("Expected error when no root is specified and no .uber file exists, but got nil")
	}

	expectedErrorPrefix := "failed to find project root:"
	if err.Error()[:len(expectedErrorPrefix)] != expectedErrorPrefix {
		t.Errorf("Expected error message to start with '%s', got '%s'", expectedErrorPrefix, err.Error())
	}
}

func TestFindProjectRoot(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "uber-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create nested directory structure
	subDir1 := filepath.Join(tempDir, "subdir1")
	subDir2 := filepath.Join(subDir1, "subdir2")
	subDir3 := filepath.Join(subDir2, "subdir3")

	if err := os.MkdirAll(subDir3, 0755); err != nil {
		t.Fatalf("Failed to create nested directories: %v", err)
	}

	// Create .uber TOML file in subdir1 (project root)
	uberFile := filepath.Join(subDir1, ".uber")
	tomlContent := `tool_paths = ["/usr/local/bin", "bin", "tools", "/opt/tools", "./scripts"]`
	if err := os.WriteFile(uberFile, []byte(tomlContent), 0644); err != nil {
		t.Fatalf("Failed to create .uber file: %v", err)
	}

	// Test finding project root from subdir3
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get original working directory: %v", err)
	}

	// Change to subdir3
	if err := os.Chdir(subDir3); err != nil {
		t.Fatalf("Failed to change to subdir3: %v", err)
	}
	defer os.Chdir(originalWd)

	// Find project root
	foundRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("findProjectRoot failed: %v", err)
	}

	// Verify we found the correct root
	expectedRoot, err := filepath.Abs(subDir1)
	if err != nil {
		t.Fatalf("Failed to get absolute path of expected root: %v", err)
	}
	// Normalize the path to handle symlinks (important on macOS)
	expectedRoot, err = filepath.EvalSymlinks(expectedRoot)
	if err != nil {
		t.Fatalf("Failed to evaluate symlinks for expected root: %v", err)
	}

	if foundRoot != expectedRoot {
		t.Errorf("Expected project root %s, got %s", expectedRoot, foundRoot)
	}
}

func TestFindProjectRootNotFound(t *testing.T) {
	// Create a temporary directory without .uber file
	tempDir, err := os.MkdirTemp("", "uber-test-no-root")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get original working directory: %v", err)
	}

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	defer os.Chdir(originalWd)

	// Try to find project root
	_, err = findProjectRoot()
	if err == nil {
		t.Error("Expected error when no .uber file is found, but got nil")
	}

	expectedError := "no .uber file found in current directory or any parent directories"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

func TestValidateProjectRoot(t *testing.T) {
	tests := []struct {
		name     string
		rootPath string
		setup    func() (string, func())
		wantErr  bool
	}{
		{
			name:     "valid project root with .uber file",
			rootPath: "/tmp",
			setup: func() (string, func()) {
				return createTempDirWithUberFile(t, "uber-test-valid")
			},
			wantErr: false,
		},
		{
			name:     "directory does not exist",
			rootPath: "/nonexistent/directory",
			setup:    nil,
			wantErr:  true,
		},
		{
			name:     "directory exists but no .uber file",
			rootPath: "/tmp",
			setup: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "uber-test-no-uber")
				if err != nil {
					t.Fatalf("Failed to create temp directory: %v", err)
				}

				cleanup := func() {
					os.RemoveAll(tempDir)
				}

				return tempDir, cleanup
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tempDir string
			var cleanup func()

			if tt.setup != nil {
				tempDir, cleanup = tt.setup()
				defer cleanup()

				// Update the test to use the actual temp directory path
				if tt.rootPath == "/tmp" {
					tt.rootPath = tempDir
				}
			}

			err := validateProjectRoot(tt.rootPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateProjectRoot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check specific error messages
			if tt.wantErr && err != nil {
				if tt.name == "directory does not exist" {
					if err.Error()[:len("specified root directory does not exist")] != "specified root directory does not exist" {
						t.Errorf("Expected error about directory not existing, got: %v", err)
					}
				} else if tt.name == "directory exists but no .uber file" {
					if err.Error() != "specified root directory does not contain a .uber file" {
						t.Errorf("Expected error about missing .uber file, got: %v", err)
					}
				}
			}
		})
	}
}
