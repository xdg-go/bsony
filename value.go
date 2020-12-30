package bsony

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var errAlreadyReleased = errors.New("value released")

type Value interface {
	Clone() Value
	CopyTo(dst []byte) int
	Err() error
	Get() interface{}
	Len() int
	Release()
	Type() Type
}

// A unsafeValue is an immutable view into a buffer.  It must
// not be used past the lifetime of that container.
type unsafeValue struct {
	factory *Factory
	t       Type
	data    []byte
	err     error
}

// unsafe value is not owned -- it's just a view into another buffer.
// it constructs a view of the given type, given that `src` is the beginning of
// the data (i.e. after the key of a document/array)
//
// XXX should we have a sync.Pool for Values?
func newValueUnsafe(f *Factory, src []byte, t Type) *unsafeValue {
	v := &unsafeValue{factory: f, t: t}
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
		docLen, _ := readInt32(src, 8+int(strLen))
		if length != 8+strLen+docLen {
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

// Clone returns a copy of an value, including copying the underlying data
// buffer.
func (v *unsafeValue) Clone() Value {
	return newOwnedValue(v.factory, v)
}

func (v *unsafeValue) CopyTo(dst []byte) int {
	return copy(dst, v.data)
}

// Err ...
func (v *unsafeValue) Err() error {
	return v.err
}

func (v *unsafeValue) Len() int {
	return len(v.data)
}

// Release ...
func (v *unsafeValue) Release() {
	v.factory = nil
	v.data = nil
	v.t = TypeInvalid
	v.err = errAlreadyReleased
}

// Type ...
func (v *unsafeValue) Type() Type {
	return v.t
}

// Get returns a decoded copy of the value or nil if the value could not be
// decoded.  It is safe to keep the result of a Get and release the source
// document.
func (v *unsafeValue) Get() interface{} {
	if v.err != nil {
		return nil
	}

	// Can ignore errors on simple type reads because we know we have
	// enough bytes from the constructor.
	switch v.t {
	case TypeNull:
		return nil

	case TypeUndefined:
		return primitive.Undefined{}

	case TypeMinKey:
		return primitive.MinKey{}

	case TypeMaxKey:
		return primitive.MaxKey{}

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
		return primitive.Timestamp{T: sec, I: inc}

	case TypeObjectID:
		x := primitive.ObjectID{}
		copy(x[0:12], v.data)
		return x

	case TypeDecimal128:
		l := binary.LittleEndian.Uint64(v.data[0:8])
		h := binary.LittleEndian.Uint64(v.data[8:16])
		return primitive.NewDecimal128(h, l)

	case TypeString:
		// Skip length and omit trailing null byte.
		return string(v.data[4 : len(v.data)-1])

	case TypeSymbol:
		// Skip length and omit trailing null byte.
		return primitive.Symbol(v.data[4 : len(v.data)-1])

	case TypeJavaScript:
		// Skip length and omit trailing null byte.
		return primitive.JavaScript(v.data[4 : len(v.data)-1])

	case TypeEmbeddedDocument:
		// XXX validate?
		src := &Doc{factory: v.factory, buf: v.data, valid: true, immutable: true}
		return src.Clone()

	case TypeArray:
		// XXX validate?
		src := &Array{d: &Doc{factory: v.factory, buf: v.data, valid: true, immutable: true}}
		return src.Clone()

	case TypeCodeWithScope:
		// Skip total CWS length to get just string length; omit trailing null
		data := v.data[4:]
		strLen, _ := readInt32(data, 0)
		code := string(data[4 : 4+strLen-1])
		src := &Doc{factory: v.factory, buf: data[4+strLen:], valid: true, immutable: true}
		scope := src.Clone()
		return CodeWithScope{Code: code, Scope: scope}

	case TypeBinary:
		// Skip the length to find the subtype byte
		x := primitive.Binary{Subtype: v.data[4]}
		payload := v.data[5:]
		// Legacy subtype 2 has another length after subtype byte
		if x.Subtype == 2 {
			payload = payload[4:]
		}
		x.Data = make([]byte, len(payload))
		copy(x.Data, payload)
		return x

	case TypeRegex:
		pattern, _ := readCString(v.data, 0)
		// Skip first string and trailing null byte
		options, _ := readCString(v.data, len(pattern)+1)
		return primitive.Regex{Pattern: pattern, Options: options}

	case TypeDBPointer:
		strLen, _ := readInt32(v.data, 0)
		ref := string(v.data[4 : 4+strLen-1])
		id := primitive.ObjectID{}
		copy(id[0:12], v.data[strLen+4:])
		return primitive.DBPointer{DB: ref, Pointer: id}
	}

	return nil
}

// XXX Have methods that return typed values?
// func (...) Int32OK() (int32, bool)

// An ownedValue contains a complete copy of its data and releases
// it to the pool when the value is released.
type ownedValue struct {
	unsafeValue
	buf []byte
}

// newOwnedValue creates an copy of an value.
func newOwnedValue(f *Factory, e Value) *ownedValue {
	var buf []byte
	if e.Type() != TypeInvalid {
		buf = f.pool.Resize(f.pool.Get(), e.Len())
		e.CopyTo(buf)
	}
	return &ownedValue{
		unsafeValue: unsafeValue{factory: f, data: buf, t: e.Type(), err: e.Err()},
		buf:         buf,
	}
}

// Release returns the owned buffer to the pool.
func (o *ownedValue) Release() {
	o.factory.release(o.buf)
	o.buf = nil
	o.unsafeValue.Release()
}
