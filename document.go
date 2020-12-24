// Copyright 2018 by David A. Golden. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package bsony

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var errShortDoc = errors.New("not enough bytes available to read value")
var errInvalidLength = errors.New("document length doesn't match buffer length")
var errMissingTerminator = errors.New("document buffer missing null terminator")
var errImmutableInvalid = errors.New("can't modify immutable or invalid document")
var errBufferReleased = errors.New("buffer released")

// XXX Wonder if A and D should implement a common interface?

// A Doc object represents a BSON document
type Doc struct {
	factory   *Factory
	buf       []byte
	valid     bool
	immutable bool
	err       error
}

// check length and null termination
func validateBSONFraming(buf []byte) error {
	length, err := readInt32(buf, 0)
	if err != nil {
		return err
	}
	if len(buf) != int(length) {
		return errInvalidLength
	}
	if buf[len(buf)-1] != 0 {
		return errMissingTerminator
	}
	return nil
}

// Valid indicates if the document is valid for use.  A document is invalid
// after its storage has been released.
func (d *Doc) Valid() bool {
	return d.valid
}

// Err returns any error recorded on the document.
func (d *Doc) Err() error {
	return d.err
}

// Release returns allocated space.  After calling this method, the document is
// invalid.  Any prior error is replaced with a "buffer released" message.
func (d *Doc) Release() {
	if d.immutable {
		return
	}
	d.factory.release(d.buf)
	d.buf = nil
	d.factory = nil
	d.valid = false
	d.err = errBufferReleased
	return
}

// Iter ...
func (d *Doc) Iter() *DocIter {
	return newDocIter(d)
}

// Reader ...
func (d *Doc) Reader() (io.Reader, error) {
	// XXX unimplemented
	return nil, nil
}

// CopyTo ...
func (d *Doc) CopyTo(dst []byte) int {
	return copy(dst, d.buf)
}

// Map ...
//
// XXX Valid only until doc is released?
func (d *Doc) Map() map[string]interface{} {
	m := make(map[string]interface{})
	iter := d.Iter()
	for iter.Type() != TypeInvalid {
		m[iter.Key()] = iter.Value()
		iter.Next()
	}
	return m
}

// XXX should there be an "OMap" method?

// Len ...
func (d *Doc) Len() int {
	return len(d.buf)
}

// grow increases the buffer size by the given amount and sets length bytes at
// start of the document to match.
func (d *Doc) grow(n int) {
	newlen := len(d.buf) + n
	d.buf = d.factory.resize(d.buf, newlen)
	binary.LittleEndian.PutUint32(d.buf[0:4], uint32(newlen))
}

// Clone ...
func (d *Doc) Clone() *Doc {
	return d.factory.NewDoc().Concat(d)
}

// Concat ..
func (d *Doc) Concat(src *Doc) *Doc {
	// Grow by len(src) less length bytes and null terminator byte
	offset := len(d.buf) - 1
	d.grow(len(src.buf) - 5)
	copy(d.buf[offset:], src.buf[4:])
	return d
}

