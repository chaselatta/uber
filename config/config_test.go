package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		tomlContent string
		want        *Config
		wantErr     bool
	}{
		{
			name: "valid tool_paths and env_setup",
			tomlContent: `tool_paths = ["/usr/local/bin", "bin"]
env_setup = "/path/to/setup.sh"`,
			want: &Config{
				ToolPaths: []string{"/usr/local/bin", "bin"},
				EnvSetup:  "/path/to/setup.sh",
			},
			wantErr: false,
		},
		{
			name:        "valid_tool_paths_with_mixed_relative_and_absolute",
			tomlContent: `tool_paths = ["/usr/local/bin", "bin", "tools", "/opt/tools", "./scripts", "../external-tools"]`,
			want: &Config{
				ToolPaths: []string{"/usr/local/bin", "bin", "tools", "/opt/tools", "./scripts", "../external-tools"},
			},
			wantErr: false,
		},
		{
			name:        "only absolute paths",
			tomlContent: `tool_paths = ["/usr/local/bin", "/opt/tools", "/home/user/bin"]`,
			want: &Config{
				ToolPaths: []string{"/usr/local/bin", "/opt/tools", "/home/user/bin"},
			},
			wantErr: false,
		},
		{
			name:        "only relative paths",
			tomlContent: `tool_paths = ["bin", "tools", "./scripts", "../external"]`,
			want: &Config{
				ToolPaths: []string{"bin", "tools", "./scripts", "../external"},
			},
			wantErr: false,
		},
		{
			name:        "empty tool_paths",
			tomlContent: `tool_paths = []`,
			want: &Config{
				ToolPaths: []string{},
			},
			wantErr: false,
		},
		{
			name:        "missing tool_paths",
			tomlContent: `# No tool_paths specified`,
			want: &Config{
				ToolPaths: nil,
			},
			wantErr: false,
		},
		{
			name:        "valid env setup only",
			tomlContent: `env_setup = "scripts/setup.sh"`,
			want: &Config{
				EnvSetup: "scripts/setup.sh",
			},
			wantErr: false,
		},
		{
			name:        "empty_env_setup",
			tomlContent: `env_setup = ""`,
			want: &Config{
				EnvSetup: "",
			},
			wantErr: false,
		},
		{
			name:        "missing_env_setup",
			tomlContent: `tool_paths = ["/usr/bin"]`,
			want: &Config{
				ToolPaths: []string{"/usr/bin"},
				EnvSetup:  "",
			},
			wantErr: false,
		},
		{
			name:        "malformed_toml",
			tomlContent: `tool_paths = [`,
			want:        nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Load with string reader
			reader := strings.NewReader(tt.tomlContent)
			got, err := Load(reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Load() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestLoadFromFile(t *testing.T) {
	// Test LoadFromFile with a temporary file
	tomlContent := `
tool_paths = ["/usr/local/bin", "bin", "tools"]
env_setup = "/path/to/env.sh"
`
	expectedConfig := &Config{
		ToolPaths: []string{"/usr/local/bin", "bin", "tools"},
		EnvSetup:  "/path/to/env.sh",
	}

	// Create temporary directory with .uber file
	tempDir, err := os.MkdirTemp("", "uber-test-config-file")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .uber file with test content
	uberFile := filepath.Join(tempDir, ".uber")
	if err := os.WriteFile(uberFile, []byte(tomlContent), 0644); err != nil {
		t.Fatalf("Failed to create .uber file: %v", err)
	}

	// Test LoadFromFile
	got, err := LoadFromFile(tempDir)
	if err != nil {
		t.Errorf("LoadFromFile() error = %v", err)
		return
	}
	if !reflect.DeepEqual(got, expectedConfig) {
		t.Errorf("LoadFromFile() = %+v, want %+v", got, expectedConfig)
	}
}

func TestLoadFromFileNotFound(t *testing.T) {
	// Create temporary directory without .uber file
	tempDir, err := os.MkdirTemp("", "uber-test-no-file")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test LoadFromFile with non-existent file
	_, err = LoadFromFile(tempDir)
	if err == nil {
		t.Error("Expected error when .uber file does not exist, but got nil")
	}

	expectedErrorPrefix := "failed to read .uber file:"
	if err.Error()[:len(expectedErrorPrefix)] != expectedErrorPrefix {
		t.Errorf("Expected error message to start with '%s', got '%s'", expectedErrorPrefix, err.Error())
	}
}
