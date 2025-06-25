# Uber - Script Launcher

Uber is a tool launcher that helps you organize and run scripts and executables from your project. It automatically finds tools based on a configuration file and executes them with the provided arguments.

## Installation

### Using Go Install (Recommended)

```bash
go install github.com/chaselatta/uber@latest
```

### Using Go Install with a specific version

```bash
go install github.com/chaselatta/uber@v1.0.0
```

### From Source

```bash
git clone https://github.com/chaselatta/uber.git
cd uber
go build -o uber ./cmd/uber
```

### Check Installation

```bash
uber --version
```

## Features

- **Automatic tool discovery**: Searches for executables in configured tool paths
- **Relative and absolute paths**: Supports both relative (project-root based) and absolute tool paths
- **Project root detection**: Automatically finds the project root by looking for a `.uber` file
- **List all available tools**: List all executable tools found in your configured tool paths with `--list-tools`
- **Verbose mode**: Get detailed information about tool discovery and execution with colored output
- **Flexible configuration**: Simple TOML-based configuration
- **Colored output**: When running in a terminal, verbose output uses colors for better readability

## Configuration

Create a `.uber` file in your project root with the following TOML format:

```toml
tool_paths = ["bin", "scripts", "/usr/local/bin", "./tools"]
```

### Environment Setup Script

You can define an environment setup script that will be executed before your tool is run. This is useful for setting up any environment variables that your tools might need.

The `env_setup` key in your `.uber` file specifies the path to this script (relative to the project root).

**Contract:**
- The script at `env_setup` must be an executable file.
- It can be written in any language (e.g., Shell, Python, Ruby).
- It must print environment variables to standard output, one per line, in `KEY=VALUE` format.

**Example `.uber` configuration:**

```toml
tool_paths = ["bin", "scripts"]
env_setup = "scripts/env_setup.sh"
reporting_cmd = "scripts/reporting.sh"
```

**Example `env_setup.sh`:**
```sh
#!/bin/sh
echo "MY_APP_NAME=My Awesome App"
echo "MY_APP_VERSION=1.2.3"
```

**Example `env_setup.py`:**
```python
#!/usr/bin/env python3
import os

print(f"MY_APP_NAME=My Awesome App")
print(f"MY_APP_VERSION=1.2.3")
```

The environment variables `MY_APP_NAME` and `MY_APP_VERSION` will be available to any tool executed by `uber`.

### Post-Execution Reporting

You can define a reporting command that will be executed after your tool has run. This is useful for sending metrics, notifications, or any other post-execution tasks.

The `reporting_cmd` key in your `.uber` file specifies the path to this script. The script can be in any language, as long as it's executable.

**Contract:**
- The script at `reporting_cmd` must be an executable file.
- It will be executed after the main tool finishes.
- The reporting command will have access to the following environment variables:
  - `UBER_EXECUTED_COMMAND`: The name of the tool that was executed.
  - `UBER_EXECUTED_TOOL_PATH`: The path where the executed tool was found.
  - `UBER_ARGS`: A string containing all the arguments passed to the tool.
  - `UBER_TIMING_FIND_TOOL_MS`: Time spent finding the tool (in milliseconds).
  - `UBER_TIMING_ENV_SETUP_MS`: Time spent in the `env_setup` script (in milliseconds).
  - `UBER_TIMING_EXECUTION_MS`: Time the tool spent executing (in milliseconds).
  - `UBER_TOTAL_TIME_MS`: Total time from tool search to execution completion.

**Example `reporting.sh`:**
```sh
#!/bin/sh
echo "--- Reporting ---"
echo "Tool: $UBER_EXECUTED_COMMAND"
echo "Args: $UBER_ARGS"
echo "Total time: $UBER_TOTAL_TIME_MS ms"
echo "--- End Reporting ---"
# You could also send these metrics to a server, a log file, etc.
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

# List all available tools
uber --list-tools

# Run with verbose output (colored when in terminal)
uber --verbose my-tool arg1 arg2
```

### Command Line Options

- `--root <path>`: Specify the project root directory (default: auto-detect)
- `--verbose` or `-v`: Enable verbose output showing tool discovery process
- `--list-tools`: List all available executable tools in the configured tool paths

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
env_setup = "scripts/env_setup.sh"
reporting_cmd = "scripts/reporting.sh"
```

## Error Handling

- **No `.uber`