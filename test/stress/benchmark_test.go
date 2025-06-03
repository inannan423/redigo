package main

import (
	"bufio"
	"fmt"
	"net"
	"redigo/config"
	"redigo/resp/handler"
	"redigo/tcp"
	"strconv"
	"sync"
	"testing"
	"time"
)

var (
	benchmarkServerOnce sync.Once
	benchmarkPort       = 6390 // Use different port to avoid conflicts
)

// setupBenchmarkServer starts a test server for benchmarking
func setupBenchmarkServer() {
	benchmarkServerOnce.Do(func() {
		config.Properties = &config.ServerProperties{
			Bind: "localhost",
			Port: benchmarkPort,
		}

		go func() {
			err := tcp.ListenAndServeWithSignal(
				&tcp.Config{
					Address: fmt.Sprintf("%s:%d", config.Properties.Bind, config.Properties.Port),
				},
				handler.MakeHandler())
			if err != nil {
				panic(err)
			}
		}()

		// Wait for server to start
		time.Sleep(500 * time.Millisecond)
	})
}

// getConnection returns a connection to the benchmark server
func getConnection(b *testing.B) net.Conn {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", benchmarkPort))
	if err != nil {
		b.Fatalf("Failed to connect to benchmark server: %v", err)
	}
	return conn
}

// sendBenchmarkCommand sends a command to Redis and reads the response
func sendBenchmarkCommand(conn net.Conn, command string) error {
	_, err := conn.Write([]byte(command))
	if err != nil {
		return err
	}

	// Read response
	reader := bufio.NewReader(conn)
	line, _, err := reader.ReadLine()
	if err != nil {
		return err
	}

	// Handle bulk string responses
	if len(line) > 0 && line[0] == '$' {
		size, err := strconv.Atoi(string(line[1:]))
		if err != nil {
			return err
		}
		if size > 0 {
			data := make([]byte, size+2) // +2 for \r\n
			_, err = reader.Read(data)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// BenchmarkSET tests SET command performance
func BenchmarkSET(b *testing.B) {
	setupBenchmarkServer()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		conn := getConnection(b)
		defer conn.Close()

		i := 0
		for pb.Next() {
			key := fmt.Sprintf("benchmark:set:%d", i)
			value := fmt.Sprintf("value_%d", i)
			command := fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
				len(key), key, len(value), value)

			err := sendBenchmarkCommand(conn, command)
			if err != nil {
				b.Fatalf("SET command failed: %v", err)
			}
			i++
		}
	})
}

// BenchmarkGET tests GET command performance
func BenchmarkGET(b *testing.B) {
	setupBenchmarkServer()

	// Prepare some data first using a separate connection
	prepConn := getConnection(b)
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("benchmark:get:%d", i)
		value := fmt.Sprintf("value_%d", i)
		command := fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
			len(key), key, len(value), value)
		sendBenchmarkCommand(prepConn, command)
	}
	prepConn.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		conn := getConnection(b)
		defer conn.Close()

		i := 0
		for pb.Next() {
			key := fmt.Sprintf("benchmark:get:%d", i%1000)
			command := fmt.Sprintf("*2\r\n$3\r\nGET\r\n$%d\r\n%s\r\n", len(key), key)

			err := sendBenchmarkCommand(conn, command)
			if err != nil {
				b.Fatalf("GET command failed: %v", err)
			}
			i++
		}
	})
}

// BenchmarkPING tests PING command performance
func BenchmarkPING(b *testing.B) {
	setupBenchmarkServer()

	command := "*1\r\n$4\r\nPING\r\n"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		conn := getConnection(b)
		defer conn.Close()

		for pb.Next() {
			err := sendBenchmarkCommand(conn, command)
			if err != nil {
				b.Fatalf("PING command failed: %v", err)
			}
		}
	})
}

// BenchmarkHSET tests HSET command performance
func BenchmarkHSET(b *testing.B) {
	setupBenchmarkServer()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		conn := getConnection(b)
		defer conn.Close()

		i := 0
		for pb.Next() {
			key := fmt.Sprintf("benchmark:hash:%d", i)
			field := "field1"
			value := fmt.Sprintf("value_%d", i)
			command := fmt.Sprintf("*4\r\n$4\r\nHSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
				len(key), key, len(field), field, len(value), value)

			err := sendBenchmarkCommand(conn, command)
			if err != nil {
				b.Fatalf("HSET command failed: %v", err)
			}
			i++
		}
	})
}

// BenchmarkHGET tests HGET command performance
func BenchmarkHGET(b *testing.B) {
	setupBenchmarkServer()

	// Prepare some hash data first using a separate connection
	prepConn := getConnection(b)
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("benchmark:hash:get:%d", i)
		field := "field1"
		value := fmt.Sprintf("value_%d", i)
		command := fmt.Sprintf("*4\r\n$4\r\nHSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
			len(key), key, len(field), field, len(value), value)
		sendBenchmarkCommand(prepConn, command)
	}
	prepConn.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		conn := getConnection(b)
		defer conn.Close()

		i := 0
		for pb.Next() {
			key := fmt.Sprintf("benchmark:hash:get:%d", i%1000)
			field := "field1"
			command := fmt.Sprintf("*3\r\n$4\r\nHGET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
				len(key), key, len(field), field)

			err := sendBenchmarkCommand(conn, command)
			if err != nil {
				b.Fatalf("HGET command failed: %v", err)
			}
			i++
		}
	})
}

