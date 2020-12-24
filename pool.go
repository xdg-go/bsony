// Copyright 2018 by David A. Golden. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package bsony

import (
	"sync"
)

// A ByteSlicePool provides an abstraction for a pool of []byte objects.
// It provides Get, Put, and Resize methods.  The Resize method allows for
// more control over allocations than relying on the native `append` function
// to grow slices.
type ByteSlicePool interface {
	Get() []byte
	Put(buf []byte)
	Resize(buf []byte, size int) []byte
}

// A BytePool wraps a sync.Pool of byte slices, but constrains byte slices
// created/returned to be between a minimum and maximum capacity.
type BytePool struct {
	pool   *sync.Pool
	minCap int
	maxCap int
}

// NewBytePool constructs a byte slice pool with minimum and maximum capacities
// for byte slices in the pool.  If minCap is negative, new slices will have
// zero capacity.  If maxCap is negative, no maximum will be applied.
func NewBytePool(minCap, maxCap int) *BytePool {
	if minCap < 0 {
		minCap = 0
	}
	return &BytePool{
		minCap: minCap,
		maxCap: maxCap,
		pool:   &sync.Pool{},
	}
}

// Get gives the caller a byte slice from the pool or a new byte slice with
// the pool's configured minimum slice capacity.  The byte slice returned will
// have its storage zeroed and have length zero.
func (p *BytePool) Get() []byte {
	bp := p.pool.Get()
	if bp == nil {
		return make([]byte, 0, p.minCap)
	}
	buf := bp.([]byte)
	buf = buf[0:cap(buf)]
	for i := range buf {
		buf[i] = 0
	}
	return buf[0:0]
}

// Put returns a byte slice to the pool if the capacity is less than or equal
// to the pool's configured maximum slice capacity.
func (p *BytePool) Put(buf []byte) {
	if p.maxCap < 0 || cap(buf) <= p.maxCap {
		p.pool.Put(buf)
	}
	return
}

// Resize returns a slice of the desired length.  If the underlying capacity is
// insufficient, a copy of the slice with doubled capacity is returned.  This is
// an intentional leaky pool abstraction, which minimizes amortized allocations
// by avoiding recyling small slices back to the pool.
func (p *BytePool) Resize(buf []byte, size int) []byte {
	if size < cap(buf) {
		return buf[0:size]
	}
	temp := make([]byte, size, cap(buf)*2)
	copy(temp, buf)
	return temp
}
