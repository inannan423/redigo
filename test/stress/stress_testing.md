# Redigo 压力测试工具完整指南

## 📖 目录

1. [项目概述](#项目概述)
2. [快速开始](#快速开始)
3. [压力测试工具](#压力测试工具)
4. [测试场景详解](#测试场景详解)
5. [性能基准与分析](#性能基准与分析)
6. [Go基准测试](#Go基准测试)
7. [性能调优指南](#性能调优指南)
8. [故障排除](#故障排除)
9. [扩展功能](#扩展功能)

## 项目概述

Redigo 压力测试工具集提供了全面的性能测试解决方案，包含三种不同类型的测试工具：

### 🎯 测试工具类型

1. **独立压力测试工具** (`main.go`) - 功能最全面的命令行压力测试工具
2. **便捷脚本** (`stress_test.sh`) - 预定义场景的快速测试脚本
3. **Go Benchmark 测试** (`benchmark_test.go`) - 集成到 Go 测试框架的基准测试

### 📁 目录结构

```
test/stress/
├── README.md                    # 使用说明
├── stress_testing.md           # 本文件 - 完整指南
├── main.go                     # 主要压力测试工具
├── stress_test.sh              # 便捷测试脚本
└── benchmark_test.go           # Go基准测试
```

### 🚀 支持的Redis命令

- **基础操作**: SET, GET, PING
- **哈希操作**: HSET, HGET
- **列表操作**: LPUSH, RPUSH, LPOP
- **集合操作**: SADD, SMEMBERS

## 快速开始

### 环境准备

```bash
# 1. 启动 Redigo 服务器 (在项目根目录)
cd /path/to/redigo
go run main.go

# 2. 进入压力测试目录 (另开终端)
cd test/stress
```

### 快速测试

```bash
# 使用便捷脚本运行所有测试场景
./stress_test.sh --scenarios

# 或从项目根目录运行
cd /path/to/redigo
./run_stress_test.sh --scenarios
```

## 压力测试工具

### 1. 独立压力测试工具 (main.go)

功能最全面的压力测试工具，支持灵活的参数配置。

#### 基本使用

```bash
# 进入压力测试目录
cd test/stress

# 基础测试 (默认：10连接，100请求)
go run main.go

# 自定义参数测试
go run main.go -c 50 -n 1000 -cmd SET -d 64

# 基于时间的测试
go run main.go -c 100 -t 30s -cmd GET --progress

# 测试不同命令
go run main.go -c 20 -n 500 -cmd HSET -d 128
go run main.go -c 50 -n 800 -cmd PING

# 测试集合操作
go run main.go -c 40 -n 800 -cmd SADD -d 64
go run main.go -c 30 -n 600 -cmd SMEMBERS
```

#### 命令行参数详解

| 参数 | 默认值 | 描述 | 示例 |
|------|--------|------|------|
| `-h` | localhost | Redis 服务器主机 | `-h 192.168.1.100` |
| `-p` | 6380 | Redis 服务器端口 | `-p 6379` |
| `-c` | 10 | 并发连接数 | `-c 50` |
| `-n` | 100 | 每个连接的请求数 | `-n 1000` |
| `-t` | 0 | 测试持续时间 | `-t 30s`, `-t 2m` |
| `-k` | test | 测试键的前缀 | `-k mytest` |
| `-cmd` | SET | 要测试的命令 | `-cmd GET` |
| `-d` | 64 | 测试数据大小（字节） | `-d 256` |
| `--progress` | false | 显示测试进度 | `--progress` |
| `--help` | - | 显示帮助信息 | `--help` |

#### 测试结果解释

```
============================================================
STRESS TEST RESULTS
============================================================
Total Requests:     50000       # 总请求数
Successful:         49950       # 成功请求数
Failed:             50          # 失败请求数
Success Rate:       99.90%      # 成功率
Test Duration:      2.45s       # 测试持续时间
Queries Per Second: 20408.16    # 每秒查询数

Latency Statistics:
  Min:              245µs       # 最小延迟
  Max:              12.3ms      # 最大延迟
  Average:          1.2ms       # 平均延迟
  50th percentile:  1.1ms       # 50% 请求的延迟
  95th percentile:  2.8ms       # 95% 请求的延迟
  99th percentile:  5.2ms       # 99% 请求的延迟
============================================================
```

### 2. 便捷测试脚本 (stress_test.sh)

预定义多种测试场景的便捷脚本，适合快速性能评估。

#### 基本使用

```bash
# 在压力测试目录运行
cd test/stress

# 显示帮助信息
./stress_test.sh --help

# 运行所有预定义场景
./stress_test.sh --scenarios

# 自定义单个测试
./stress_test.sh -c 100 -n 2000 -cmd SET

# 带进度显示的测试
./stress_test.sh -c 50 -t 30s -cmd GET --progress
```

#### 从项目根目录运行

```bash
# 使用便捷脚本 (推荐)
./run_stress_test.sh --scenarios
./run_stress_test.sh -c 40 -n 800 -cmd SADD -d 64
```

#### 脚本参数

| 参数 | 描述 | 示例 |
|------|------|------|
| `-h, --host` | Redis服务器主机 | `-h localhost` |
| `-p, --port` | Redis服务器端口 | `-p 6380` |
| `-c, --connections` | 并发连接数 | `-c 50` |
| `-n, --requests` | 每连接请求数 | `-n 1000` |
| `-t, --time` | 测试持续时间 | `-t 30s` |
| `-cmd, --command` | 测试命令 | `-cmd SADD` |
| `-d, --data-size` | 数据大小 | `-d 128` |
| `--progress` | 显示进度 | `--progress` |
| `--scenarios` | 运行所有场景 | `--scenarios` |

## 测试场景详解

### 预定义测试场景

便捷脚本包含8个精心设计的测试场景：

#### Scenario 1: Basic SET operations
- **目标**: 测试基本写入性能
- **配置**: 50连接，1000请求，64字节数据
- **用途**: 评估基础写入能力

#### Scenario 2: GET operations  
- **目标**: 测试基本读取性能
- **配置**: 30连接，1500请求，64字节数据
- **用途**: 评估基础读取能力

#### Scenario 3: PING test
- **目标**: 测试网络延迟和连接性能
- **配置**: 100连接，500请求
- **用途**: 评估最低延迟基准

#### Scenario 4: Hash operations
- **目标**: 测试哈希数据结构性能
- **配置**: 20连接，800请求，128字节数据
- **用途**: 评估复杂数据结构处理能力

#### Scenario 5: List operations
- **目标**: 测试列表数据结构性能
- **配置**: 25连接，600请求，32字节数据
- **用途**: 评估队列操作性能

#### Scenario 6: Set ADD operations ✨ 新增
- **目标**: 测试集合添加操作性能
- **配置**: 40连接，800请求，64字节数据
- **用途**: 评估集合写入性能

#### Scenario 7: Set MEMBERS operations ✨ 新增
- **目标**: 测试集合成员查询性能
- **配置**: 30连接，600请求
- **用途**: 评估集合读取性能

#### Scenario 8: Mixed Set operations ✨ 新增
- **目标**: 测试混合集合操作性能
- **配置**: 20连接，400 SADD + 400 SMEMBERS
- **用途**: 评估真实场景下的集合性能

### 自定义测试场景

#### 高并发测试
```bash
# 高并发SET测试
./stress_test.sh -c 200 -n 500 -cmd SET -d 64

# 高并发集合测试
./stress_test.sh -c 100 -n 500 -cmd SADD -d 128
```

#### 大数据测试
```bash
# 1KB数据测试
./stress_test.sh -c 20 -n 200 -cmd SET -d 1024

# 4KB数据测试
./stress_test.sh -c 10 -n 100 -cmd SET -d 4096
```

#### 长时间压力测试
```bash
# 5分钟持续测试
./stress_test.sh -c 50 -t 5m -cmd GET --progress

# 集合操作长时间测试
./stress_test.sh -c 30 -t 2m -cmd SMEMBERS --progress
```

## Go基准测试

### 运行基准测试

```bash
# 进入压力测试目录
cd test/stress

# 运行所有基准测试
go test -bench=. -benchmem

# 运行特定基准测试
go test -bench=BenchmarkSET -benchmem

# 长时间基准测试
go test -bench=. -benchtime=10s

# 不同CPU配置测试
go test -bench=. -cpu=1,2,4

# 详细输出
go test -bench=. -benchmem -v
```

### 可用的基准测试

#### 单命令基准测试
- `BenchmarkSET` - SET 命令性能测试
- `BenchmarkGET` - GET 命令性能测试  
- `BenchmarkPING` - PING 命令性能测试
- `BenchmarkHSET` - HSET 命令性能测试
- `BenchmarkHGET` - HGET 命令性能测试
- `BenchmarkLPUSH` - LPUSH 命令性能测试
- `BenchmarkSADD` - SADD 命令性能测试 ✨ 新增

#### 复合基准测试
- `BenchmarkMixed` - 混合命令性能测试
- `BenchmarkConcurrentConnections` - 并发连接性能测试

### 基准测试结果解释

```
BenchmarkSET-8    	   10000	    134567 ns/op	     128 B/op	       3 allocs/op
```

**结果解读**:
- `BenchmarkSET-8`: 基准测试名称（8表示GOMAXPROCS）
- `10000`: 运行次数
- `134567 ns/op`: 每次操作平均耗时（纳秒）
- `128 B/op`: 每次操作平均内存分配
- `3 allocs/op`: 每次操作平均内存分配次数

## 故障排除

### 常见问题与解决方案

#### 1. 连接失败
```bash
# 问题: Cannot connect to Redis server
# 检查服务器状态
netstat -tlnp | grep 6380
ps aux | grep redigo

# 解决方案
cd /path/to/redigo
go run main.go
```

#### 2. 性能异常低

**诊断步骤**:
```bash
# 检查系统资源
top
free -h
iostat 1

# 检查网络状态
netstat -i
ss -tuln
```

**常见原因**:
- 系统资源不足
- 网络延迟高
- Redis服务器过载
- 文件描述符限制

#### 3. 高错误率

**检查命令**:
```bash
# 降低并发数重试
go run main.go -c 10 -n 100 -cmd SET

# 增加超时时间
go run main.go -c 50 -n 1000 -cmd SET -timeout 5s
```

#### 4. 内存不足
```bash
# 检查内存使用
free -h
cat /proc/meminfo

# 减少并发和数据大小
go run main.go -c 20 -n 500 -d 32 -cmd SET
```

### 调试技巧

#### 启用详细日志
```bash
# 显示进度信息
go run main.go -c 50 -n 1000 --progress

# 使用小规模测试调试
go run main.go -c 1 -n 10 -cmd SET
```

#### 分步测试
```bash
# 1. 测试连接
go run main.go -c 1 -n 1 -cmd PING

# 2. 测试基础操作
go run main.go -c 5 -n 10 -cmd SET

# 3. 逐步增加负载
go run main.go -c 10 -n 100 -cmd SET
go run main.go -c 50 -n 1000 -cmd SET
```

## 扩展功能

### 自定义命令支持

如需添加新的Redis命令支持，可以修改 `main.go`:

```go
case "CUSTOM":
    // 添加自定义命令
    return fmt.Sprintf("*2\r\n$6\r\nCUSTOM\r\n$%d\r\n%s\r\n", len(key), key)
```

### 结果导出

#### 保存测试结果
```bash
# 导出到文件
go run main.go -c 50 -n 1000 > test_results.txt

# 带时间戳的结果
go run main.go -c 50 -n 1000 | tee results_$(date +%Y%m%d_%H%M%S).log
```

### 自动化批量测试

#### 创建批量测试脚本
```bash
#!/bin/bash
# batch_test.sh

echo "开始批量性能测试..."

# 不同并发数测试
for connections in 10 20 50 100; do
    echo "测试并发数: $connections"
    go run main.go -c $connections -n 1000 -cmd SET
    echo "---"
done

# 不同命令测试
for cmd in SET GET PING SADD SMEMBERS; do
    echo "测试命令: $cmd"
    go run main.go -c 50 -n 1000 -cmd $cmd
    echo "---"
done
```