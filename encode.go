package bsony

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
)

func hasEnoughBytes(b []byte, offset int, n int) error {
	if len(b)-offset < n {
		return errShortDoc
	}
	return nil
}

func writeTypeAndKey(dst []byte, offset int, t Type, key string) int {
	dst[offset] = byte(t)
	nullByteOffset := offset + 1 + len(key)
	copy(dst[offset+1:nullByteOffset], []byte(key))
	dst[nullByteOffset] = 0
	// return offset for next thing to write
	return nullByteOffset + 1
}

func readTypeAndKey(src []byte, offset int) (Type, string, error) {
	if err := hasEnoughBytes(src, offset, 2); err != nil {
		return 0, "", err
	}
	t := src[offset]
	nullByteOffset := bytes.IndexByte(src[offset+1:], 0)
	if nullByteOffset == -1 {
		return 0, "", errors.New("key not found")
	}
	key := string(src[offset+1 : offset+1+nullByteOffset])
	return Type(t), key, nil
}

func writeFloat64(dst []byte, offset int, value float64) int {
	bits := math.Float64bits(value)
	binary.LittleEndian.PutUint64(dst[offset:offset+8], bits)
	// return offset for next thing to write
	return offset + 8
}

func readFloat64(src []byte, offset int) (float64, error) {
	if err := hasEnoughBytes(src, offset, 8); err != nil {
		return 0, err
	}
	bits := binary.LittleEndian.Uint64(src[offset : offset+8])
	return math.Float64frombits(bits), nil
}

func writeInt32(dst []byte, offset int, value int32) int {
	binary.LittleEndian.PutUint32(dst[offset:offset+4], uint32(value))
	return offset + 4
}

func writeUint32(dst []byte, offset int, value uint32) int {
	binary.LittleEndian.PutUint32(dst[offset:offset+4], value)
	return offset + 4
}

func readInt32(src []byte, offset int) (int32, error) {
	if err := hasEnoughBytes(src, offset, 4); err != nil {
		return 0, err
	}
	bits := binary.LittleEndian.Uint32(src[offset : offset+4])
	return int32(bits), nil
}

func writeInt64(dst []byte, offset int, value int64) int {
	binary.LittleEndian.PutUint64(dst[offset:offset+8], uint64(value))
	return offset + 8
}

func writeUint64(dst []byte, offset int, value uint64) int {
	binary.LittleEndian.PutUint64(dst[offset:offset+8], value)
	return offset + 8
}

func readInt64(src []byte, offset int) (int64, error) {
	if err := hasEnoughBytes(src, offset, 8); err != nil {
		return 0, err
	}
	bits := binary.LittleEndian.Uint64(src[offset : offset+8])
	return int64(bits), nil
}

func writeString(dst []byte, offset int, value string) int {
	strlen := len(value)
	// Length is string length plus 1 for null byte
	offset = writeInt32(dst, offset, int32(strlen)+1)
	copy(dst[offset:offset+strlen], []byte(value))
	dst[offset+strlen] = 0
	return offset + strlen + 1
}

func writeCString(dst []byte, offset int, value string) int {
	strlen := len(value)
	copy(dst[offset:offset+strlen], []byte(value))
	dst[offset+strlen] = 0
	return offset + strlen + 1
}

func readCString(src []byte, offset int) (string, error) {
	if err := hasEnoughBytes(src, offset, 1); err != nil {
		return "", err
	}
	nullPos := bytes.IndexByte(src[offset:], 0)
	if nullPos == -1 {
		return "", errors.New("cstring null terminator not found")
	}
	return string(src[offset : offset+nullPos]), nil
}
