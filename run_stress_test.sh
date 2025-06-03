#!/bin/bash

# Convenience script to run stress tests from project root
# This script forwards all arguments to the stress test tool

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
STRESS_DIR="$SCRIPT_DIR/test/stress"

# Check if stress test directory exists
if [ ! -d "$STRESS_DIR" ]; then
    echo "Error: Stress test directory not found at $STRESS_DIR"
    exit 1
fi

# Change to stress test directory and run the script
cd "$STRESS_DIR"

# Forward all arguments to the stress test script
./stress_test.sh "$@" 