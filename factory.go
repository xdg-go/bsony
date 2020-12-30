// Copyright 2018 by David A. Golden. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

// package bsony ...
package bsony

// A Factory object is a factory for generating BSON documents and arrays.  If Pool
// is nil, byte slices will be created as needed and not recycled.
type Factory struct {
	pool ByteSlicePool

	// XXX should we have pools for D, A, Value, etc.?
}

// New returns a new Factory based on a byte slice pool with minimum slice
// capacity of 256 bytes and no maximum slice capacity.
func New() *Factory {
	return NewFromPool(NewBytePool(256, -1))
}

// NewFromPool returns a new Factory from a provided ByteSlicePool
// capacity of 256 bytes and no maximum slice capacity.
func NewFromPool(pool ByteSlicePool) *Factory {
	return &Factory{pool: pool}
}

// NewDoc returns a new, empty BSON document
func (f *Factory) NewDoc() *Doc {
	buf := f.pool.Get()
	d := &Doc{factory: f, buf: buf, valid: true}
	d.grow(5)
	return d
}

// Doc returns a BSON document based on a slice of bytes.  The document
// takes ownership of buf and the caller should not use it after calling
// NewDocFromBytes.
func (f *Factory) NewDocFromBytes(buf []byte) (*Doc, error) {
	if err := validateBSONFraming(buf); err != nil {
		return nil, err
	}
	return &Doc{factory: f, buf: buf, valid: true}, nil
}

// NewArray returns a BSON array.  Any arguments will be added to the array.
func (f *Factory) NewArray(xs ...interface{}) *Array {
	ary := &Array{d: f.NewDoc()}
	ary.Add(xs...)
	return ary
}

// release returns a byte slice to a pool
func (f *Factory) release(bs []byte) {
	f.pool.Put(bs)
}

// resize changes the size of a byte slice via the pool
func (f *Factory) resize(bs []byte, size int) []byte {
	return f.pool.Resize(bs, size)
}
