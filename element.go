package bsony

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

type Element interface {
	Release()
	Type() Type
	Data() []byte
	Err() error
	Clone() *Element
	Value() interface{}
}

// A ownedElement ...
type ownedElement struct {
	pool *Factory
	t    Type
	data []byte
	owns bool // true if this Element is responsible for putting back to the pool
	err  error
}

// unsafe element is not owned -- it's just a view into another buffer.
// it constructs a view of the given type, given that `src` is the beginning of
// the data (i.e. after the key of a document/array)
//
// XXX should we have a sync.Pool for Elements?
func newElementUnsafe(bx *Factory, src []byte, t Type) *ownedElement {
	v := &ownedElement{pool: bx, t: t}
	var err error
	switch t {
	case TypeNull, TypeUndefined, TypeMinKey, TypeMaxKey:
		v.data = src[0:0]

	case TypeBoolean:
		if err = hasEnoughBytes(src, 0, 1); err != nil {
			v.err = err
			return v
		}
		v.data = src[0:1]

	case TypeInt32:
		if err = hasEnoughBytes(src, 0, 4); err != nil {
			v.err = err
			return v
		}
		v.data = src[0:4]

	case TypeDouble, TypeInt64, TypeDateTime, TypeTimestamp:
		if err = hasEnoughBytes(src, 0, 8); err != nil {
			v.err = err
			return v
		}
		v.data = src[0:8]

	case TypeObjectID:
		if err = hasEnoughBytes(src, 0, 12); err != nil {
			v.err = err
			return v
		}
		v.data = src[0:12]

	case TypeDecimal128:
		if err = hasEnoughBytes(src, 0, 16); err != nil {
			v.err = err
			return v
		}
		v.data = src[0:16]

	case TypeString, TypeSymbol, TypeJavaScript:
		// Minimum bytes:  length + null == 5
		if err = hasEnoughBytes(src, 0, 5); err != nil {
			v.err = err
			return v
		}
		length, _ := readInt32(src, 0)
		// For these types, encoded length does not include itself
		length = length + 4
		if err = hasEnoughBytes(src, 0, int(length)); err != nil {
			v.err = err
			return v
		}
		v.data = src[0:length]

	case TypeEmbeddedDocument, TypeArray:
		// Minimum bytes:  length + null == 5
		if err = hasEnoughBytes(src, 0, 5); err != nil {
			v.err = err
			return v
		}
		length, _ := readInt32(src, 0)
		// For these types, encoded length includes itself
		if err = hasEnoughBytes(src, 0, int(length)); err != nil {
			v.err = err
			return v
		}
		// Requires null terminator or these types are invalid
		if src[length-1] != 0 {
			v.err = fmt.Errorf("%s value missing null terminator", t)
			return v
		}
		v.data = src[0:length]

	case TypeCodeWithScope:
		// Minimum bytes:  length + length + null + length + null == 14
		if err = hasEnoughBytes(src, 0, 14); err != nil {
			v.err = err
			return v
		}
		length, _ := readInt32(src, 0)
		// For this type, encoded length includes itself
		if err = hasEnoughBytes(src, 0, int(length)); err != nil {
			v.err = err
			return v
		}
		// encoded string length must leave room for doc length + null
		strLen, _ := readInt32(src, 4)
		if length-strLen < 5 {
			v.err = fmt.Errorf("%s string length too long", t)
			return v
		}
		// encoded doc length must consume rest of the bytes
		docLen, _ := readInt32(src, 4+int(strLen))
		if length-strLen-docLen != 0 {
			v.err = fmt.Errorf("%s scope size invalid", t)
			return v
		}
		// Requires null terminator or this type are invalid
		if src[length-1] != 0 {
			v.err = fmt.Errorf("%s scope missing null terminator", t)
			return v
		}
		v.data = src[0:length]

	case TypeBinary:
		// Minimum bytes: length + subtype byte == 5
		if err = hasEnoughBytes(src, 0, 5); err != nil {
			v.err = err
			return v
		}
		length, _ := readInt32(src, 0)
		// For this type, encoded length does not includes itself or the
		// binary subtype byte
		length = length + 5
		if err = hasEnoughBytes(src, 0, int(length)); err != nil {
			v.err = err
			return v
		}
		// XXX Maybe check subtype 2 length is valid
		v.data = src[0:length]

	case TypeRegex:
		// Minimum bytes: 2 cstring null terminators
		if err = hasEnoughBytes(src, 0, 2); err != nil {
			v.err = err
			return v
		}
		first := bytes.IndexByte(src, 0)
		if first == -1 {
			v.err = errors.New("regex cstring unterminated")
			return v
		}
		second := bytes.IndexByte(src[first+1:], 0)
		if second == -1 {
			v.err = errors.New("regex cstring unterminated")
			return v
		}
		// Length is length of each part plus two null bytes; we know the
		// length is valid because we searched the original slice.
		v.data = src[0 : first+second+2]

	case TypeDBPointer:
		// Minimum bytes: length + null + 12-byte OID == 17
		if err = hasEnoughBytes(src, 0, 17); err != nil {
			v.err = err
			return v
		}
		length, _ := readInt32(src, 0)
		// For this type, encoded length does not include itself or
		// trailing 12 bytes of OID
		length = length + 16
		if err = hasEnoughBytes(src, 0, int(length)); err != nil {
			v.err = err
			return v
		}
		v.data = src[0:length]
	}
	return v
}

