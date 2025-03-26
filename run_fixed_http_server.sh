#!/bin/bash

# Exit on any error
set -e

# Build the server binary if it doesn't exist
if [ ! -f "./build/server" ]; then
  echo "Building server binary..."
  mkdir -p build
  go build -o build/server ./cmd/server/
fi

# Run the server with fixed HTTP mode for Cursor compatibility
echo "Starting MCP server with fixed HTTP transport on port 8080..."
exec ./build/server -stdio=false -fixed-http=true -http=:8080 