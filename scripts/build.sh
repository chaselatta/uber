#!/bin/bash

# Build script for uber
set -e

echo "Building uber..."

# Clean previous builds
rm -rf dist/

# Build for current platform
go build -o uber ./cmd/uber

echo "Build complete! Binary created as 'uber'"
echo "You can test it with: ./uber --version"
