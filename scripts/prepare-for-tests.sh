#!/usr/bin/env bash

mkdir -p /tmp/mort-tests/local/dir/
mkdir -p /tmp/mort-tests/local/dir/a/b/c
mkdir -p /tmp/mort-tests/local/dir2/a/b/c
mkdir -p /tmp/mort-tests/remote/dir

echo "test" > /tmp/mort-tests/local/file
echo "test" > /tmp/mort-tests/remote/file

# Create 1GB file - cross-platform approach
if command -v fallocate >/dev/null 2>&1; then
    # Linux
    fallocate -l 1G /tmp/mort-tests/local/big.img
elif command -v mkfile >/dev/null 2>&1; then
    # macOS
    mkfile 1g /tmp/mort-tests/local/big.img
else
    # Fallback using dd (slower but universal)
    dd if=/dev/zero of=/tmp/mort-tests/local/big.img bs=1m count=1024 2>/dev/null
fi

cp -r pkg/processor/benchmark/local/* /tmp/mort-tests/local/