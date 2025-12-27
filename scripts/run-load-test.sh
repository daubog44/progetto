#!/bin/bash
echo "Starting K6 Load Test..."
TEST_FILE=${1:-/scripts/load.js}
echo "Starting K6 Load Test with $TEST_FILE..."
docker compose run --rm k6 run $TEST_FILE
echo "Test Completed. Check Grafana."
