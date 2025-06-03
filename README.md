# Redigo

<img width="1506" alt="poster" src="https://github.com/user-attachments/assets/de61424e-99d5-45d2-9e24-f6d46b295bda" />


🚀 **从零开始，用 Go 语言手写一个完整的 Redis 服务器**

带有完整项目笔记，没有遗漏任何细节，适合用于学习 Redis 的实现原理。

## 📖 笔记

### 🌐 在线笔记
📚 **[完整笔记 - 在线阅读](https://redigo.vercel.app/)** 

本地运行笔记：
```bash
cd guide
npm install
npm run dev
# 访问 http://localhost:3000
```

### 笔记目录

#### 🏗️ 基础架构
1. **[TCP 服务器搭建](https://redigo.vercel.app/tcp)** 

2. **[RESP 协议解析](https://redigo.vercel.app/resp)**

3. **[内存数据库核心](https://redigo.vercel.app/database)**

#### 🔧 数据结构篇
4. **[哈希表实现](https://redigo.vercel.app/hash)**

5. **[链表结构](https://redigo.vercel.app/list)**

6. **[集合实现](https://redigo.vercel.app/set)**

7. **[有序集合](https://redigo.vercel.app/zset)**

#### 🚀 高级特性
8. **[压力测试与性能评估](https://redigo.vercel.app/stress-testing)**

9. **[并发安全](https://redigo.vercel.app/concurrency)**

10. **[数据持久化](https://redigo.vercel.app/persistence)**

11. **[集群模式](https://redigo.vercel.app/cluster)**

## ✨ 核心特性

### 🎯 已实现功能
- ✅ **网络层**：TCP 服务器
- ✅ **协议层**：RESP 协议解析器
- ✅ **存储引擎**：内存数据库
- ✅ **数据结构**：String、List、Hash、Set、ZSet
- ✅ **压力测试**：完整的性能测试工具集
- ✅ **并发安全**：Key级别细粒度锁定机制
- ✅ **持久化**：AOF (Append Only File) 机制
- ✅ **集群**：一致性哈希

### 🔧 支持的 Redis 命令

#### 🔑 键操作命令
```bash
DEL key [key ...]              # 删除一个或多个键
EXISTS key [key ...]           # 检查键是否存在
FLUSHDB                        # 清空当前数据库
TYPE key                       # 获取键的数据类型
RENAME key newkey              # 重命名键
RENAMENX key newkey            # 仅当新键不存在时重命名
KEYS pattern                   # 查找匹配模式的键
```

#### 📝 字符串操作
```bash
SET key value                  # 设置键值对
GET key                        # 获取键的值
SETNX key value               # 仅当键不存在时设置
GETSET key value              # 设置新值并返回旧值
STRLEN key                    # 获取字符串长度
```

#### 📋 列表操作
```bash
LPUSH key value [value ...]   # 从左侧插入元素
RPUSH key value [value ...]   # 从右侧插入元素
LPOP key                      # 从左侧弹出元素
RPOP key                      # 从右侧弹出元素
LRANGE key start stop         # 获取指定范围的元素
LLEN key                      # 获取列表长度
LINDEX key index              # 获取指定位置的元素
LSET key index value          # 设置指定位置的元素值
```

#### 🏠 哈希操作
```bash
HSET key field value          # 设置哈希字段
HGET key field                # 获取哈希字段值
HEXISTS key field             # 检查哈希字段是否存在
HDEL key field [field ...]    # 删除哈希字段
HLEN key                      # 获取哈希字段数量
HGETALL key                   # 获取所有字段和值
HKEYS key                     # 获取所有字段名
HVALS key                     # 获取所有字段值
HMGET key field [field ...]   # 获取多个字段值
HMSET key field value [field value ...]  # 设置多个字段
HSETNX key field value        # 仅当字段不存在时设置
```

#### 🎯 集合操作
```bash
SADD key member [member ...]  # 添加集合成员
SCARD key                     # 获取集合成员数量
SISMEMBER key member          # 检查成员是否在集合中
SMEMBERS key                  # 获取所有集合成员
SREM key member [member ...]  # 删除集合成员
SPOP key [count]              # 随机弹出集合成员
SRANDMEMBER key [count]       # 随机获取集合成员
SUNION key [key ...]          # 计算集合并集
SUNIONSTORE dest key [key ...]  # 存储集合并集
SINTER key [key ...]          # 计算集合交集
SINTERSTORE dest key [key ...]  # 存储集合交集
SDIFF key [key ...]           # 计算集合差集
SDIFFSTORE dest key [key ...]   # 存储集合差集
```

#### ⚖️ 有序集合操作
```bash
ZADD key score member [score member ...]  # 添加有序集合成员
ZSCORE key member             # 获取成员分数
ZCARD key                     # 获取有序集合成员数量
ZRANGE key start stop [WITHSCORES]  # 按索引范围获取成员
ZREM key member [member ...]  # 删除有序集合成员
ZCOUNT key min max            # 统计分数范围内的成员数量
ZRANK key member              # 获取成员排名
```

#### 🔧 系统命令
```bash
PING                          # 测试连接
SELECT index                  # 选择数据库
```

## 🚀 快速开始

### 环境要求
- Go 1.21+
- Git
- Node.js 18+ (可选，用于运行笔记)

### 查看笔记的方式
```bash
# 1. 克隆项目
git clone https://github.com/inannan423/redigo.git
cd redigo

# 2. 启动笔记（可选，可以访问 https://redigo.vercel.app）
cd guide
npm install
npm run dev
# 访问 http://localhost:3000 开始学习


# 3. 按笔记进度切换分支学习
git checkout tcp-server    # 第一章：TCP 服务器
git checkout resp-parser   # 第二章：RESP 协议
git checkout database      # 第三章：数据库核心
# ... 更多分支见笔记
```

### 方式二：直接运行完整版 🏃‍♂️
```bash
# 1. 克隆项目
git clone https://github.com/inannan423/redigo.git
cd redigo

# 2. 启动单机模式
go run main.go

# 3. 启动集群模式（需要配置 redis.conf）
# 编辑 redis.conf 设置集群节点
go run main.go
```

### 客户端连接测试
```bash
# 使用 Redis 官方客户端
redis-cli -h localhost -p 6379

# 测试基本命令
127.0.0.1:6379> SET hello world
OK
127.0.0.1:6379> GET hello
"world"
127.0.0.1:6379> PING
PONG
```

## 📊 性能基准与压力测试

Redigo 提供了完善的压力测试工具来评估服务器性能。所有压力测试工具都位于 `test/stress/` 目录中。

### 🚀 快速开始压力测试

```bash
# 1. 启动 Redigo 服务器
go run main.go

# 2. 运行基础压力测试（另开终端）
./run_stress_test.sh -c 50 -n 1000 -cmd SET

# 3. 运行集合操作测试
./run_stress_test.sh -c 40 -n 800 -cmd SADD -d 64

# 4. 运行所有预定义测试场景
./run_stress_test.sh --scenarios
```

### 📈 性能测试工具

#### 1. 独立压力测试工具
功能最全面的压力测试工具，支持多种参数配置：

```bash
# 进入压力测试目录
cd test/stress

# 基础使用
go run main.go

# 自定义参数测试
go run main.go -c 50 -n 1000 -cmd SET -d 64

# 基于时间的测试
go run main.go -c 100 -t 30s -cmd GET --progress

# 测试集合操作
go run main.go -c 40 -n 800 -cmd SADD -d 64
go run main.go -c 30 -n 600 -cmd SMEMBERS
```

**支持的命令**: 
- **基础操作**: SET, GET, PING
- **哈希操作**: HSET, HGET  
- **列表操作**: LPUSH, RPUSH, LPOP
- **集合操作**: SADD, SMEMBERS

#### 2. 便捷测试脚本
预定义多种测试场景的脚本，包括集合操作测试：

```bash
# 从项目根目录运行
./run_stress_test.sh --scenarios  # 运行所有8个预定义场景

# 或直接在压力测试目录运行
cd test/stress
./stress_test.sh --scenarios
```

**测试场景包括**:
- Scenario 1-2: 基础操作（SET, GET）
- Scenario 3: 连接测试（PING）
- Scenario 4: 哈希操作 (HSET)
- Scenario 5: 列表操作 (LPUSH)
- Scenario 6-8: 集合操作 (SADD, SMEMBERS, 混合测试)

#### 3. Go Benchmark 测试
集成到 Go 测试框架的基准测试：

```bash
# 进入压力测试目录
cd test/stress

# 运行所有基准测试
go test -bench=. -benchmem

# 运行特定基准测试
go test -bench=BenchmarkSET -benchmem

# 并发安全测试
go test -race -bench=. -benchtime=10s
```

### 📊 性能测试结果

> 测试环境：Apple M2 芯片，16GB 内存，macOS Sonoma

#### 🏆 综合性能表现

| 测试场景 | 并发连接 | 总请求数 | QPS | 成功率 | 最小延迟 | 最大延迟 |
|---------|---------|---------|-----|-------|----------|----------|
| **PING测试** | 100 | 50,000 | **103,983** | 100% | 11.6µs | 48.2ms |
| **集合查询** | 30 | 18,000 | **99,470** | 100% | 14.8µs | 20.8ms |
| **字符串写入** | 50 | 50,000 | **96,388** | 100% | 12.5µs | 12.8ms |
| **列表操作** | 25 | 15,000 | **92,092** | 100% | 15.3µs | 5.3ms |
| **集合写入** | 40 | 32,000 | **87,480** | 100% | 12.7µs | 17.3ms |
| **字符串读取** | 30 | 45,000 | **78,369** | 100% | 12.2µs | 44.6ms |
| **哈希操作** | 20 | 16,000 | **5,087** | 100% | 52.9µs | 73.6ms |

#### 📋 详细测试结果

```bash
# 场景1：字符串操作性能
SET Operations (50 connections, 1000 requests each)
✅ 50,000 requests completed in 518.74ms
📊 QPS: 96,387.60 | Success Rate: 100%

# 场景2：PING连接测试  
PING Test (100 connections, 500 requests each)
✅ 50,000 requests completed in 480.85ms
📊 QPS: 103,982.58 | Success Rate: 100%

# 场景3：集合操作性能
SADD Operations (40 connections, 800 requests each)
✅ 32,000 requests completed in 365.80ms  
📊 QPS: 87,479.98 | Success Rate: 100%

SMEMBERS Operations (30 connections, 600 requests each)
✅ 18,000 requests completed in 180.96ms
📊 QPS: 99,470.27 | Success Rate: 100%
```

📋 **完整压力测试指南**: [test/stress/stress_testing.md](test/stress/stress_testing.md)

## 🗓 TODO

- [ ] 完善集群模式
- [ ] 实现更多 Redis 命令
- [ ] 增加更多数据结构支持
- [ ] 提升测试覆盖率

## 🤝 贡献指南

欢迎贡献！

- 🐛 **Bug 修复**：发现问题请提交 Issue，或者直接提交 Commit
- 📚 **文档改进**：让笔记更清晰易懂
- ✨ **新功能**：实现更多 Redis 命令
- 🎯 **性能优化**：优化 Redigo 的性能
- 🧪 **测试用例**：提高代码覆盖率

### 如何贡献
1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交改动 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 💬 学习交流

- 📖 **笔记问题**：查看[在线文档](https://redigo.vercel.app)
- 🐛 **Bug 反馈**：提交 [Issue](https://github.com/inannan423/redigo/issues)
- 💡 **功能建议**：在 Issues 中标记 `enhancement`
- 📧 **邮件咨询**：jetzihan@outlook.com

## 📜 开源协议

本项目采用 GPL-3.0 协议，详情请查看 [LICENSE](LICENSE) 文件。

## 🙏 致谢

- [Godis](https://github.com/HDT3213/godis) 本项目学习了 Godis 的设计思路和部分实现，感谢大佬们的贡献！

---

⭐ **如果这个项目对你有帮助，请给我一个 Star！**

📧 **有问题？** 欢迎提交 [Issue](https://github.com/inannan423/redigo/issues) 或发邮件讨论。
