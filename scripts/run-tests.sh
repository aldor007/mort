#!/usr/bin/env bash

rm -rf /tmp/mort-tests

mkdir -p /tmp/mort-tests/local/dir/
mkdir -p /tmp/mort-tests/local/dir/a/b/c
mkdir -p /tmp/mort-tests/local/dir2/a/b/c
mkdir -p /tmp/mort-tests/remote/dir

echo "test" > /tmp/mort-tests/local/file
echo "test" > /tmp/mort-tests/remote/file

cp -r pkg/processor/benchmark/local/* /tmp/mort-tests/local/

MORT_PORT=8091
export MORT_PORT

go run cmd/mort/mort.go -config tests-int/config.yml  &
pid=$!
sleep 15

./node_modules/.bin/mocha tests-int/*.Spec.js
TEST_RESULT=$?
echo
kill -9  $pid
if [[ $TEST_RESULT -eq 0 ]]; then
    rm -rf /tmp/mort-tests
fi
unset MORT_PORT
exit ${TEST_RESULT}