// Add ...
//
// XXX rethink which of these actually need pointer support? all or none?
func (d *Doc) Add(k string, v interface{}) *Doc {
	if d.immutable || !d.valid {
		d.err = errImmutableInvalid
		return d
	}

	switch x := v.(type) {
	// Type 01 - double
	case float32:
		return d.AddDouble(k, float64(x))
	case float64:
		return d.AddDouble(k, x)

	// Type 02 - string
	case string:
		return d.AddString(k, x)

	// Type 03 - document
	case Doc:
		return d.AddDoc(k, &x)
	case *Doc:
		return d.AddDoc(k, x)

	// Type 04 - array
	case Array:
		return d.AddArray(k, &x)
	case *Array:
		return d.AddArray(k, x)

	// Type 05 - binary
	case primitive.Binary:
		return d.AddBinary(k, &x)
	case *primitive.Binary:
		return d.AddBinary(k, x)

	// Type 06 - undefined (deprecated)
	case primitive.Undefined:
		return d.AddUndefined(k)

	// Type 07 - ObjectID
	case primitive.ObjectID:
		return d.AddOID(k, x)

	// Type 08 - boolean
	case bool:
		return d.AddBool(k, x)

	// Type 09 - UTC DateTime
	case primitive.DateTime:
		return d.AddDateTime(k, x)
	case time.Time:
		return d.AddDateTimeFromTime(k, x)
	case *time.Time:
		return d.AddDateTimeFromTime(k, *x)

	// Type 0A - null
	case nil:
		return d.AddNull(k)

	// Type 0B - regular expression
	case primitive.Regex:
		return d.AddRegex(k, x)
	case *primitive.Regex:
		return d.AddRegex(k, *x)

	// Type 0C - DBPointer (deprecated)
	case primitive.DBPointer:
		return d.AddDBPointer(k, x)
	case *primitive.DBPointer:
		return d.AddDBPointer(k, *x)

	// Type 0D - JavaScript code
	case primitive.JavaScript:
		return d.AddJavaScript(k, x)

	// Type 0E - Symbol (deprecated)
	case primitive.Symbol:
		return d.AddSymbol(k, x)

		// Type 0F - JavaScript code with scope
	case CodeWithScope:
		return d.AddCodeScope(k, x)
	case primitive.CodeWithScope:
		buf, err := bson.Marshal(x.Scope)
		if err != nil {
			d.err = fmt.Errorf("error marshaling scope for key '%s': %w", k, err)
			return d
		}
		scope, err := d.factory.NewDocFromBytes(buf)
		if err != nil {
			d.err = fmt.Errorf("error marshaling scope for key '%s': buffer invalid: %w", k, err)
			return d
		}
		return d.AddCodeScope(k, CodeWithScope{Code: string(x.Code), Scope: scope})

	// Type 10 - 32-bit integer
	case int32:
		return d.AddInt32(k, x)

		// Type 11 - timestamp
	case primitive.Timestamp:
		return d.AddTimestamp(k, x)

	// Type 12 - 64-bit integer
	case int64:
		return d.AddInt64(k, x)

		// Type 13 - 128-bit decimal floating point
	case primitive.Decimal128:
		return d.AddDecimal128(k, x)
	case *primitive.Decimal128:
		return d.AddDecimal128(k, *x)

	// Type FF - Min key
	case primitive.MinKey:
		return d.AddMinKey(k)

	// Type 7F - Max key
	case primitive.MaxKey:
		return d.AddMaxKey(k)

	default:
		panic(fmt.Sprintf("unsupported type: %T", v))
	}
}

// AddDouble ...
func (d *Doc) AddDouble(k string, v float64) *Doc {
	if d.immutable || !d.valid {
		d.err = errImmutableInvalid
		return d
	}
	offset := len(d.buf) - 1
	// Add space for type byte + len(key) + null byte + 8 for float64
	d.grow(10 + len(k))
	offset = writeTypeAndKey(d.buf, offset, TypeDouble, k)
	writeFloat64(d.buf, offset, v)
	d.buf[len(d.buf)-1] = 0
	return d
}

// AddString ...
func (d *Doc) AddString(k string, v string) *Doc {
	if d.immutable || !d.valid {
		d.err = errImmutableInvalid
		return d
	}
	offset := len(d.buf) - 1
	// Add space for type byte + len(key) + null byte + length + string + null byte
	d.grow(7 + len(k) + len(v))
	offset = writeTypeAndKey(d.buf, offset, TypeString, k)
	writeString(d.buf, offset, v)
	d.buf[len(d.buf)-1] = 0
	return d
}

// AddDoc ...
func (d *Doc) AddDoc(k string, v *Doc) *Doc {
	if d.immutable || !d.valid {
		d.err = errImmutableInvalid
		return d
	}
	offset := len(d.buf) - 1
	// Add space for type byte + len(key) + null byte + sub document length
	d.grow(len(k) + v.Len() + 2)
	offset = writeTypeAndKey(d.buf, offset, TypeEmbeddedDocument, k)
	copy(d.buf[offset:], v.buf)
	d.buf[len(d.buf)-1] = 0
	return d
}

// AddArray ...
func (d *Doc) AddArray(k string, v *Array) *Doc {
	if d.immutable || !d.valid {
		d.err = errImmutableInvalid
		return d
	}
	offset := len(d.buf) - 1
	// Add space for type byte + len(key) + null byte + array length
	d.grow(len(k) + v.Len() + 2)
	offset = writeTypeAndKey(d.buf, offset, TypeArray, k)
	copy(d.buf[offset:], v.d.buf)
	d.buf[len(d.buf)-1] = 0
	return d
}

// AddBinary ...
func (d *Doc) AddBinary(k string, v *primitive.Binary) *Doc {
	if d.immutable || !d.valid {
		d.err = errImmutableInvalid
		return d
	}
	offset := len(d.buf) - 1
	// Add space for type byte + len(key) + null byte + length + subtype byte +
	// payload; payload for subtype 2 also adds payload length bytes
	dataSize := len(v.Data)
	if v.Subtype == 2 {
		dataSize += 4
	}
	d.grow(7 + len(k) + dataSize)
	offset = writeTypeAndKey(d.buf, offset, TypeBinary, k)
	offset = writeInt32(d.buf, offset, int32(dataSize))
	d.buf[offset] = v.Subtype
	offset++
	if v.Subtype == 2 {
		offset = writeInt32(d.buf, offset, int32(len(v.Data)))
	}
	copy(d.buf[offset:], v.Data)
	d.buf[len(d.buf)-1] = 0
	return d
}

