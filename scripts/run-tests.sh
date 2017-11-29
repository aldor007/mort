#!/usr/bin/env bash

rm -rf /tmp/mort-tests

mkdir -p /tmp/mort-tests/local/dir/
mkdir -p /tmp/mort-tests/local/dir/a/b/c
mkdir -p /tmp/mort-tests/local/dir2/a/b/c
mkdir -p /tmp/mort-tests/remote/dir

echo "test" > /tmp/mort-tests/local/file
echo "test" > /tmp/mort-tests/remote/file

MORT_PORT=$(( ( RANDOM % 1024 )  + 6012 ))
export MORT_PORT

go run cmd/mort/mort.go -listen ":${MORT_PORT}" -config tests-int/config.yml > mort.logs &
pid=$!
sleep 15

./node_modules/.bin/mocha tests-int/*.Spec.js
TEST_RESULT=$?
echo
kill -9 $pid
if [[ $TEST_RESULT -eq 0 ]]; then
    rm -rf /tmp/mort-tests
fi
unset MORT_PORT