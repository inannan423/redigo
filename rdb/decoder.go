package rdb

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"redigo/interface/database"
	"redigo/lib/logger"
	"redigo/resp/connection"
	"redigo/resp/reply"
	"strconv"
	"time"
)

// Decoder 用于从 RDB 文件解码数据
type Decoder struct {
	reader    io.Reader
	bufReader *bufio.Reader
	db        database.Database
}

// NewDecoder 创建一个新的 RDB 解码器
func NewDecoder(r io.Reader, db database.Database) *Decoder {
	bufReader := bufio.NewReader(r)
	return &Decoder{
		reader:    r,
		bufReader: bufReader,
		db:        db,
	}
}

// Decode 解码 RDB 文件并加载到数据库
func (d *Decoder) Decode() error {
	// 1. 读取魔数和版本号
	if err := d.readMagic(); err != nil {
		return err
	}

	// 2. 循环读取操作码和数据
	for {
		opcode, err := d.bufReader.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		switch opcode {
		case rdbOpcodeAuxField:
			if err := d.readAuxField(); err != nil {
				return err
			}
		case rdbOpcodeSelectDB:
			if err := d.readSelectDB(); err != nil {
				return err
			}
		case rdbOpcodeResizeDB:
			if err := d.readResizeDB(); err != nil {
				return err
			}
		case rdbOpcodeExpireTime:
			if err := d.readExpireTime(false); err != nil {
				return err
			}
		case rdbOpcodeExpireTimeMS:
			if err := d.readExpireTime(true); err != nil {
				return err
			}
		case rdbOpcodeEOF:
			// 读取校验和
			if err := d.readChecksum(); err != nil {
				return err
			}
			return nil
		default:
			// 处理数据类型
			if err := d.readKeyValue(opcode); err != nil {
				return err
			}
		}
	}

	return nil
}

// readMagic 读取 RDB 魔数和版本号
func (d *Decoder) readMagic() error {
	// 读取魔数 "REDIS"
	magic := make([]byte, 5)
	if _, err := io.ReadFull(d.bufReader, magic); err != nil {
		return err
	}
	if string(magic) != rdbMagic {
		return fmt.Errorf("invalid RDB magic: %s", string(magic))
	}

	// 读取版本号
	version := make([]byte, 4)
	if _, err := io.ReadFull(d.bufReader, version); err != nil {
		return err
	}
	
	logger.Info("RDB version: " + string(version))
	return nil
}

// readAuxField 读取辅助字段
func (d *Decoder) readAuxField() error {
	key, err := d.readString()
	if err != nil {
		return err
	}
	val, err := d.readString()
	if err != nil {
		return err
	}
	
	logger.Info(fmt.Sprintf("RDB aux field: %s = %s", key, val))
	return nil
}

// readSelectDB 读取数据库选择标记
func (d *Decoder) readSelectDB() error {
	dbIndex, err := d.readLength()
	if err != nil {
		return err
	}
	
	// 切换到指定数据库
	fakeConn := &connection.Connection{}
	selectCmd := [][]byte{[]byte("SELECT"), []byte(strconv.Itoa(int(dbIndex)))}
	result := d.db.Exec(fakeConn, selectCmd)
	
	if reply.IsErrReply(result) {
		return fmt.Errorf("failed to select db %d", dbIndex)
	}
	
	return nil
}

// readResizeDB 读取数据库大小调整标记
func (d *Decoder) readResizeDB() error {
	dbSize, err := d.readLength()
	if err != nil {
		return err
	}
	expireSize, err := d.readLength()
	if err != nil {
		return err
	}
	
	logger.Info(fmt.Sprintf("RDB resize db: size=%d, expire=%d", dbSize, expireSize))
	return nil
}

// readExpireTime 读取过期时间
func (d *Decoder) readExpireTime(ms bool) error {
	var expireTime uint64
	var err error
	
	if ms {
		// 毫秒级时间戳
		buf := make([]byte, 8)
		_, err = io.ReadFull(d.bufReader, buf)
		if err != nil {
			return err
		}
		expireTime = binary.LittleEndian.Uint64(buf)
	} else {
		// 秒级时间戳
		buf := make([]byte, 4)
		_, err = io.ReadFull(d.bufReader, buf)
		if err != nil {
			return err
		}
		expireTime = uint64(binary.LittleEndian.Uint32(buf)) * 1000 // 转换为毫秒
	}
	
	// 读取键值对类型
	valType, err := d.bufReader.ReadByte()
	if err != nil {
		return err
	}
	
	// 读取键
	key, err := d.readString()
	if err != nil {
		return err
	}
	
	// 读取值
	val, err := d.readValue(valType)
	if err != nil {
		return err
	}
	
	// 执行 SET 命令设置键值对
	fakeConn := &connection.Connection{}
	setCmd := [][]byte{[]byte("SET"), []byte(key), val}
	d.db.Exec(fakeConn, setCmd)
	
	// 设置过期时间（简化实现，实际应该使用 PEXPIREAT 命令）
	if expireTime > 0 {
		expireCmd := [][]byte{[]byte("PEXPIREAT"), []byte(key), []byte(strconv.FormatInt(int64(expireTime), 10))}
		d.db.Exec(fakeConn, expireCmd)
	}
	
	return nil
}