// AddUndefined ...
func (d *Doc) AddUndefined(k string) *Doc {
	if d.immutable || !d.valid {
		d.err = errImmutableInvalid
		return d
	}
	offset := len(d.buf) - 1
	// Add space for type byte + len(key) + null byte; no data bytes
	d.grow(2 + len(k))
	offset = writeTypeAndKey(d.buf, offset, TypeUndefined, k)
	d.buf[len(d.buf)-1] = 0
	return d
}

// AddOID ...
func (d *Doc) AddOID(k string, v primitive.ObjectID) *Doc {
	if d.immutable || !d.valid {
		d.err = errImmutableInvalid
		return d
	}
	offset := len(d.buf) - 1
	// Add space for type byte + len(key) + null byte + OID length (12)
	d.grow(len(k) + 14)
	offset = writeTypeAndKey(d.buf, offset, TypeObjectID, k)
	copy(d.buf[offset:], v[:])
	d.buf[len(d.buf)-1] = 0
	return d
}

// AddBool ...
func (d *Doc) AddBool(k string, v bool) *Doc {
	if d.immutable || !d.valid {
		d.err = errImmutableInvalid
		return d
	}
	offset := len(d.buf) - 1
	// Add space for type byte + len(key) + null byte + boolean byte
	d.grow(len(k) + 3)
	offset = writeTypeAndKey(d.buf, offset, TypeBoolean, k)
	if v {
		d.buf[offset] = 1
	} else {
		d.buf[offset] = 0
	}
	d.buf[len(d.buf)-1] = 0
	return d
}

// AddDateTime ...
func (d *Doc) AddDateTime(k string, v primitive.DateTime) *Doc {
	if d.immutable || !d.valid {
		d.err = errImmutableInvalid
		return d
	}
	offset := len(d.buf) - 1
	// Add space for type byte + len(key) + null byte + 8 for int64
	d.grow(10 + len(k))
	offset = writeTypeAndKey(d.buf, offset, TypeDateTime, k)
	writeInt64(d.buf, offset, int64(v))
	d.buf[len(d.buf)-1] = 0
	return d
}

// AddDateTimeFromTime ...
func (d *Doc) AddDateTimeFromTime(k string, v time.Time) *Doc {
	return d.AddDateTime(k, primitive.NewDateTimeFromTime(v))
}

// AddNull ...
func (d *Doc) AddNull(k string) *Doc {
	if d.immutable || !d.valid {
		d.err = errImmutableInvalid
		return d
	}
	offset := len(d.buf) - 1
	// Add space for type byte + len(key) + null byte; no data bytes
	d.grow(2 + len(k))
	offset = writeTypeAndKey(d.buf, offset, TypeNull, k)
	d.buf[len(d.buf)-1] = 0
	return d
}

// AddRegex ...
func (d *Doc) AddRegex(k string, v primitive.Regex) *Doc {
	if d.immutable || !d.valid {
		d.err = errImmutableInvalid
		return d
	}
	offset := len(d.buf) - 1
	// Add space for type byte + len(key) + null byte + len(pattern) + null byte
	// + len(flags) + null byte
	d.grow(4 + len(k) + len(v.Pattern) + len(v.Options))
	offset = writeTypeAndKey(d.buf, offset, TypeRegex, k)
	offset = writeCString(d.buf, offset, v.Pattern)
	writeCString(d.buf, offset, v.Options)
	d.buf[len(d.buf)-1] = 0
	return d
}

// AddDBPointer ...
func (d *Doc) AddDBPointer(k string, v primitive.DBPointer) *Doc {
	if d.immutable || !d.valid {
		d.err = errImmutableInvalid
		return d
	}
	offset := len(d.buf) - 1
	// Add space for type byte + len(key) + null byte + length + string + null
	// byte + 12-byte ID
	d.grow(19 + len(k) + len(v.DB))
	offset = writeTypeAndKey(d.buf, offset, TypeDBPointer, k)
	offset = writeString(d.buf, offset, v.DB)
	copy(d.buf[offset:], v.Pointer[:])
	d.buf[len(d.buf)-1] = 0
	return d
}

// AddJavaScript ...
func (d *Doc) AddJavaScript(k string, v primitive.JavaScript) *Doc {
	if d.immutable || !d.valid {
		d.err = errImmutableInvalid
		return d
	}
	offset := len(d.buf) - 1
	// Add space for type byte + len(key) + null byte + length + string + null byte
	d.grow(7 + len(k) + len(v))
	offset = writeTypeAndKey(d.buf, offset, TypeJavaScript, k)
	writeString(d.buf, offset, string(v))
	d.buf[len(d.buf)-1] = 0
	return d
}

