package rdb

import (
	"encoding/binary"
	"hash/crc64"
	"io"
)

const (
	// RDB 版本号
	rdbVersion = "0006"

	// RDB 魔数
	rdbMagic = "REDIS"

	// 操作码
	rdbOpcodeAuxField    = 0xFA // 辅助字段
	rdbOpcodeResizeDB    = 0xFB // 数据库大小调整
	rdbOpcodeExpireTime  = 0xFD // 过期时间（秒）
	rdbOpcodeExpireTimeMS = 0xFC // 过期时间（毫秒）
	rdbOpcodeSelectDB    = 0xFE // 选择数据库
	rdbOpcodeEOF         = 0xFF // 结束标记

	// 字符串编码类型
	rdbTypeString = 0 // 字符串

	// 长度编码前缀
	rdbLenEncoding6Bit  = 0x00 // 00xxxxxx - 6位长度
	rdbLenEncoding14Bit = 0x40 // 01xxxxxx - 14位长度
	rdbLenEncoding32Bit = 0x80 // 10xxxxxx - 32位长度
	rdbLenEncodingEnc   = 0xC0 // 11xxxxxx - 特殊编码

	// 整数编码
	rdbEncInt8  = 0 // 8位整数
	rdbEncInt16 = 1 // 16位整数
	rdbEncInt32 = 2 // 32位整数
	rdbEncLZF   = 3 // LZF 压缩
)

// crc64Table 用于 RDB 文件校验
type crc64Table struct {
	table [256]uint64
}

// makeCRC64Table 创建 CRC64 查找表
func makeCRC64Table() *crc64Table {
	t := &crc64Table{}
	for i := 0; i < 256; i++ {
		crc := uint64(i)
		for j := 0; j < 8; j++ {
			if crc&1 == 1 {
				crc = (crc >> 1) ^ 0xad93d23594c935a9 // CRC64 多项式
			} else {
				crc >>= 1
			}
		}
		t.table[i] = crc
	}
	return t
}

// crc64TableInstance 全局 CRC64 查找表
var crc64TableInstance = makeCRC64Table()

// calcCRC64 计算 CRC64 校验和
func calcCRC64(data []byte) uint64 {
	crc := uint64(0)
	for _, b := range data {
		crc = crc64TableInstance.table[(crc^uint64(b))&0xff] ^ (crc >> 8)
	}
	return crc
}

// writeCRC64 写入 CRC64 校验和到 writer
func writeCRC64(w io.Writer, crc uint64) error {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, crc)
	_, err := w.Write(buf)
	return err
}

// readCRC64 从 reader 读取 CRC64 校验和
func readCRC64(r io.Reader) (uint64, error) {
	buf := make([]byte, 8)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(buf), nil
}
