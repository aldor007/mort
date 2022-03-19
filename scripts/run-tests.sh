#!/usr/bin/env bash

bash scripts/prepare-for-tests.sh

pkill -f 'tests-int/config.yml'

MORT_PORT=8091
export MORT_PORT
MORT_HOST=localhost
export MORT_HOST

go run cmd/mort/mort.go -config tests-int/mort.yml  &
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
pkill -f 'tests-int/config.yml'
exit ${TEST_RESULT}