// AddSymbol ...
func (d *Doc) AddSymbol(k string, v primitive.Symbol) *Doc {
	if d.immutable || !d.valid {
		d.err = errImmutableInvalid
		return d
	}
	offset := len(d.buf) - 1
	// Add space for type byte + len(key) + null byte + string + null byte
	d.grow(7 + len(k) + len(v))
	offset = writeTypeAndKey(d.buf, offset, TypeSymbol, k)
	writeString(d.buf, offset, string(v))
	d.buf[len(d.buf)-1] = 0
	return d
}

// AddCodeScope ...
func (d *Doc) AddCodeScope(k string, v CodeWithScope) *Doc {
	if d.immutable || !d.valid {
		d.err = errImmutableInvalid
		return d
	}
	if v.Scope == nil {
		return d.AddJavaScript(k, primitive.JavaScript(v.Code))
	}
	offset := len(d.buf) - 1
	// CodeWithScope length is total length bytes + code length bytes + code
	// length + null byte + scope document length
	dataSize := 9 + len(v.Code) + v.Scope.Len()
	// Add space for type byte + len(key) + null byte + dataSize
	d.grow(2 + len(k) + dataSize)
	offset = writeTypeAndKey(d.buf, offset, TypeCodeWithScope, k)
	offset = writeInt32(d.buf, offset, int32(dataSize))
	offset = writeString(d.buf, offset, v.Code)
	copy(d.buf[offset:], v.Scope.buf)
	d.buf[len(d.buf)-1] = 0
	return d
}

// AddInt32 ...
func (d *Doc) AddInt32(k string, v int32) *Doc {
	if d.immutable || !d.valid {
		d.err = errImmutableInvalid
		return d
	}
	offset := len(d.buf) - 1
	// Add space for type byte + len(key) + null byte + 4 for int32
	d.grow(6 + len(k))
	offset = writeTypeAndKey(d.buf, offset, TypeInt32, k)
	writeInt32(d.buf, offset, v)
	d.buf[len(d.buf)-1] = 0
	return d
}

// AddTimestamp ...
func (d *Doc) AddTimestamp(k string, v primitive.Timestamp) *Doc {
	if d.immutable || !d.valid {
		d.err = errImmutableInvalid
		return d
	}
	offset := len(d.buf) - 1
	// Add space for type byte + len(key) + null byte + 8 typestamp bytes
	d.grow(10 + len(k))
	offset = writeTypeAndKey(d.buf, offset, TypeTimestamp, k)
	offset = writeUint32(d.buf, offset, v.I)
	writeUint32(d.buf, offset, v.T)
	d.buf[len(d.buf)-1] = 0
	return d
}

// AddInt64 ...
func (d *Doc) AddInt64(k string, v int64) *Doc {
	if d.immutable || !d.valid {
		d.err = errImmutableInvalid
		return d
	}
	offset := len(d.buf) - 1
	// Add space for type byte + len(key) + null byte + 8 for int64
	d.grow(10 + len(k))
	offset = writeTypeAndKey(d.buf, offset, TypeInt64, k)
	writeInt64(d.buf, offset, v)
	d.buf[len(d.buf)-1] = 0
	return d
}

// AddDecimal128 ...
func (d *Doc) AddDecimal128(k string, v primitive.Decimal128) *Doc {
	if d.immutable || !d.valid {
		d.err = errImmutableInvalid
		return d
	}
	h, l := v.GetBytes()
	offset := len(d.buf) - 1
	// Add space for type byte + len(key) + null byte + 16 for decimal128
	d.grow(18 + len(k))
	offset = writeTypeAndKey(d.buf, offset, TypeDecimal128, k)
	offset = writeUint64(d.buf, offset, l)
	writeUint64(d.buf, offset, h)
	d.buf[len(d.buf)-1] = 0
	return d
}

// AddMaxKey ...
func (d *Doc) AddMaxKey(k string) *Doc {
	if d.immutable || !d.valid {
		d.err = errImmutableInvalid
		return d
	}
	offset := len(d.buf) - 1
	// Add space for type byte + len(key) + null byte; no data bytes
	d.grow(2 + len(k))
	offset = writeTypeAndKey(d.buf, offset, TypeMaxKey, k)
	d.buf[len(d.buf)-1] = 0
	return d
}

// AddMinKey ...
func (d *Doc) AddMinKey(k string) *Doc {
	if d.immutable || !d.valid {
		d.err = errImmutableInvalid
		return d
	}
	offset := len(d.buf) - 1
	// Add space for type byte + len(key) + null byte; no data bytes
	d.grow(2 + len(k))
	offset = writeTypeAndKey(d.buf, offset, TypeMinKey, k)
	d.buf[len(d.buf)-1] = 0
	return d
}
