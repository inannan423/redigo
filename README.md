# Redigo

<img width="1506" alt="poster" src="https://github.com/user-attachments/assets/de61424e-99d5-45d2-9e24-f6d46b295bda" />


🚀 **从零开始，用 Go 语言手写一个完整的 Redis 服务器**

带有完整项目笔记，没有遗漏任何细节，适合用于学习 Redis 的实现原理。

## 📖 笔记

### 🌐 在线笔记
📚 **[完整笔记 - 在线阅读](https://redigo.vercel.app/)** 

> Deepwiki：[AI 生成的技术文档，可以和代码对话](https://deepwiki.com/inannan423/redigo)

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

8. **[数据持久化](https://redigo.vercel.app/persistence)**

9. **[集群模式](https://redigo.vercel.app/cluster)**

10. **[并发安全](https://redigo.vercel.app/concurrency)**

## ✨ 核心特性

### 🎯 已实现功能
- ✅ **网络层**：TCP 服务器
- ✅ **协议层**：RESP 协议解析器
- ✅ **存储引擎**：内存数据库
- ✅ **数据结构**：String、List、Hash、Set、ZSet
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
redis-cli -h localhost -p 6380

# 测试基本命令
127.0.0.1:6380> SET hello world
OK
127.0.0.1:6380> GET hello
"world"
127.0.0.1:6380> PING
PONG
```

## 📊 性能基准与压力测试

Redis 提供了 `redis-benchmark` 工具来测试性能，以下是详细的使用指导：

### 📋 基础用法

#### 安装 redis-benchmark
确保已安装 Redis 客户端工具：
```bash
# macOS
brew install redis

# Ubuntu/Debian
sudo apt-get install redis-tools

# CentOS/RHEL
sudo yum install redis
```

#### 基本测试命令
```bash
# 启动 Redigo 服务器
go run main.go

# 在另一个终端运行基准测试
redis-benchmark -h localhost -p 6380 -n 100000 -c 50
```

### 🎯 常用测试场景

#### 字符串操作性能测试
```bash
# SET 命令测试
redis-benchmark -h localhost -p 6380 -n 100000 -c 50 -t set

# GET 命令测试
redis-benchmark -h localhost -p 6380 -n 100000 -c 50 -t get

# 混合 SET/GET 测试
redis-benchmark -h localhost -p 6380 -n 50000 -c 25 -t set,get
```

#### 列表操作性能测试
```bash
# LPUSH 测试
redis-benchmark -h localhost -p 6380 -n 50000 -c 25 -t lpush

# LPOP 测试
redis-benchmark -h localhost -p 6380 -n 50000 -c 25 -t lpop

# LRANGE 测试
redis-benchmark -h localhost -p 6380 -n 10000 -c 10 -t lrange_100,lrange_300,lrange_500
```

#### 哈希操作性能测试
```bash
# HSET 测试
redis-benchmark -h localhost -p 6380 -n 50000 -c 25 -t hset,hget
```

#### 集合操作性能测试
```bash
# SADD 测试
redis-benchmark -h localhost -p 6380 -n 50000 -c 25 -t sadd,spop
```

#### 有序集合操作性能测试
```bash
# ZADD 测试
redis-benchmark -h localhost -p 6380 -n 50000 -c 25 -t zadd,zrem

# ZRANGE 测试
redis-benchmark -h localhost -p 6380 -n 10000 -c 10 -t zadd,zrange_100,zrange_300,zrange_500
```

#### 一次测试所有命令
```bash
redis-benchmark -h localhost -p 6380 -n 100000 -c 50 -t set,get,lpush,lpop,lrange,hset,hget,hdel,hlen,hkeys,hvals,hmget,hmset,hsetnx,sadd,spop,smembers,srem,sinter,sinterstore,sdiff,sdiffstore,zadd,zrem,zcard,zrange,zcount,zrank
```

### 📊 参数详解

| 参数 | 描述 | 示例 |
|------|------|------|
| `-h <hostname>` | Redis 服务器地址 | `-h localhost` |
| `-p <port>` | Redis 服务器端口 | `-p 6380` |
| `-n <requests>` | 总请求数 | `-n 100000` |
| `-c <clients>` | 并发连接数 | `-c 50` |
| `-d <size>` | 数据大小（字节） | `-d 1024` |
| `-t <tests>` | 指定测试命令 | `-t set,get,lpush` |
| `-k <boolean>` | 保持连接 | `-k 1` |
| `-r <keyspacelen>` | 键空间大小 | `-r 100000` |
| `-P <pipeline>` | 管道请求数 | `-P 10` |
| `-q` | 静默模式，只显示结果 | `-q` |
| `--csv` | CSV 格式输出 | `--csv` |

### 📈 性能指标解读

测试完成后，redis-benchmark 会显示以下关键指标：

```
====== SET ======
  100000 requests completed in 1.23 seconds
  50 parallel clients
  3 bytes payload
  keep alive: 1

99.95% <= 1 milliseconds
100.00% <= 2 milliseconds
81234.56 requests per second
```

**关键指标说明：**
- **Requests per second (RPS)**：每秒处理的请求数，越高越好
- **Latency percentiles**：延迟百分位数，显示响应时间分布
- **平均延迟**：所有请求的平均响应时间
- **吞吐量**：服务器的数据处理能力

### 📋 实际测试结果分析

测试环境： MacBook Pro M2, 16GB RAM, macOS 15.2

基于 `redis-benchmark -h localhost -p 6380 -n 100000 -c 50 -t set,get,lpush,lpop,lrange,hset,hget,hdel,hlen,hkeys,hvals,hmget,hmset,hsetnx,sadd,spop,smembers,srem,sinter,sinterstore,sdiff,sdiffstore,zadd,zrem,zcard,zrange,zcount,zrank` 的综合测试结果：

#### 🚀 核心操作性能表现

| 操作类型 | QPS | 平均延迟(ms) | P95延迟(ms) | P99延迟(ms) |
|---------|-----|-------------|-------------|-------------|
| **SET** | 148,368 | 0.193 | 0.295 | 0.671 |
| **GET** | 149,031 | 0.186 | 0.279 | 0.447 |
| **LPUSH** | 163,666 | 0.176 | 0.247 | 0.399 |
| **LPOP** | 153,610 | 0.184 | 0.279 | 0.359 |
| **HSET** | 163,132 | 0.174 | 0.247 | 0.335 |
| **SADD** | 143,062 | 0.193 | 0.287 | 0.407 |
| **SPOP** | 160,772 | 0.175 | 0.255 | 0.311 |
| **ZADD** | 162,866 | 0.177 | 0.247 | 0.359 |

#### 📋 范围查询性能分析

| LRANGE操作 | QPS | 平均延迟(ms) | P95延迟(ms) | P99延迟(ms) | 适用场景 |
|-----------|-----|-------------|-------------|-------------|----------|
| **LRANGE_100** | 45,167 | 0.613 | 1.087 | 2.703 | 小数据量查询 |
| **LRANGE_300** | 22,619 | 1.146 | 1.647 | 2.719 | 中等数据量查询 |
| **LRANGE_500** | 15,352 | 1.663 | 2.279 | 3.487 | 大数据量查询 |
| **LRANGE_600** | 13,344 | 1.881 | 2.375 | 2.903 | 超大数据量查询 |

#### 🎯 性能亮点

基础操作（SET/GET/LPUSH/HSET等）均达到 **14万+ QPS**，最高性能的 LPUSH 操作达到 **16.3万+ QPS**，所有基础操作平均延迟均低于 **0.2ms**，P95 延迟保持在 **0.3ms** 以内，P99 延迟控制在 **0.7ms** 以内

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
