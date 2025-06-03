#!/bin/bash

# Redis Stress Test Script
# This script provides convenient stress testing for the Redigo server

set -e

# Default configuration
HOST="localhost"
PORT=6380
CONNECTIONS=50
REQUESTS=1000
COMMAND="SET"
DATA_SIZE=64

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}Redis Stress Test Script${NC}"
echo -e "${GREEN}================================${NC}"

# Function to print usage
usage() {
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  -h, --host HOST       Redis server host (default: localhost)"
    echo "  -p, --port PORT       Redis server port (default: 6379)"
    echo "  -c, --connections N   Number of concurrent connections (default: 50)"
    echo "  -n, --requests N      Number of requests per connection (default: 1000)"
    echo "  -t, --time DURATION   Test duration (e.g., 30s, 1m)"
    echo "  -cmd, --command CMD   Command to test (default: SET)"
    echo "  -d, --data-size N     Size of test data in bytes (default: 64)"
    echo "  --progress            Show progress during test"
    echo "  --help                Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 -c 100 -n 2000 -cmd SET"
    echo "  $0 -h localhost -p 6379 -t 30s -cmd GET --progress"
    echo "  $0 -c 20 -cmd HSET -d 128"
    echo "  $0 -c 40 -cmd SADD -d 64"
    echo "  $0 -c 30 -cmd SMEMBERS --scenarios"
}

# Function to run predefined test scenarios
run_scenarios() {
    echo -e "\n${YELLOW}Running predefined test scenarios...${NC}\n"
    
    # Scenario 1: Basic SET operations
    echo -e "${GREEN}Scenario 1: Basic SET operations (50 connections, 1000 requests each)${NC}"
    go run main.go -h $HOST -p $PORT -c 50 -n 1000 -cmd SET -d 64
    
    echo -e "\n${GREEN}Scenario 2: GET operations (30 connections, 1500 requests each)${NC}"
    go run main.go -h $HOST -p $PORT -c 30 -n 1500 -cmd GET -d 64
    
    echo -e "\n${GREEN}Scenario 3: PING test (100 connections, 500 requests each)${NC}"
    go run main.go -h $HOST -p $PORT -c 100 -n 500 -cmd PING
    
    echo -e "\n${GREEN}Scenario 4: Hash operations (20 connections, 800 requests each)${NC}"
    go run main.go -h $HOST -p $PORT -c 20 -n 800 -cmd HSET -d 128
    
    echo -e "\n${GREEN}Scenario 5: List operations (25 connections, 600 requests each)${NC}"
    go run main.go -h $HOST -p $PORT -c 25 -n 600 -cmd LPUSH -d 32
    
    echo -e "\n${GREEN}Scenario 6: Set ADD operations (40 connections, 800 requests each)${NC}"
    go run main.go -h $HOST -p $PORT -c 40 -n 800 -cmd SADD -d 64
    
    echo -e "\n${GREEN}Scenario 7: Set MEMBERS operations (30 connections, 600 requests each)${NC}"
    go run main.go -h $HOST -p $PORT -c 30 -n 600 -cmd SMEMBERS -d 64
    
    echo -e "\n${GREEN}Scenario 8: Mixed Set operations (20 connections, 400 SADD + 400 SMEMBERS each)${NC}"
    echo -e "${YELLOW}  Running SADD operations...${NC}"
    go run main.go -h $HOST -p $PORT -c 20 -n 400 -cmd SADD -d 64
    echo -e "${YELLOW}  Running SMEMBERS operations...${NC}"
    go run main.go -h $HOST -p $PORT -c 20 -n 400 -cmd SMEMBERS -d 64
    
    echo -e "\n${YELLOW}All scenarios completed!${NC}"
}

# Function to check if server is running
check_server() {
    if nc -z $HOST $PORT 2>/dev/null; then
        echo -e "${GREEN}✓ Redis server is running on $HOST:$PORT${NC}"
        return 0
    else
        echo -e "${RED}✗ Cannot connect to Redis server at $HOST:$PORT${NC}"
        echo -e "${YELLOW}Please make sure the Redis server is running first.${NC}"
        echo -e "${YELLOW}You can start it with: go run main.go${NC}"
        return 1
    fi
}

# Parse command line arguments
SCENARIOS=false
PROGRESS=""
DURATION=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--host)
            HOST="$2"
            shift 2
            ;;
        -p|--port)
            PORT="$2"
            shift 2
            ;;
        -c|--connections)
            CONNECTIONS="$2"
            shift 2
            ;;
        -n|--requests)
            REQUESTS="$2"
            shift 2
            ;;
        -t|--time)
            DURATION="-t $2"
            shift 2
            ;;
        -cmd|--command)
            COMMAND="$2"
            shift 2
            ;;
        -d|--data-size)
            DATA_SIZE="$2"
            shift 2
            ;;
        --progress)
            PROGRESS="-progress"
            shift
            ;;
        --scenarios)
            SCENARIOS=true
            shift
            ;;
        --help)
            usage
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            usage
            exit 1
            ;;
    esac
done

# Check if server is running
if ! check_server; then
    exit 1
fi

# Run scenarios or single test
if $SCENARIOS; then
    run_scenarios
else
    echo -e "\n${YELLOW}Running stress test with following configuration:${NC}"
    echo "  Host: $HOST"
    echo "  Port: $PORT"
    echo "  Connections: $CONNECTIONS"
    if [[ -n "$DURATION" ]]; then
        echo "  Duration: ${DURATION#-t }"
    else
        echo "  Requests per connection: $REQUESTS"
    fi
    echo "  Command: $COMMAND"
    echo "  Data size: $DATA_SIZE bytes"
    echo ""
    
    # Build and run the stress test command
    if [[ -n "$DURATION" ]]; then
        go run main.go -h $HOST -p $PORT -c $CONNECTIONS $DURATION -cmd $COMMAND -d $DATA_SIZE $PROGRESS
    else
        go run main.go -h $HOST -p $PORT -c $CONNECTIONS -n $REQUESTS -cmd $COMMAND -d $DATA_SIZE $PROGRESS
    fi
fi 