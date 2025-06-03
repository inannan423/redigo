package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// StressTestConfig holds configuration for stress testing
type StressTestConfig struct {
	Host         string        // Redis server host
	Port         int           // Redis server port
	Connections  int           // Number of concurrent connections
	Requests     int           // Total number of requests per connection
	Duration     time.Duration // Test duration (0 means use request count)
	KeyPrefix    string        // Prefix for test keys
	Command      string        // Command to test (SET, GET, HSET, etc.)
	DataSize     int           // Size of test data in bytes
	Pipeline     int           // Pipeline size (0 means no pipeline)
	ShowProgress bool          // Show progress during test
}

// TestResult holds the results of a stress test
type TestResult struct {
	TotalRequests   int64         // Total requests sent
	SuccessRequests int64         // Successful requests
	FailedRequests  int64         // Failed requests
	Duration        time.Duration // Total test duration
	MinLatency      time.Duration // Minimum latency
	MaxLatency      time.Duration // Maximum latency
	AvgLatency      time.Duration // Average latency
	P50Latency      time.Duration // 50th percentile latency
	P95Latency      time.Duration // 95th percentile latency
	P99Latency      time.Duration // 99th percentile latency
	QPS             float64       // Queries per second
}

// StressTester performs stress testing on Redis server
type StressTester struct {
	config *StressTestConfig
	mutex  sync.Mutex
}

// NewStressTester creates a new stress tester
func NewStressTester(config *StressTestConfig) *StressTester {
	return &StressTester{
		config: config,
	}
}

// generateTestData generates random test data of specified size
func generateTestData(size int) string {
	if size <= 0 {
		return "testvalue"
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	data := make([]byte, size)
	for i := range data {
		data[i] = charset[rand.Intn(len(charset))]
	}
	return string(data)
}

// buildCommand builds Redis command based on test configuration
func (st *StressTester) buildCommand(keyIndex int) string {
	key := fmt.Sprintf("%s:%d", st.config.KeyPrefix, keyIndex)
	value := generateTestData(st.config.DataSize)

	switch strings.ToUpper(st.config.Command) {
	case "SET":
		return fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
			len(key), key, len(value), value)
	case "GET":
		return fmt.Sprintf("*2\r\n$3\r\nGET\r\n$%d\r\n%s\r\n", len(key), key)
	case "HSET":
		field := "field1"
		return fmt.Sprintf("*4\r\n$4\r\nHSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
			len(key), key, len(field), field, len(value), value)
	case "HGET":
		field := "field1"
		return fmt.Sprintf("*3\r\n$4\r\nHGET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
			len(key), key, len(field), field)
	case "PING":
		return "*1\r\n$4\r\nPING\r\n"
	case "LPUSH":
		return fmt.Sprintf("*3\r\n$5\r\nLPUSH\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
			len(key), key, len(value), value)
	case "RPUSH":
		return fmt.Sprintf("*3\r\n$5\r\nRPUSH\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
			len(key), key, len(value), value)
	case "LPOP":
		return fmt.Sprintf("*2\r\n$4\r\nLPOP\r\n$%d\r\n%s\r\n", len(key), key)
	case "SADD":
		return fmt.Sprintf("*3\r\n$4\r\nSADD\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
			len(key), key, len(value), value)
	case "SMEMBERS":
		return fmt.Sprintf("*2\r\n$8\r\nSMEMBERS\r\n$%d\r\n%s\r\n", len(key), key)
	default:
		// Default to SET command
		return fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
			len(key), key, len(value), value)
	}
}

// readResponse reads response from Redis server
func readResponse(conn net.Conn) (string, error) {
	reader := bufio.NewReader(conn)

	// Read first line to determine response type
	line, _, err := reader.ReadLine()
	if err != nil {
		return "", err
	}

	response := string(line) + "\r\n"

	// Handle different response types
	switch line[0] {
	case '+', '-', ':':
		// Simple string, error, or integer - already complete
		return response, nil
	case '$':
		// Bulk string
		size, err := strconv.Atoi(string(line[1:]))
		if err != nil || size < 0 {
			return response, nil
		}

		// Read the actual string data
		data := make([]byte, size+2) // +2 for \r\n
		_, err = reader.Read(data)
		if err != nil {
			return response, err
		}
		response += string(data)
	case '*':
		// Array - simplified handling, just read next few lines
		count, err := strconv.Atoi(string(line[1:]))
		if err != nil || count <= 0 {
			return response, nil
		}

		// For simplicity, read up to 10 more lines
		for i := 0; i < count && i < 10; i++ {
			line, _, err := reader.ReadLine()
			if err != nil {
				break
			}
			response += string(line) + "\r\n"
		}
	}

	return response, nil
}

