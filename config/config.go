package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds the configuration from the .uber TOML file
type Config struct {
	ToolPaths    []string `toml:"tool_paths"`
	EnvSetup     string   `toml:"env_setup"`
	ReportingCmd string   `toml:"reporting_cmd"`
}

// Load loads the TOML configuration from an io.Reader
func Load(r io.Reader) (*Config, error) {
	// Parse the TOML data
	var config Config
	_, err := toml.NewDecoder(r).Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse .uber file: %w", err)
	}

	return &config, nil
}

// LoadFromFile loads the TOML configuration from the .uber file in the project root
func LoadFromFile(projectRoot string) (*Config, error) {
	uberFile := filepath.Join(projectRoot, ".uber")

	// Open the TOML file
	file, err := os.Open(uberFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read .uber file: %w", err)
	}
	defer file.Close()

	// Load the configuration
	return Load(file)
}
