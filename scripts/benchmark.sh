#!/usr/bin/env bash

set -e
echo "" > bench.txt

for d in $(go list ./... | grep -v vendor); do
    go test -bench 'Benchmark' "$d" >> bench.txt
done