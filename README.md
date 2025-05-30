# Redigo

🚀 **从零开始，用 Go 语言手写一个完整的 Redis 服务器！**

Redigo 不仅仅是一个 Redis 的 Go 语言实现，更是一个**完整的学习教程项目**。通过详细的分步指南，你将深入理解 Redis 的核心原理，学会从 0 构建一个高性能的内存数据库。

## 🎯 特点

- **10+ 章节**的详细教程，从基础到高级循序渐进
- **代码演进式**教学，每个功能都有对应的 Git 分支
- 有较为详细的原理讲解
- **不跳过任何细节，适合初学者和有经验的开发者**
- 支持集群模式、数据持久化、多种数据结构
- 遵循 Redis 官方协议规范（RESP）
- 可以与 Redis 客户端兼容

## 📖 教学指南

### 🌐 在线文档
📚 **[完整教学指南 - 在线阅读](https://redigo.vercel.app/)** 

本地运行教学文档：
```bash
cd guide
npm install
npm run dev
# 访问 http://localhost:3000
```

### 教程目录

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

## ✨ 核心特性

### 🎯 已实现功能
- ✅ **网络层**：TCP 服务器
- ✅ **协议层**：RESP 协议解析器
- ✅ **存储引擎**：内存数据库
- ✅ **数据结构**：String、List、Hash、Set、ZSet
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
- Node.js 18+ (可选，用于运行教学文档)

### 方式一：跟随教程学习 📚
```bash
# 1. 克隆项目
git clone https://github.com/inannan423/redigo.git
cd redigo

# 2. 启动教学文档（可选，可以访问 https://redigo.vercel.app）
cd guide
npm install
npm run dev
# 访问 http://localhost:3000 开始学习


# 3. 按教程进度切换分支学习
git checkout tcp-server    # 第一章：TCP 服务器
git checkout resp-parser   # 第二章：RESP 协议
git checkout database      # 第三章：数据库核心
# ... 更多分支见教程
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

## 📊 性能基准

TODO: 添加性能测试结果

## 🗓 TODO

- [ ] 完善集群模式
- [ ] 实现更多 Redis 命令
- [ ] 增加更多数据结构支持
- [ ] 提升测试覆盖率

## 🤝 贡献指南

我们欢迎各种形式的贡献！

- 🐛 **Bug 修复**：发现问题请提交 Issue，或者直接提交 Commit
- 📚 **文档改进**：让教程更清晰易懂
- ✨ **新功能**：实现更多 Redis 命令
- 🎯 **性能优化**：让 Redigo 更快
- 🧪 **测试用例**：提高代码覆盖率
- 📝 **教程反馈**：改进学习体验

### 如何贡献
1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交改动 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 💬 学习交流

- 📖 **教程问题**：查看[在线文档](https://redigo-guide.vercel.app)
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