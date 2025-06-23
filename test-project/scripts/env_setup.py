#!/usr/bin/env python3
import os
import sys
import argparse

def main():
    # Default greeting
    greeting = "Hello from the env_setup.py script!"

    # Use argparse to parse the UBER_GLOBAL_COMMAND_ARGS
    parser = argparse.ArgumentParser()
    parser.add_argument('--name', type=str, help='A name to greet')

    # Get args from environment variable and parse them
    global_args_str = os.environ.get("UBER_GLOBAL_COMMAND_ARGS", "")
    
    # We only care about arguments we define, so we use parse_known_args
    # to ignore other arguments that might be passed to the main uber command
    # or the tool being executed.
    if global_args_str:
        args, unknown = parser.parse_known_args(global_args_str.split())
        if args.name:
            greeting = f"Hello, {args.name}!"

    # The script should output lines in the format `KEY=VALUE`
    print(f'DEMO_GREETING="{greeting}"')

    # This informational message is printed to stderr
    print("env_setup.py: Setting DEMO_GREETING", file=sys.stderr)

if __name__ == "__main__":
    main() 