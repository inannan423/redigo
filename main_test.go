package main

import (
	"fmt"
	"net"
	"os"
	"redigo/config"
	"redigo/resp/handler"
	"redigo/tcp"
	"testing"
	"time"
)

func TestRedisHashCommands(t *testing.T) {
	// Start Redis server (make sure the server is properly configured)

	// Wait for the server to start
	time.Sleep(1 * time.Second)

	// Connect to Redis server
	conn, err := net.Dial("tcp", "localhost:6380")
	if err != nil {
		t.Fatalf("Failed to connect to Redis server: %v", err)
	}
	defer conn.Close()

	// Test HSET command
	t.Run("HSET", func(t *testing.T) {
		// Send HSET command
		sendCommand(conn, "*4\r\n$4\r\nHSET\r\n$4\r\nuser\r\n$4\r\nname\r\n$4\r\nJohn\r\n")

		// Read reply
		reply := readReply(conn)

		// Expected reply is ":1\r\n" (integer reply 1, means a new field was set)
		if reply != ":1\r\n" {
			t.Errorf("Expected HSET to return :1, got %s", reply)
		}
	})

	// Test HGET command
	t.Run("HGET", func(t *testing.T) {
		// Send HGET command
		sendCommand(conn, "*3\r\n$4\r\nHGET\r\n$4\r\nuser\r\n$4\r\nname\r\n")

		// Read reply
		reply := readReply(conn)

		// Expected reply is "$4\r\nJohn\r\n" (bulk reply, content is John)
		if reply != "$4\r\nJohn\r\n" {
			t.Errorf("Expected HGET to return $4\\r\\nJohn\\r\\n, got %s", reply)
		}
	})

	// Test HEXISTS command
	t.Run("HEXISTS", func(t *testing.T) {
		// Send HEXISTS command, check for existing field
		sendCommand(conn, "*3\r\n$7\r\nHEXISTS\r\n$4\r\nuser\r\n$4\r\nname\r\n")

		// Read reply
		reply := readReply(conn)

		// Expected reply is ":1\r\n" (integer reply 1, field exists)
		if reply != ":1\r\n" {
			t.Errorf("Expected HEXISTS to return :1, got %s", reply)
		}

		// Send HEXISTS command, check for non-existing field
		sendCommand(conn, "*3\r\n$7\r\nHEXISTS\r\n$4\r\nuser\r\n$3\r\nage\r\n")

		// Read reply
		reply = readReply(conn)

		// Expected reply is ":0\r\n" (integer reply 0, field does not exist)
		if reply != ":0\r\n" {
			t.Errorf("Expected HEXISTS to return :0, got %s", reply)
		}
	})

	// Test HSET updating existing field
	t.Run("HSET_UPDATE", func(t *testing.T) {
		// Send HSET command, update existing field
		sendCommand(conn, "*4\r\n$4\r\nHSET\r\n$4\r\nuser\r\n$4\r\nname\r\n$4\r\nJane\r\n")

		// Read reply
		reply := readReply(conn)

		// Expected reply is ":0\r\n" (integer reply 0, updated existing field)
		if reply != ":0\r\n" {
			t.Errorf("Expected HSET to return :0, got %s", reply)
		}

		// Verify field is updated
		sendCommand(conn, "*3\r\n$4\r\nHGET\r\n$4\r\nuser\r\n$4\r\nname\r\n")
		reply = readReply(conn)
		if reply != "$4\r\nJane\r\n" {
			t.Errorf("Expected HGET to return $4\\r\\nJane\\r\\n, got %s", reply)
		}
	})

	// Test HMSET command
	t.Run("HMSET", func(t *testing.T) {
		// Send HMSET command
		sendCommand(conn, "*8\r\n$5\r\nHMSET\r\n$4\r\nuser\r\n$4\r\nname\r\n$4\r\nJohn\r\n$3\r\nage\r\n$2\r\n30\r\n$8\r\nusername\r\n$9\r\njohndoe42\r\n")
		// Read reply
		reply := readReply(conn)

		// Expected reply is "+OK\r\n" (status reply, operation successful)
		if reply != "+OK\r\n" {
			t.Errorf("Expected HMSET to return +OK, got %s", reply)
		}
	})

	// Test HMGET command
	t.Run("HMGET", func(t *testing.T) {
		// Send HMGET command
		sendCommand(conn, "*5\r\n$5\r\nHMGET\r\n$4\r\nuser\r\n$4\r\nname\r\n$3\r\nage\r\n$5\r\nemail\r\n")

		// Read reply
		reply := readReply(conn)

		// Parse multi-bulk reply (simplified, just check if it starts with "*3")
		if reply[0:3] != "*3\r" {
			t.Errorf("Expected HMGET to return multi-bulk reply with 3 elements, got %s", reply)
		}
	})

	// Test HGETALL command
	t.Run("HGETALL", func(t *testing.T) {
		// Send HGETALL command
		sendCommand(conn, "*2\r\n$7\r\nHGETALL\r\n$4\r\nuser\r\n")

		// Read reply
		reply := readReply(conn)

		// Parse multi-bulk reply (simplified, just check if it starts with "*")
		if reply[0:1] != "*" {
			t.Errorf("Expected HGETALL to return multi-bulk reply, got %s", reply)
		}
	})

	// Test HDEL command
	t.Run("HDEL", func(t *testing.T) {
		// Send HDEL command
		sendCommand(conn, "*3\r\n$4\r\nHDEL\r\n$4\r\nuser\r\n$4\r\nname\r\n")

		// Read reply
		reply := readReply(conn)

		// Expected reply is ":1\r\n" (integer reply 1, one field deleted)
		if reply != ":1\r\n" {
			t.Errorf("Expected HDEL to return :1, got %s", reply)
		}
	})
}

// Send command to Redis server
func sendCommand(conn net.Conn, cmd string) {
	_, err := conn.Write([]byte(cmd))
	if err != nil {
		panic(fmt.Sprintf("Failed to send command: %v", err))
	}
}

// Read reply from Redis server
func readReply(conn net.Conn) string {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		panic(fmt.Sprintf("Failed to read reply: %v", err))
	}
	return string(buf[:n])
}

func TestMain(m *testing.M) {
	// 在测试前启动服务器
	go func() {
		// 使用硬编码配置而不是解析命令行参数
		config.Properties = &config.ServerProperties{
			Bind: "localhost",
			Port: 6380,
		}

		// 启动服务器
		tcp.ListenAndServeWithSignal(
			&tcp.Config{
				Address: fmt.Sprintf("%s:%d", config.Properties.Bind, config.Properties.Port),
			},
			handler.MakeHandler())
	}()

	// 等待服务器启动
	time.Sleep(1 * time.Second)

	// 运行测试
	code := m.Run()

	// 退出测试
	os.Exit(code)
}