// BenchmarkLPUSH tests LPUSH command performance
func BenchmarkLPUSH(b *testing.B) {
	setupBenchmarkServer()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		conn := getConnection(b)
		defer conn.Close()

		i := 0
		for pb.Next() {
			key := fmt.Sprintf("benchmark:list:%d", i%100)
			value := fmt.Sprintf("value_%d", i)
			command := fmt.Sprintf("*3\r\n$5\r\nLPUSH\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
				len(key), key, len(value), value)

			err := sendBenchmarkCommand(conn, command)
			if err != nil {
				b.Fatalf("LPUSH command failed: %v", err)
			}
			i++
		}
	})
}

// BenchmarkSADD tests SADD command performance
func BenchmarkSADD(b *testing.B) {
	setupBenchmarkServer()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		conn := getConnection(b)
		defer conn.Close()

		i := 0
		for pb.Next() {
			key := fmt.Sprintf("benchmark:set:%d", i%100) // Use limited keys to build sets
			member := fmt.Sprintf("member_%d", i)
			command := fmt.Sprintf("*3\r\n$4\r\nSADD\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
				len(key), key, len(member), member)

			err := sendBenchmarkCommand(conn, command)
			if err != nil {
				b.Fatalf("SADD command failed: %v", err)
			}
			i++
		}
	})
}

// BenchmarkMixed tests mixed operations
func BenchmarkMixed(b *testing.B) {
	setupBenchmarkServer()

	// Prepare some data using a separate connection
	prepConn := getConnection(b)
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("benchmark:mixed:%d", i)
		value := fmt.Sprintf("value_%d", i)
		setCmd := fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
			len(key), key, len(value), value)
		sendBenchmarkCommand(prepConn, setCmd)
	}
	prepConn.Close()

	commands := []string{
		"SET", "GET", "HSET", "HGET", "PING",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		conn := getConnection(b)
		defer conn.Close()

		i := 0
		for pb.Next() {
			cmdType := commands[i%len(commands)]
			var command string

			switch cmdType {
			case "SET":
				key := fmt.Sprintf("benchmark:mixed:%d", i%100)
				value := fmt.Sprintf("new_value_%d", i)
				command = fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
					len(key), key, len(value), value)
			case "GET":
				key := fmt.Sprintf("benchmark:mixed:%d", i%100)
				command = fmt.Sprintf("*2\r\n$3\r\nGET\r\n$%d\r\n%s\r\n", len(key), key)
			case "HSET":
				key := fmt.Sprintf("benchmark:mixed:hash:%d", i%100)
				field := "field1"
				value := fmt.Sprintf("hash_value_%d", i)
				command = fmt.Sprintf("*4\r\n$4\r\nHSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
					len(key), key, len(field), field, len(value), value)
			case "HGET":
				key := fmt.Sprintf("benchmark:mixed:hash:%d", i%100)
				field := "field1"
				command = fmt.Sprintf("*3\r\n$4\r\nHGET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
					len(key), key, len(field), field)
			case "PING":
				command = "*1\r\n$4\r\nPING\r\n"
			}

			err := sendBenchmarkCommand(conn, command)
			if err != nil {
				b.Fatalf("%s command failed: %v", cmdType, err)
			}
			i++
		}
	})
}

// BenchmarkConcurrentConnections tests performance with multiple connections
func BenchmarkConcurrentConnections(b *testing.B) {
	setupBenchmarkServer()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		conn := getConnection(b)
		defer conn.Close()

		i := 0
		for pb.Next() {
			key := fmt.Sprintf("benchmark:concurrent:%d", i)
			value := fmt.Sprintf("value_%d", i)
			command := fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
				len(key), key, len(value), value)

			err := sendBenchmarkCommand(conn, command)
			if err != nil {
				b.Fatalf("SET command failed: %v", err)
			}
			i++
		}
	})
}

// BenchmarkZADD tests ZADD command performance
func BenchmarkZADD(b *testing.B) {
	setupBenchmarkServer()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		conn := getConnection(b)
		defer conn.Close()

		i := 0
		for pb.Next() {
			key := fmt.Sprintf("benchmark:zset:%d", i%100)
			member := fmt.Sprintf("member_%d", i)
			score := fmt.Sprintf("%.2f", float64(i%100))
			command := fmt.Sprintf("*4\r\n$4\r\nZADD\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
				len(key), key, len(score), score, len(member), member)

			err := sendBenchmarkCommand(conn, command)
			if err != nil {
				b.Fatalf("ZADD command failed: %v", err)
			}
			i++
		}
	})
}

// Example usage function for documentation
func ExampleBenchmark() {
	// Run all benchmarks:
	// go test -bench=. -benchmem
	//
	// Run specific benchmark:
	// go test -bench=BenchmarkSET -benchmem
	//
	// Run with specific duration:
	// go test -bench=. -benchtime=10s
	//
	// Run with specific CPU count:
	// go test -bench=. -cpu=1,2,4
	fmt.Println("Use 'go test -bench=.' to run benchmarks")
}
