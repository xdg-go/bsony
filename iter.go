package bsony

import (
	"bytes"
	"fmt"
)

// A DocIter ...
//
// An initial call to Next() is required to initialize the first value.
//
// WARNING: the DocIter directly references the underlying data; because buffers
// may be reused, you MUST NOT keep a DocIter beyond the lifetime of the source
// document.
type DocIter struct {
	d      *Doc
	offset int          // start of type byte for an value or terminating null
	keyLen int          // -1 means end-of-doc or null byte not found
	vu     *unsafeValue // view to the value; nil if not yet parsed
}

func newDocIter(d *Doc) *DocIter {
	// XXX Do size/validity check on d?
	i := &DocIter{d: d, offset: 4}
	return i
}

func (i *DocIter) parseNextValue() {
	// If offset is at/beyond the end of the buffer, we're done.
	if i.offset >= i.d.Len()-1 {
		i.keyLen = -1
		i.vu = newValueUnsafe(i.d.factory, nil, 0)
		return
	}
	// Key starts after the type byte at the offset and goes to a null byte. If
	// there is no null byte, we have a bad document and let the -1 keyLen
	// signal the problem.
	i.keyLen = bytes.IndexByte(i.d.buf[i.offset+1:], 0)
	if i.keyLen == -1 {
		i.vu = newValueUnsafe(i.d.factory, nil, 0)
		return
	}

	// Data begins after type byte, key length and null byte
	i.vu = newValueUnsafe(i.d.factory, i.d.buf[i.offset+i.keyLen+2:], Type(i.d.buf[i.offset]))

	// If type byte, key, null and i.vu length consumes the full buffer
	// including the terminator byte, then the i.vu has a bad internal length
	if i.offset+i.keyLen+len(i.vu.data)+2 >= i.d.Len() {
		i.vu.err = fmt.Errorf("invalid internal length exceeds container")
	}

}

// Next advances the iterator, if possible.  It returns true if a value is
// available.
func (i *DocIter) Next() bool {
	// On the first call to Next(), i.vu will be nil, so we initialize it
	// without advancing.
	if i.vu == nil {
		i.parseNextValue()
		return i.keyLen != -1
	}

	// If keyLen is already -1, we're already at the end
	if i.keyLen == -1 {
		return false
	}

	// The next value (or final null byte) starts after type byte, keyLen, null
	// byte, and length of ValueUnsafe bytes
	i.offset += i.keyLen + len(i.vu.data) + 2
	i.parseNextValue()

	return i.keyLen != -1
}

// Type returns the type for the current value of the iterator
// or TypeInvalid if the end of the document has reached or the document
// is corrupted.
func (i *DocIter) Type() Type {
	if i.vu == nil {
		return TypeInvalid
	}
	return i.vu.Type()
}

// Key returns the key for the current value of the iterator.  If the
// the end of the document has been reached, the empty string will be
// returned.
func (i *DocIter) Key() string {
	if i.keyLen <= 0 {
		return ""
	}
	// Key begins after type byte
	return string(i.d.buf[i.offset+1 : i.offset+i.keyLen+1])
}

// Value returns a copy of the current value of the iterator or nil if the
// end of the document has been reached or if the value could not be parsed.
// It is safe to keep the value copy and release the source document.
func (i *DocIter) Value() Value {
	if i.Type() == TypeInvalid {
		return nil
	}
	return i.vu.Clone()
}

// ValueUnsafe returns an object with raw type and byte slice of data for the
// current value of the iterator.  The Type field will be zero and the Data
// slice nil if the end of the document is reached.  The Err field will be
// non-nil if an error occured parsing the ValueUnsafe.
//
// WARNING: the ValueUnsafe directly references the underlying data: (1) you
// MUST NOT modify the bytes of a ValueUnsafe; (2) because buffers may be
// reused, you MUST NOT keep a ValueUnsafe beyond the lifetime of the source
// document.
func (i *DocIter) ValueUnsafe() Value {
	return i.vu
}

// Get returns the value of the current value of the iterator or nil if
// the end of the document has been reached or if the value could not be
// parsed.  The Get is always a copy of any underlying data; it is safe to
// keep the result of a Get and release the source document.
func (i *DocIter) Get() interface{} {
	return i.vu.Get()
}

// XXX Should this have methods for typed decoding?  E.g. `Int32OK`?

// Err returns any error from parsing the current value of the iterator.
func (i *DocIter) Err() error {
	if i.Type() == TypeInvalid {
		return fmt.Errorf("invalid value or iterator exhausted")
	}
	return i.vu.Err()
}

// An ArrayIter ...
//
// WARNING: the ArrayIter directly references the underlying data; because
// buffers may be reused, you MUST NOT keep an ArrayIter beyond the lifetime of
// the source array.
type ArrayIter struct {
	di *DocIter
	n  int
}

func newArrayIter(a *Array) *ArrayIter {
	// XXX Do size/validity check on a?
	di := &DocIter{d: a.d, offset: 4}
	return &ArrayIter{di: di, n: -1}
}

// Next advances the iterator, if possible
func (i *ArrayIter) Next() bool {
	if i.di.Next() {
		i.n++
		return true
	}
	i.n = -1
	return false
}

// Type returns the type for the current value of the iterator
// or TypeInvalid if the end of the array has reached or the array
// is corrupted.
func (i *ArrayIter) Type() Type {
	return i.di.Type()
}

// Index returns a zero-based index for the current value of the iterator.  If
// Next has not been called, or if the end of the array has been reached, or if
// the value could not be parsed, this method returns -1.
func (i *ArrayIter) Index() int {
	return i.n
}

// Value ... (returns a copy)
func (i *ArrayIter) Value() Value {
	return i.di.Value()
}

// ValueUnsafe returns an object with raw type and byte slice of data for the
// current value of the iterator.  The Type field will be zero and the Data
// slice nil if the end of the array is reached.  The Err field will be
// non-nil if an error occured parsing the ValueUnsafe.
//
// WARNING: the ValueUnsafe directly references the underlying data: (1) you
// MUST NOT modify the bytes of a ValueUnsafe; (2) because buffers may be
// reused, you MUST NOT keep a ValueUnsafe beyond the lifetime of the source
// array.
func (i *ArrayIter) ValueUnsafe() Value {
	return i.di.ValueUnsafe()
}

// Get returns the value of the current value of the iterator or nil if
// the end of the document has been reached or if the value could not be
// parsed.  The Get is always a copy of any underlying data; it is safe to
// keep the result of a Get and release the source document.
func (a *ArrayIter) Get() interface{} {
	return a.di.Get()
}

// Err returns any error from parsing the current value of the iterator.
func (a *ArrayIter) Err() error {
	return a.di.Err()
}
