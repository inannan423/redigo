package rdb

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"redigo/datastruct/dict"
	"redigo/interface/database"
	"redigo/lib/logger"
	"strconv"
	"time"
)

// Encoder 用于将数据编码为 RDB 格式
type Encoder struct {
	writer   io.Writer
	bufWriter *bufio.Writer
	db       database.Database
	crc64    uint64
}

// NewEncoder 创建一个新的 RDB 编码器
func NewEncoder(w io.Writer, db database.Database) *Encoder {
	bufWriter := bufio.NewWriter(w)
	return &Encoder{
		writer:    w,
		bufWriter: bufWriter,
		db:        db,
		crc64:     0,
	}
}

// Encode 将数据库编码为 RDB 格式
func (e *Encoder) Encode() error {
	// 1. 写入 RDB 魔数和版本号
	if err := e.writeMagic(); err != nil {
		return err
	}

	// 2. 写入辅助字段（如 Redis 版本、创建时间等）
	if err := e.writeAuxFields(); err != nil {
		return err
	}

	// 3. 写入数据库数据
	if err := e.writeDatabases(); err != nil {
		return err
	}

	// 4. 写入结束标记和校验和
	if err := e.writeFooter(); err != nil {
		return err
	}

	// 刷新缓冲区
	return e.bufWriter.Flush()
}

// writeMagic 写入 RDB 魔数和版本号
func (e *Encoder) writeMagic() error {
	magic := []byte(rdbMagic)
	version := []byte(rdbVersion)
	
	data := append(magic, version...)
	_, err := e.bufWriter.Write(data)
	return err
}

// writeAuxFields 写入辅助字段
func (e *Encoder) writeAuxFields() error {
	// 写入 Redis 版本
	if err := e.writeAuxField("redis-ver", "6.0.0"); err != nil {
		return err
	}
	
	// 写入创建时间
	if err := e.writeAuxField("ctime", strconv.FormatInt(time.Now().Unix(), 10)); err != nil {
		return err
	}
	
	return nil
}

// writeAuxField 写入单个辅助字段
func (e *Encoder) writeAuxField(key, value string) error {
	// 写入操作码
	if _, err := e.bufWriter.Write([]byte{rdbOpcodeAuxField}); err != nil {
		return err
	}
	
	// 写入 key
	if err := e.writeString(key); err != nil {
		return err
	}
	
	// 写入 value
	if err := e.writeString(value); err != nil {
		return err
	}
	
	return nil
}

// writeDatabases 写入所有数据库数据
func (e *Encoder) writeDatabases() error {
	// 获取数据库实例
	sdb, ok := e.db.(*database.StandaloneDatabase)
	if !ok {
		return fmt.Errorf("unsupported database type")
	}
	
	// 写入每个数据库
	for i, db := range sdb.GetDBs() {
		if db == nil || db.GetDict() == nil {
			continue
		}
		
		// 获取数据库中的键值对
		dict := db.GetDict()
		if dict.Len() == 0 {
			continue
		}
		
		// 写入数据库选择标记
		if err := e.writeSelectDB(i); err != nil {
			return err
		}
		
		// 写入数据库大小信息
		if err := e.writeResizeDB(dict.Len(), dict.Len()); err != nil {
			return err
		}
		
		// 遍历所有键值对
		dict.ForEach(func(key string, val interface{}) bool {
			// 获取键的过期时间
			expireTime := db.GetExpireTime(key)
			
			// 写入键值对
			if err := e.writeKV(key, val, expireTime); err != nil {
				logger.Error("write kv error: " + err.Error())
			}
			return true
		})
	}
	
	return nil
}

// writeSelectDB 写入数据库选择标记
func (e *Encoder) writeSelectDB(dbIndex int) error {
	if _, err := e.bufWriter.Write([]byte{rdbOpcodeSelectDB}); err != nil {
		return err
	}
	return e.writeLength(uint32(dbIndex))
}

// writeResizeDB 写入数据库大小信息
func (e *Encoder) writeResizeDB(dbSize, expireSize int) error {
	if _, err := e.bufWriter.Write([]byte{rdbOpcodeResizeDB}); err != nil {
		return err
	}
	if err := e.writeLength(uint32(dbSize)); err != nil {
		return err
	}
	return e.writeLength(uint32(expireSize))
}

