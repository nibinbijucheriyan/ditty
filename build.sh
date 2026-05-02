#!/bin/bash

# Build script for Intelligent Observability Agent

set -e

echo "Building Intelligent Observability Agent..."

# Create bin directory
mkdir -p bin

# Build Risk Monitor
echo "Building Risk Monitor..."
cd monitor
go build -o ../bin/risk-monitor monitor.go
cd ..

# Build MCP Server
echo "Building MCP Server..."
cd mcp-server
go build -o ../bin/mcp-server mcp_server.go
cd ..

echo "Build complete! Binaries are in the bin/ directory:"
ls -lh bin/

echo ""
echo "To run the components:"
echo "  ./bin/mcp-server"
echo "  ./bin/risk-monitor"

# Made with Bob
