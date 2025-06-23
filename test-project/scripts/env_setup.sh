#!/bin/bash
# This script is sourced by uber before running a tool.
# You can use this to set up any environment variables your tools might need.

# Default greeting
GREETING="Hello from the env_setup.sh script!"

# Check for --name flag in the UBER_GLOBAL_COMMAND_ARGS environment variable
if [ -n "$UBER_GLOBAL_COMMAND_ARGS" ]; then
  # Read the string of arguments into a bash array
  read -ra global_args <<<"$UBER_GLOBAL_COMMAND_ARGS"

  # Loop through the array to find the --name flag
  for i in "${!global_args[@]}"; do
    if [[ "${global_args[$i]}" == "--name" && $((i + 1)) -lt "${#global_args[@]}" ]]; then
      GREETING="Hello, ${global_args[$((i + 1))]}!"
      break
    fi
  done
fi

export DEMO_GREETING="$GREETING"

# This informational message is printed to stderr so it doesn't pollute the `env` output
echo "env_setup.sh: Setting DEMO_GREETING" >&2