// Release ...
func (v *ownedElement) Release() {
	v.pool = nil
	// XXX set error for "already released"?
	if v.owns {
		// XXX probably should be `pool.put` or something
		v.pool.pool.Put(v.data)
	}
	// XXX clear data and type?
}

// Type ...
func (v *ownedElement) Type() Type {
	return v.t
}

// Data ...
func (v *ownedElement) Data() []byte {
	return v.data
}

// Err ...
func (v *ownedElement) Err() error {
	return v.err
}

// Clone ...
// copies an element, taking ownership of the buffer
func (v *ownedElement) Clone() *ownedElement {
	buf := v.pool.pool.Resize(v.pool.pool.Get(), len(v.data))
	copy(buf, v.data)
	return &ownedElement{pool: v.pool, t: v.t, data: buf, err: v.err}
}

// Value returns ... (decoded copy of data)
func (v *ownedElement) Value() interface{} {
	if v.err != nil {
		return nil
	}

	// XXX big type switch, calling out to decoding routines
	// Can ignore errors on simple type reads because we know we have
	// enough bytes.
	switch v.t {
	case TypeNull, TypeUndefined:
		return nil

	case TypeMinKey:
		return MinKey{}

	case TypeMaxKey:
		return MaxKey{}

	case TypeBoolean:
		return v.data[0] != 0

	case TypeInt32:
		x, _ := readInt32(v.data, 0)
		return x

	case TypeDouble:
		x, _ := readFloat64(v.data, 0)
		return x

	case TypeInt64:
		x, _ := readInt64(v.data, 0)
		return x

	case TypeDateTime:
		x, _ := readInt64(v.data, 0)
		sec := x / 1000
		ns := (x % 1000) * 1000000
		return time.Unix(sec, ns)

	case TypeTimestamp:
		inc := binary.LittleEndian.Uint32(v.data[0:4])
		sec := binary.LittleEndian.Uint32(v.data[4:8])
		return Timestamp{Seconds: sec, Increment: inc}

	case TypeObjectID:
		x := OID{}
		copy(x[0:12], v.data)
		return x

	case TypeDecimal128:
		l := binary.LittleEndian.Uint64(v.data[0:8])
		h := binary.LittleEndian.Uint64(v.data[8:16])
		return Decimal128{H: h, L: l}

	case TypeString, TypeSymbol:
		// Skip length and omit trailing null byte.
		return string(v.data[4 : len(v.data)-1])

	case TypeJavaScript:
		// Skip length and omit trailing null byte.
		return Code{Code: string(v.data[4:])}

	case TypeEmbeddedDocument:
		src := &Doc{factory: v.pool, buf: v.data, valid: true, immutable: true}
		return src.Clone()

	case TypeArray:
		src := &Array{d: &Doc{factory: v.pool, buf: v.data, valid: true, immutable: true}}
		return src.Clone()

	case TypeCodeWithScope:
		// Skip length and omit trailing null byte.
		strLen, _ := readInt32(v.data, 4)
		code := string(v.data[4 : strLen-1])
		// XXX This is wrong!
		scope := &Doc{}
		return Code{Code: code, Scope: scope}

	case TypeBinary:
		// Skip the length to find the subtype byte
		x := Binary{Subtype: v.data[4]}
		payload := v.data[5:]
		// Legacy subtype 2 has another length after subtype byte
		if x.Subtype == 2 {
			payload = payload[4:]
		}
		x.Payload = make([]byte, len(payload))
		copy(x.Payload, payload)
		return x

	case TypeRegex:
		pattern, _ := readCString(v.data, 0)
		// Skip first string and trailing null byte
		flags, _ := readCString(v.data, len(pattern)+1)
		return Regex{Pattern: pattern, Flags: flags}

	case TypeDBPointer:
		strLen, _ := readInt32(v.data, 0)
		ref := string(v.data[4 : 4+strLen])
		id := OID{}
		copy(id[0:12], v.data[strLen+5:])
		return DBPointer{Ref: ref, ID: id}
	}

	return nil
}

// XXX Have methods that return typed values?
// func (...) Int32OK() (int32, bool)
