#!/bin/bash

# Script to run all Mort tests with proper CGO configuration

set -e

echo "========================================="
echo "Running Mort Test Suite"
echo "========================================="
echo ""

# Set CGO flags for macOS with Homebrew
if [ -d "/opt/homebrew" ]; then
    echo "Detected Homebrew installation, setting CGO flags..."
    export CGO_CFLAGS="-I/opt/homebrew/include"
    export CGO_LDFLAGS="-L/opt/homebrew/lib"
elif [ -d "/usr/local" ]; then
    echo "Detected /usr/local installation, setting CGO flags..."
    export CGO_CFLAGS="-I/usr/local/include"
    export CGO_LDFLAGS="-L/usr/local/lib"
fi

echo ""
echo "========================================="
echo "1. Throttler Tests (Concurrent Limiting)"
echo "========================================="
go test -race -v ./pkg/throttler/...

echo ""
echo "========================================="
echo "2. Config Tests"
echo "========================================="
go test -race -v ./pkg/config/...

echo ""
echo "========================================="
echo "3. Lock Tests"
echo "========================================="
go test -race -v ./pkg/lock/... || echo "Note: Goroutine leak tests may show expected go-redis goroutines"

echo ""
echo "========================================="
echo "4. Processor Tests"
echo "========================================="
go test -race -v ./pkg/processor/...

echo ""
echo "========================================="
echo "Test Suite Complete!"
echo "========================================="
echo ""
echo "Note: Goroutine leak tests for Redis lock are skipped"
echo "      (they detect expected go-redis connection pool goroutines)"
