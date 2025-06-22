package config

import (
	"fmt"
	"strings"
)

func ExampleLoad() {
	// Example TOML content as a string
	tomlContent := `tool_paths = ["/usr/local/bin", "bin", "tools", "/opt/tools"]`

	// Load configuration from string
	reader := strings.NewReader(tomlContent)
	config, err := Load(reader)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	// Print the loaded tool paths
	fmt.Printf("Tool paths: %v\n", config.ToolPaths)

	// Output:
	// Tool paths: [/usr/local/bin bin tools /opt/tools]
}
