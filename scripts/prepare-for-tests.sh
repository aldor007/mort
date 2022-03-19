#!/usr/bin/env bash

mkdir -p /tmp/mort-tests/local/dir/
mkdir -p /tmp/mort-tests/local/dir/a/b/c
mkdir -p /tmp/mort-tests/local/dir2/a/b/c
mkdir -p /tmp/mort-tests/remote/dir

echo "test" > /tmp/mort-tests/local/file
echo "test" > /tmp/mort-tests/remote/file
fallocate -l 1G /tmp/mort-tests/local/big.img

cp -r pkg/processor/benchmark/local/* /tmp/mort-tests/local/