// runWorker runs stress test for a single connection
func (st *StressTester) runWorker(ctx context.Context, workerId int, results chan<- TestResult) {
	var totalRequests, successRequests, failedRequests int64
	var latencies []time.Duration

	// Connect to Redis server
	address := fmt.Sprintf("%s:%d", st.config.Host, st.config.Port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Printf("Worker %d: Failed to connect to %s: %v\n", workerId, address, err)
		results <- TestResult{FailedRequests: 1}
		return
	}
	defer conn.Close()

	// Set connection timeout
	conn.SetDeadline(time.Now().Add(time.Minute))

	startTime := time.Now()
	requestCount := 0

	for {
		select {
		case <-ctx.Done():
			// Test duration exceeded or context cancelled
			goto finish
		default:
			// Continue with requests
		}

		// Check if we've reached the request limit
		if st.config.Duration == 0 && requestCount >= st.config.Requests {
			break
		}

		// Send command
		keyIndex := rand.Intn(1000) // Use random key index for more realistic testing
		command := st.buildCommand(keyIndex)

		requestStart := time.Now()
		_, err := conn.Write([]byte(command))
		if err != nil {
			failedRequests++
			continue
		}

		// Read response
		_, err = readResponse(conn)
		requestEnd := time.Now()

		if err != nil {
			failedRequests++
		} else {
			successRequests++
			latency := requestEnd.Sub(requestStart)
			latencies = append(latencies, latency)
		}

		totalRequests++
		requestCount++

		// Reset connection deadline
		conn.SetDeadline(time.Now().Add(time.Minute))
	}

finish:
	duration := time.Since(startTime)

	// Calculate statistics
	result := TestResult{
		TotalRequests:   totalRequests,
		SuccessRequests: successRequests,
		FailedRequests:  failedRequests,
		Duration:        duration,
	}

	// Calculate latency statistics
	if len(latencies) > 0 {
		sort.Slice(latencies, func(i, j int) bool {
			return latencies[i] < latencies[j]
		})

		result.MinLatency = latencies[0]
		result.MaxLatency = latencies[len(latencies)-1]

		var totalLatency time.Duration
		for _, lat := range latencies {
			totalLatency += lat
		}
		result.AvgLatency = totalLatency / time.Duration(len(latencies))

		if len(latencies) > 0 {
			result.P50Latency = latencies[len(latencies)*50/100]
		}
		if len(latencies) > 0 {
			result.P95Latency = latencies[len(latencies)*95/100]
		}
		if len(latencies) > 0 {
			result.P99Latency = latencies[len(latencies)*99/100]
		}
	}

	if duration > 0 {
		result.QPS = float64(successRequests) / duration.Seconds()
	}

	results <- result
}

// Run executes the stress test
func (st *StressTester) Run() *TestResult {
	fmt.Printf("Starting stress test...\n")
	fmt.Printf("Target: %s:%d\n", st.config.Host, st.config.Port)
	fmt.Printf("Connections: %d\n", st.config.Connections)
	fmt.Printf("Command: %s\n", st.config.Command)
	fmt.Printf("Data size: %d bytes\n", st.config.DataSize)

	if st.config.Duration > 0 {
		fmt.Printf("Duration: %v\n", st.config.Duration)
	} else {
		fmt.Printf("Requests per connection: %d\n", st.config.Requests)
	}
	fmt.Println()

	// Create context for managing test duration
	ctx := context.Background()
	if st.config.Duration > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, st.config.Duration)
		defer cancel()
	}

	// Channel to collect results from workers
	results := make(chan TestResult, st.config.Connections)

	// Start workers
	startTime := time.Now()
	for i := 0; i < st.config.Connections; i++ {
		go st.runWorker(ctx, i, results)
	}

	// Progress reporting
	if st.config.ShowProgress {
		go func() {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					elapsed := time.Since(startTime)
					fmt.Printf("Progress: %.1fs elapsed\n", elapsed.Seconds())
				}
			}
		}()
	}

	// Collect results from all workers
	var totalResult TestResult
	var allLatencies []time.Duration

	for i := 0; i < st.config.Connections; i++ {
		result := <-results
		totalResult.TotalRequests += result.TotalRequests
		totalResult.SuccessRequests += result.SuccessRequests
		totalResult.FailedRequests += result.FailedRequests

		if result.Duration > totalResult.Duration {
			totalResult.Duration = result.Duration
		}

		// Update latency statistics
		if totalResult.MinLatency == 0 || (result.MinLatency > 0 && result.MinLatency < totalResult.MinLatency) {
			totalResult.MinLatency = result.MinLatency
		}
		if result.MaxLatency > totalResult.MaxLatency {
			totalResult.MaxLatency = result.MaxLatency
		}
	}

	// Calculate aggregated latency statistics
	if len(allLatencies) > 0 {
		sort.Slice(allLatencies, func(i, j int) bool {
			return allLatencies[i] < allLatencies[j]
		})

		var totalLatency time.Duration
		for _, lat := range allLatencies {
			totalLatency += lat
		}
		totalResult.AvgLatency = totalLatency / time.Duration(len(allLatencies))
		totalResult.P50Latency = allLatencies[len(allLatencies)*50/100]
		totalResult.P95Latency = allLatencies[len(allLatencies)*95/100]
		totalResult.P99Latency = allLatencies[len(allLatencies)*99/100]
	}

	// Calculate final QPS
	if totalResult.Duration > 0 {
		totalResult.QPS = float64(totalResult.SuccessRequests) / totalResult.Duration.Seconds()
	}

	return &totalResult
}

