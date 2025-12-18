#!/usr/bin/env bash

set -e

# Set CGO flags for macOS with Homebrew (needed for brotli)
if [ -d "/opt/homebrew" ]; then
    export CGO_CFLAGS="-I/opt/homebrew/include"
    export CGO_LDFLAGS="-L/opt/homebrew/lib"
elif [ -d "/usr/local" ]; then
    export CGO_CFLAGS="-I/usr/local/include"
    export CGO_LDFLAGS="-L/usr/local/lib"
fi

echo "" > coverage.txt

for d in $(go list ./... | grep -v vendor); do
    go test -bench=. -race -coverprofile=profile.out -covermode=atomic "$d"
    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done