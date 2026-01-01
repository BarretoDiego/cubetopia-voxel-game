#!/bin/bash

# Voxel Engine - Go Version Run Script
# This script builds and runs the Go voxel engine

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
GO_DIR="$SCRIPT_DIR/go"

echo "==================================="
echo "  Voxel Engine - Go Build & Run"
echo "==================================="
echo

cd "$GO_DIR"

# Download dependencies
echo "📦 Downloading dependencies..."
go mod tidy
go mod download

# Build
echo "🔨 Building..."
go build -o voxelgame ./cmd/voxelgame

# Run
echo "🎮 Running..."
echo
./voxelgame