// readKeyValue 读取键值对
func (d *Decoder) readKeyValue(valType byte) error {
	// 读取键
	key, err := d.readString()
	if err != nil {
		return err
	}
	
	// 读取值
	val, err := d.readValue(valType)
	if err != nil {
		return err
	}
	
	// 执行 SET 命令设置键值对
	fakeConn := &connection.Connection{}
	setCmd := [][]byte{[]byte("SET"), []byte(key), val}
	d.db.Exec(fakeConn, setCmd)
	
	return nil
}

// readValue 根据类型读取值
func (d *Decoder) readValue(valType byte) ([]byte, error) {
	switch valType {
	case rdbTypeString:
		return d.readStringBytes()
	default:
		return nil, fmt.Errorf("unsupported value type: %d", valType)
	}
}

// readString 读取字符串
func (d *Decoder) readString() (string, error) {
	buf, err := d.readStringBytes()
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

// readStringBytes 读取字符串字节
func (d *Decoder) readStringBytes() ([]byte, error) {
	// 读取长度
	length, special, err := d.readLengthWithEncoding()
	if err != nil {
		return nil, err
	}
	
	// 处理特殊编码
	if special {
		switch length {
		case rdbEncInt8:
			// 8位整数
			buf := make([]byte, 1)
			if _, err := io.ReadFull(d.bufReader, buf); err != nil {
				return nil, err
			}
			val := int8(buf[0])
			return []byte(strconv.FormatInt(int64(val), 10)), nil
			
		case rdbEncInt16:
			// 16位整数
			buf := make([]byte, 2)
			if _, err := io.ReadFull(d.bufReader, buf); err != nil {
				return nil, err
			}
			val := int16(binary.LittleEndian.Uint16(buf))
			return []byte(strconv.FormatInt(int64(val), 10)), nil
			
		case rdbEncInt32:
			// 32位整数
			buf := make([]byte, 4)
			if _, err := io.ReadFull(d.bufReader, buf); err != nil {
				return nil, err
			}
			val := int32(binary.LittleEndian.Uint32(buf))
			return []byte(strconv.FormatInt(int64(val), 10)), nil
			
		case rdbEncLZF:
			// LZF 压缩（暂不实现）
			return nil, errors.New("LZF compression not supported")
			
		default:
			return nil, fmt.Errorf("unknown special encoding: %d", length)
		}
	}
	
	// 读取普通字符串
	buf := make([]byte, length)
	if _, err := io.ReadFull(d.bufReader, buf); err != nil {
		return nil, err
	}
	
	return buf, nil
}

// readLength 读取长度
func (d *Decoder) readLength() (uint32, error) {
	length, _, err := d.readLengthWithEncoding()
	return length, err
}

// readLengthWithEncoding 读取长度，返回长度值、是否为特殊编码、错误
func (d *Decoder) readLengthWithEncoding() (uint32, bool, error) {
	// 读取第一个字节
	b, err := d.bufReader.ReadByte()
	if err != nil {
		return 0, false, err
	}
	
	// 检查编码类型
	typeBits := (b & 0xC0) >> 6
	
	switch typeBits {
	case 0:
		// 00xxxxxx - 6位长度
		return uint32(b & 0x3F), false, nil
		
	case 1:
		// 01xxxxxx xxxxxxxx - 14位长度
		b2, err := d.bufReader.ReadByte()
		if err != nil {
			return 0, false, err
		}
		return uint32(((uint16(b) & 0x3F) << 8) | uint16(b2)), false, nil
		
	case 2:
		// 10xxxxxx - 32位长度
		buf := make([]byte, 4)
		if _, err := io.ReadFull(d.bufReader, buf); err != nil {
			return 0, false, err
		}
		return binary.BigEndian.Uint32(buf), false, nil
		
	case 3:
		// 11xxxxxx - 特殊编码
		specialType := b & 0x3F
		return uint32(specialType), true, nil
	}
	
	return 0, false, errors.New("unknown length encoding")
}

// readChecksum 读取校验和
func (d *Decoder) readChecksum() error {
	// 读取 8 字节校验和
	checksum := make([]byte, 8)
	if _, err := io.ReadFull(d.bufReader, checksum); err != nil {
		return err
	}
	
	// 这里可以验证校验和
	// 简化实现，暂不验证
	
	return nil
}

// LoadFromFile 从 RDB 文件加载数据到数据库
func LoadFromFile(filename string, db database.Database) error {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，不是错误
			return nil
		}
		return err
	}
	defer file.Close()

	decoder := NewDecoder(file, db)
	return decoder.Decode()
}