// PrintResults prints formatted test results
func PrintResults(result *TestResult) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("STRESS TEST RESULTS")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Printf("Total Requests:     %d\n", result.TotalRequests)
	fmt.Printf("Successful:         %d\n", result.SuccessRequests)
	fmt.Printf("Failed:             %d\n", result.FailedRequests)

	if result.TotalRequests > 0 {
		fmt.Printf("Success Rate:       %.2f%%\n",
			float64(result.SuccessRequests)/float64(result.TotalRequests)*100)
	}

	fmt.Printf("Test Duration:      %v\n", result.Duration)
	fmt.Printf("Queries Per Second: %.2f\n", result.QPS)

	if result.SuccessRequests > 0 && result.MinLatency > 0 {
		fmt.Println("\nLatency Statistics:")
		fmt.Printf("  Min:              %v\n", result.MinLatency)
		fmt.Printf("  Max:              %v\n", result.MaxLatency)
		if result.AvgLatency > 0 {
			fmt.Printf("  Average:          %v\n", result.AvgLatency)
		}
		if result.P50Latency > 0 {
			fmt.Printf("  50th percentile:  %v\n", result.P50Latency)
		}
		if result.P95Latency > 0 {
			fmt.Printf("  95th percentile:  %v\n", result.P95Latency)
		}
		if result.P99Latency > 0 {
			fmt.Printf("  99th percentile:  %v\n", result.P99Latency)
		}
	}

	fmt.Println(strings.Repeat("=", 60))
}

func main() {
	// Parse command line arguments
	var config StressTestConfig

	flag.StringVar(&config.Host, "h", "localhost", "Redis server host")
	flag.IntVar(&config.Port, "p", 6380, "Redis server port")
	flag.IntVar(&config.Connections, "c", 10, "Number of concurrent connections")
	flag.IntVar(&config.Requests, "n", 100, "Number of requests per connection")
	flag.DurationVar(&config.Duration, "t", 0, "Test duration (e.g., 30s, 1m)")
	flag.StringVar(&config.KeyPrefix, "k", "test", "Key prefix for test keys")
	flag.StringVar(&config.Command, "cmd", "SET", "Command to test (SET, GET, HSET, HGET, PING, LPUSH, RPUSH, LPOP, SADD, SMEMBERS)")
	flag.IntVar(&config.DataSize, "d", 64, "Size of test data in bytes")
	flag.IntVar(&config.Pipeline, "P", 0, "Pipeline size (not implemented yet)")
	flag.BoolVar(&config.ShowProgress, "progress", false, "Show progress during test")

	help := flag.Bool("help", false, "Show help message")

	flag.Parse()

	if *help {
		fmt.Println("Redis Stress Test Tool")
		fmt.Println("Usage: go run main.go [options]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  go run main.go -c 50 -n 1000 -cmd SET")
		fmt.Println("  go run main.go -h localhost -p 6380 -t 30s -cmd GET")
		fmt.Println("  go run main.go -c 100 -cmd PING -progress")
		fmt.Println("  go run main.go -c 20 -n 500 -cmd HSET -d 128")
		fmt.Println("  go run main.go -c 40 -n 800 -cmd SADD -d 64")
		fmt.Println("  go run main.go -c 30 -n 600 -cmd SMEMBERS")
		return
	}

	// Validate configuration
	if config.Connections <= 0 {
		fmt.Println("Error: Number of connections must be positive")
		os.Exit(1)
	}

	if config.Duration == 0 && config.Requests <= 0 {
		fmt.Println("Error: Either duration or request count must be specified")
		os.Exit(1)
	}

	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Create and run stress tester
	tester := NewStressTester(&config)
	result := tester.Run()

	// Print results
	PrintResults(result)
}