// writeKV 写入键值对
func (e *Encoder) writeKV(key string, val interface{}, expireTime int64) error {
	// 写入过期时间（如果有）
	if expireTime > 0 {
		now := time.Now().UnixNano() / 1e6 // 转换为毫秒
		if expireTime > now {
			// 写入毫秒级过期时间
			if _, err := e.bufWriter.Write([]byte{rdbOpcodeExpireTimeMS}); err != nil {
				return err
			}
			expireBytes := make([]byte, 8)
			binary.LittleEndian.PutUint64(expireBytes, uint64(expireTime))
			if _, err := e.bufWriter.Write(expireBytes); err != nil {
				return err
			}
		} else {
			// 已过期，跳过
			return nil
		}
	}
	
	// 写入键
	if err := e.writeString(key); err != nil {
		return err
	}
	
	// 根据值的类型写入不同的数据格式
	switch v := val.(type) {
	case []byte:
		// 字符串类型
		if _, err := e.bufWriter.Write([]byte{rdbTypeString}); err != nil {
			return err
		}
		if err := e.writeString(string(v)); err != nil {
			return err
		}
	default:
		// 其他类型暂不实现，记录错误
		logger.Error("unsupported value type")
	}
	
	return nil
}

// writeLength 写入长度编码
func (e *Encoder) writeLength(length uint32) error {
	if length < (1 << 6) {
		// 00xxxxxx - 6位长度
		return e.bufWriter.WriteByte(byte(length))
	} else if length < (1 << 14) {
		// 01xxxxxx xxxxxxxx - 14位长度
		b := make([]byte, 2)
		b[0] = byte(rdbLenEncoding14Bit | (length >> 8))
		b[1] = byte(length & 0xFF)
		_, err := e.bufWriter.Write(b)
		return err
	} else {
		// 10xxxxxx xxxxxxxx xxxxxxxx xxxxxxxx xxxxxxxx - 32位长度
		if err := e.bufWriter.WriteByte(byte(rdbLenEncoding32Bit)); err != nil {
			return err
		}
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, length)
		_, err := e.bufWriter.Write(b)
		return err
	}
}

// writeString 写入字符串
func (e *Encoder) writeString(s string) error {
	// 尝试将字符串解析为整数，如果成功则使用整数编码
	if val, err := strconv.ParseInt(s, 10, 32); err == nil {
		return e.writeInteger(int64(val))
	}
	
	// 写入长度
	if err := e.writeLength(uint32(len(s))); err != nil {
		return err
	}
	
	// 写入字符串内容
	_, err := e.bufWriter.WriteString(s)
	return err
}

// writeInteger 写入整数编码
func (e *Encoder) writeInteger(val int64) error {
	if val >= -128 && val <= 127 {
		// 8位整数
		if err := e.bufWriter.WriteByte(byte(rdbLenEncodingEnc | rdbEncInt8)); err != nil {
			return err
		}
		return e.bufWriter.WriteByte(byte(val))
	} else if val >= -32768 && val <= 32767 {
		// 16位整数
		if err := e.bufWriter.WriteByte(byte(rdbLenEncodingEnc | rdbEncInt16)); err != nil {
			return err
		}
		b := make([]byte, 2)
		binary.LittleEndian.PutUint16(b, uint16(val))
		_, err := e.bufWriter.Write(b)
		return err
	} else if val >= -2147483648 && val <= 2147483647 {
		// 32位整数
		if err := e.bufWriter.WriteByte(byte(rdbLenEncodingEnc | rdbEncInt32)); err != nil {
			return err
		}
		b := make([]byte, 4)
		binary.LittleEndian.PutUint32(b, uint32(val))
		_, err := e.bufWriter.Write(b)
		return err
	}
	
	// 超出范围，作为字符串处理
	return e.writeString(strconv.FormatInt(val, 10))
}

// writeFooter 写入文件结束标记和校验和
func (e *Encoder) writeFooter() error {
	// 写入结束标记
	if _, err := e.bufWriter.Write([]byte{rdbOpcodeEOF}); err != nil {
		return err
	}
	
	// 计算并写入校验和（这里简化处理，实际应该计算整个文件的 CRC64）
	// 在实际实现中，我们需要在写入过程中计算 CRC
	// 这里先写入 8 字节的 0 作为占位
	crc := make([]byte, 8)
	_, err := e.bufWriter.Write(crc)
	return err
}

// encodeValue 编码值（用于不同类型的数据结构）
func (e *Encoder) encodeValue(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		return e.writeString(string(v))
	case string:
		return e.writeString(v)
	case int:
		return e.writeInteger(int64(v))
	case int64:
		return e.writeInteger(v)
	default:
		// 尝试转换为字符串
		return e.writeString(fmt.Sprintf("%v", v))
	}
}

// SaveToFile 将数据库保存到 RDB 文件
func SaveToFile(filename string, db database.Database) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := NewEncoder(file, db)
	return encoder.Encode()
}
