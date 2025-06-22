# Uber - Script Launcher

Uber is a tool launcher that helps you organize and run scripts and executables from your project. It automatically finds tools based on a configuration file and executes them with the provided arguments.

## Features

- **Automatic tool discovery**: Searches for executables in configured tool paths
- **Relative and absolute paths**: Supports both relative (project-root based) and absolute tool paths
- **Project root detection**: Automatically finds the project root by looking for a `.uber` file
- **Verbose mode**: Get detailed information about tool discovery and execution with colored output
- **Flexible configuration**: Simple TOML-based configuration
- **Colored output**: When running in a terminal, verbose output uses colors for better readability

## Installation

Build the tool from source:

```bash
go build -o uber ./cmd/uber
```

## Configuration

Create a `.uber` file in your project root with the following TOML format:

```toml
tool_paths = ["bin", "scripts", "/usr/local/bin", "./tools"]
```

### Tool Paths

- **Relative paths** (e.g., `"bin"`, `"scripts"`, `"./tools"`): Searched relative to the project root
- **Absolute paths** (e.g., `"/usr/local/bin"`, `"/opt/tools"`): Searched as-is

## Usage

### Basic Usage

```bash
# Run a tool (uber will search for 'my-tool' in configured paths)
uber my-tool

# Run a tool with arguments
uber my-tool arg1 arg2

# Run with verbose output (colored when in terminal)
uber --verbose my-tool arg1 arg2
```

### Command Line Options

- `--root <path>`: Specify the project root directory (default: auto-detect)
- `--verbose` or `-v`: Enable verbose output showing tool discovery process

### Colored Output

When running in a terminal, verbose mode uses colors to make output more readable:

- **🟢 Green**: Actions being performed (tool found, executing)
- **🟡 Yellow**: Warnings or non-critical issues (tool not found in specific path)
- **🔴 Red**: Errors (tool not found anywhere, configuration issues)

When output is redirected to a file or pipe, colors are automatically disabled.

### Examples

```bash
# From your project directory
uber hello world
# Searches for 'hello' in configured tool paths and runs it with 'world' as argument

# With verbose mode (colored output in terminal)
uber --verbose test-script arg1 arg2
# Shows which paths were searched and where the tool was found

# Specify project root
uber --root /path/to/project my-tool
```

## How It Works

1. **Project Root Detection**: Uber looks for a `.uber` file in the current directory or any parent directory
2. **Configuration Loading**: Loads the TOML configuration from the `.uber` file
3. **Tool Search**: Searches for the specified tool in each configured tool path (in order)
4. **Execution**: When found, executes the tool with the remaining arguments

## Example Project Structure

```
my-project/
├── .uber                    # Configuration file
├── bin/
│   ├── hello               # Executable script
│   └── build               # Build script
├── scripts/
│   ├── deploy              # Deployment script
│   └── test                # Test runner
└── tools/
    └── custom-tool         # Custom tool
```

Example `.uber` file:
```toml
tool_paths = ["bin", "scripts", "tools", "/usr/local/bin"]
```

## Error Handling

- **No `.uber` file found**: Error message with suggestion to create one
- **Tool not found**: Searches all configured paths and reports if not found
- **Non-executable file**: Skips files that aren't executable
- **Invalid configuration**: Reports TOML parsing errors

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o uber ./cmd/uber
``` 