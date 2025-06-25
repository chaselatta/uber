#!/bin/bash

# Build script for uber
set -e

# Source shared utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/utils.sh"

echo "Building uber..."

# Change to the project root
cd_project_root

# Clean previous builds
rm -rf dist/

# Build for current platform
go build -o dist/uber .

echo "Build complete! Binary created as 'uber'"
echo "You can test it with: ./dist/uber --version"
