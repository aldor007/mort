#!/usr/bin/env bash

set -e  # Exit on error

echo "Preparing test environment..."
bash scripts/prepare-for-tests.sh

echo "Stopping any existing mort processes..."
pkill -f 'tests-int/mort.yml' || true
pkill -f 'mort.*8091' || true
sleep 2

MORT_PORT=8091
export MORT_PORT
MORT_HOST=localhost
export MORT_HOST

echo "Starting mort server on ${MORT_HOST}:${MORT_PORT}..."
# Start mort with output redirected for debugging
go run cmd/mort/mort.go -config tests-int/mort.yml > /tmp/mort-test.log 2>&1 &
pid=$!

echo "Mort PID: $pid"

# Wait for mort to be ready (max 30 seconds)
echo "Waiting for mort to be ready..."
for i in {1..30}; do
    if curl -s "http://${MORT_HOST}:${MORT_PORT}" > /dev/null 2>&1; then
        echo "Mort is ready!"
        break
    fi
    if ! kill -0 $pid 2>/dev/null; then
        echo "ERROR: Mort process died during startup!"
        echo "Last 50 lines of mort log:"
        tail -50 /tmp/mort-test.log
        exit 1
    fi
    echo "Waiting for mort to start... ($i/30)"
    sleep 1
done

# Check if mort is still running
if ! kill -0 $pid 2>/dev/null; then
    echo "ERROR: Mort is not running!"
    echo "Last 50 lines of mort log:"
    tail -50 /tmp/mort-test.log
    exit 1
fi

echo "Running integration tests..."
./node_modules/.bin/mocha --file ./tests-int/setup-mort.js tests-int/*.Spec.js
TEST_RESULT=$?

echo ""
echo "Cleaning up..."

# Check if mort crashed during tests
if ! kill -0 $pid 2>/dev/null; then
    echo "WARNING: Mort crashed during tests!"
    echo "Last 100 lines of mort log:"
    tail -100 /tmp/mort-test.log
fi

# Stop mort gracefully first, then forcefully if needed
kill $pid 2>/dev/null || true
sleep 2
kill -9 $pid 2>/dev/null || true

if [[ $TEST_RESULT -eq 0 ]]; then
    echo "All tests passed! Cleaning up test files..."
    rm -rf /tmp/mort-tests
else
    echo "Tests failed! Keeping test files for debugging."
    echo "Mort log available at: /tmp/mort-test.log"
fi

unset MORT_PORT
unset MORT_HOST
pkill -f 'tests-int/mort.yml' || true

exit ${TEST_RESULT}
