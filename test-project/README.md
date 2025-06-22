# Uber Test Project

This is a test project for the Uber tool that demonstrates various features including the new environment variables functionality.

## Project Structure

```
test-project/
├── .uber              # Uber configuration file
├── bin/               # Executable tools
│   └── hello          # Simple hello script
├── scripts/           # Script tools
│   ├── test-script    # Basic test script
│   └── env-demo       # Environment variables demo
└── README.md          # This file
```

## Configuration

The `.uber` file configures the tool paths:

```toml
tool_paths = ["bin", "scripts", "/usr/local/bin"]
```

## Available Tools

### bin/hello
A simple script that demonstrates basic argument passing and environment variables.

**Usage:**
```bash
uber hello [arguments]
```

**Features:**
- Displays received arguments
- Shows Uber environment variables
- Shows current working directory

### scripts/test-script
A more comprehensive script that demonstrates environment variable usage.

**Usage:**
```bash
uber test-script [arguments]
```

**Features:**
- Displays Uber environment variables
- Shows practical examples of using the variables
- Demonstrates path construction

### scripts/env-demo
A comprehensive demo script that showcases all environment variable features.

**Usage:**
```bash
uber env-demo
```

**Features:**
- Validates environment variables
- Shows project structure
- Demonstrates practical use cases
- Provides examples for script development

## Environment Variables

When Uber executes a tool, it provides the following environment variables:

### UBER_BIN_PATH
The absolute path to the Uber binary that was originally invoked.

**Example:**
```bash
UBER_BIN_PATH=/usr/local/bin/uber
```

**Use cases:**
- Call the Uber binary again from within a script
- Reference the Uber installation location
- Build toolchains that depend on Uber

### UBER_PROJECT_ROOT
The absolute path to the project root directory (where the `.uber` file is located).

**Example:**
```bash
UBER_PROJECT_ROOT=/path/to/your/project
```

**Use cases:**
- Reference files relative to the project root
- Build absolute paths to project resources
- Ensure scripts work regardless of current working directory
- Access project configuration files

## Example Usage

```bash
# Run the hello script
uber hello world

# Run the environment demo
uber env-demo

# Run test script with arguments
uber test-script arg1 arg2
```

## Testing the Environment Variables

You can test that the environment variables are working correctly by running:

```bash
# This will show the environment variables in action
uber env-demo

# This will show basic usage
uber hello

# This will show practical examples
uber test-script
```

The scripts will display the actual values of `UBER_BIN_PATH` and `UBER_PROJECT_ROOT` that are passed to them by the Uber tool. 