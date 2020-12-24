package bsony

import (
	"bytes"
)

// A DocIter ...
//
// WARNING: the DocIter directly references the underlying data; because buffers
// may be reused, you MUST NOT keep a DocIter beyond the lifetime of the source
// document.
type DocIter struct {
	d      *Doc
	offset int           // XXX start of type byte for element??
	keyLen int           // -1 means end-of-doc or null byte not found
	vu     *ownedElement // cached copy
}

func newDocIter(d *Doc) *DocIter {
	// XXX Do size/validity check on d?
	i := &DocIter{d: d, offset: 4}
	i.findKeyLen()
	return i
}

func (i *DocIter) findKeyLen() {
	if i.offset >= i.d.Len()-1 {
		i.keyLen = -1
	}
	// Key starts after the type byte at the offset
	i.keyLen = bytes.IndexByte(i.d.buf[i.offset+1:], 0)
}

// Next advances the iterator, if possible
func (i *DocIter) Next() {
	if i.vu == nil {
		i.ElementUnsafe()
	}
	// Next element (or final null byte) starts after type byte, keyLen, null
	// byte, and length of ElementUnsafe bytes
	i.offset += i.keyLen + len(i.vu.data) + 2
	i.vu = nil
	i.findKeyLen()
}

// Type returns the type for the current element of the iterator
// or TypeInvalid if the end of the document has reached or the document
// is corrupted.
func (i *DocIter) Type() Type {
	return Type(i.d.buf[i.offset])
}

// Key returns the key for the current element of the iterator.  If the
// the end of the document has been reached, the empty string will be
// returned.
func (i *DocIter) Key() string {
	if i.keyLen < 0 {
		return ""
	}
	// Key begins after type byte
	return string(i.d.buf[i.offset+1 : i.offset+i.keyLen+1])
}

// Arr ... (reference to subarr; immutable; lifetime issues)
// Returns invalid A object if iterator is not at an array.
func (i *DocIter) Arr() *Array {
	if i.Type() != TypeArray {
		return &Array{}
	}
	if i.vu == nil {
		i.ElementUnsafe()
	}
	return &Array{d: &Doc{factory: i.d.factory, buf: i.vu.data, valid: true, immutable: true}}
}

// Doc ... (reference to subdoc; immutable; lifetime issues)
func (i *DocIter) Doc() *Doc {
	if i.Type() != TypeEmbeddedDocument {
		return &Doc{}
	}
	if i.vu == nil {
		i.ElementUnsafe()
	}
	return &Doc{factory: i.d.factory, buf: i.vu.data, valid: true, immutable: true}
}

// Element returns a copy of the current element of the iterator or nil if the
// end of the document has been reached or if the element could not be parsed.
// It is safe to keep the element copy and release the source document.
//
// XXX but this element has no key?!  How is this better/different than Value?
// (Not decoded.)
func (i *DocIter) Element() *ownedElement {
	if i.Type() == TypeInvalid {
		return nil
	}
	if i.vu == nil {
		i.ElementUnsafe()
	}
	return i.vu.Clone()
}

// Value returns the value of the current element of the iterator or nil if
// the end of the document has been reached or if the element could not be
// parsed.  The Value is always a copy of any underlying data; it is safe to
// keep a Value and release the source document.
func (i *DocIter) Value() interface{} {
	if i.vu == nil {
		i.ElementUnsafe()
	}
	return i.vu.Value()
}

// XXX Should this have methods for typed decoding?  E.g. `Int32OK`?

// OK returns true when Value() would return a non-nil result (except for
// types that return nil, such as TypeNull).  It is equivalent to checking
// that ElementUnsafe() would not have the Err field set.
func (i *DocIter) OK() bool {
	if i.Type() == TypeInvalid {
		return false
	}
	if i.vu == nil {
		i.ElementUnsafe()
	}
	return i.vu.err == nil
}

// ElementUnsafe returns a struct with raw type and byte slice of data for the
// current element of the iterator.  The Type field will be zero and the Data
// slice nil if the end of the document is reached.  The Err field will be
// non-nil if an error occured parsing the ElementUnsafe.
//
// WARNING: the ElementUnsafe directly references the underlying data: (1) you
// MUST NOT modify the bytes of a ElementUnsafe; (2) because buffers may be
// reused, you MUST NOT keep a ElementUnsafe beyond the lifetime of the source
// document.
func (i *DocIter) ElementUnsafe() *ownedElement {
	if i.vu == nil {
		i.vu = newElementUnsafe(i.d.factory, i.d.buf[i.offset+i.keyLen+2:], i.Type())
	}
	// Data begins after type byte, key length and null byte
	return i.vu
}

// An ArrayIter ...
//
// WARNING: the ArrayIter directly references the underlying data; because
// buffers may be reused, you MUST NOT keep an ArrayIter beyond the lifetime of
// the source array.
type ArrayIter struct {
	di *DocIter
	n  int // XXX this doesn't seem wired up?
}

// Next advances the iterator, if possible
func (i *ArrayIter) Next() {
	i.di.Next()
}

// Type returns the type for the current element of the iterator
// or TypeInvalid if the end of the array has reached or the array
// is corrupted.
func (i *ArrayIter) Type() Type {
	return i.di.Type()
}

// Index returns a zero-based index for the current element of the iterator.
// If the end of the array has been reached or if the element could not be
// parsed, this method returns -1.
func (i *ArrayIter) Index() int {
	if i.di.Type() == TypeInvalid {
		return -1
	}
	return i.n
}

// Arr ... (reference to subarr; immutable; lifetime issues)
// Returns invalid A object if iterator is not at an array.
func (i *ArrayIter) Arr() *Array {
	return i.di.Arr()
}

// Doc ... (reference to subdoc; immutable; lifetime issues)
func (i *ArrayIter) Doc() *Doc {
	return i.di.Doc()
}

// Element ...
func (i *ArrayIter) Element() interface{} {
	return i.di.Element()
}

// ElementOK returns true when Value() would return a non-nil result (except for
// types that return nil, such as TypeNull).  It is equivalent to checking
// that ElementUnsafe() would not have the Err field set.
func (i *ArrayIter) ElementOK() bool {
	return i.di.OK()
}

// ElementUnsafe returns a struct with raw type and byte slice of data for the
// current element of the iterator.  The Type field will be zero and the Data
// slice nil if the end of the array is reached.  The Err field will be
// non-nil if an error occured parsing the ElementUnsafe.
//
// WARNING: the ElementUnsafe directly references the underlying data: (1) you
// MUST NOT modify the bytes of a ElementUnsafe; (2) because buffers may be
// reused, you MUST NOT keep a ElementUnsafe beyond the lifetime of the source
// array.
func (i *ArrayIter) ElementUnsafe() *ownedElement {
	return i.di.ElementUnsafe()
}
