#!/bin/bash

# Exit on any error
set -e

# Build the server binary if it doesn't exist
if [ ! -f "./build/server" ]; then
  echo "Building server binary..."
  mkdir -p build
  go build -o build/server ./cmd/server/
fi

# Run the server with stdio mode
echo "Starting MCP server in stdio mode..."
exec ./build/server -stdio 