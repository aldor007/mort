#!/usr/bin/env bash

mkdir -p /tmp/mort-tests/local/dir
mkdir -p /tmp/mort-tests/remote/dir

echo "test" > /tmp/mort-tests/local/file
echo "test" > /tmp/mort-tests/remote/file

MORT_PORT=$(( ( RANDOM % 1024 )  + 6012 ))
export MORT_PORT

go run cmd/mort.go -listen ":${MORT_PORT}" -config tests-int/config.yml > mort.logs &
pid=$!
sleep 10

./node_modules/.bin/mocha tests-int/*.Spec.js

echo
kill -9 $pid
rm -rf /tmp/mort-tests
unset MORT_PORT