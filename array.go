// Copyright 2018 by David A. Golden. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package bsony

import (
	"io"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// An Array ...
type Array struct {
	// internally, an A is a wrapper around a D, plus an indication of its
	// length.
	n int  // number of elements
	d *Doc // underlying storage; keys are indices
}

// Valid indicates if the array is valid for use.  An array is invalid after its
// storage has been released.
func (a *Array) Valid() bool {
	return a.d.valid
}

// Err returns any error recorded on the array.
func (a *Array) Err() error {
	return a.d.err
}

// Release returns allocated space.  After calling this method, the
// array is invalid.
func (a *Array) Release() {
	a.d.Release()
}

// Iter ...
func (a *Array) Iter() *ArrayIter {
	return &ArrayIter{di: newDocIter(a.d)}
}

// Reader ...
func (a *Array) Reader() (io.Reader, error) {
	return a.d.Reader()
}

// CopyTo ...
func (a *Array) CopyTo(dst []byte) int {
	return a.d.CopyTo(dst)
}

// Len ...
func (a *Array) Len() int {
	return len(a.d.buf)
}

// Concat ..
func (a *Array) Concat(src *Array) *Array {
	// XXX Need to iterate src and add valueUnsafe to keep key indexes
	// correct.
	return a
}

// Clone ...
func (a *Array) Clone() *Array {
	return &Array{d: a.d.Clone(), n: a.n}
}

// Add ...
func (a *Array) Add(xs ...interface{}) *Array {
	if a.d.immutable || !a.d.valid {
		a.d.err = errImmutableInvalid
		return a
	}
	for _, v := range xs {
		a.d.Add(strconv.Itoa(a.n), v)
		a.n++
	}
	return a
}

// AddDouble ...
func (a *Array) AddDouble(v float64) *Array {
	if a.d.immutable || !a.d.valid {
		a.d.err = errImmutableInvalid
		return a
	}
	a.d.AddDouble(strconv.Itoa(a.n), v)
	a.n++
	return a
}

// AddString ...
func (a *Array) AddString(v string) *Array {
	if a.d.immutable || !a.d.valid {
		a.d.err = errImmutableInvalid
		return a
	}
	a.d.AddString(strconv.Itoa(a.n), v)
	a.n++
	return a
}

// AddDoc ...
func (a *Array) AddDoc(v *Doc) *Array {
	if a.d.immutable || !a.d.valid {
		a.d.err = errImmutableInvalid
		return a
	}
	a.d.AddDoc(strconv.Itoa(a.n), v)
	a.n++
	return a
}

// AddArray ...
func (a *Array) AddArray(v *Array) *Array {
	if a.d.immutable || !a.d.valid {
		a.d.err = errImmutableInvalid
		return a
	}
	a.d.AddArray(strconv.Itoa(a.n), v)
	a.n++
	return a
}

// AddBinary ...
func (a *Array) AddBinary(v *Binary) *Array {
	if a.d.immutable || !a.d.valid {
		a.d.err = errImmutableInvalid
		return a
	}
	a.d.AddBinary(strconv.Itoa(a.n), v)
	a.n++
	return a
}

// AddOID ...
func (a *Array) AddOID(v primitive.ObjectID) *Array {
	if a.d.immutable || !a.d.valid {
		a.d.err = errImmutableInvalid
		return a
	}
	a.d.AddOID(strconv.Itoa(a.n), v)
	a.n++
	return a
}

// AddBool ...
func (a *Array) AddBool(v bool) *Array {
	if a.d.immutable || !a.d.valid {
		a.d.err = errImmutableInvalid
		return a
	}
	a.d.AddBool(strconv.Itoa(a.n), v)
	a.n++
	return a
}

// AddDateTime ...
func (a *Array) AddDateTime(v time.Time) *Array {
	if a.d.immutable || !a.d.valid {
		a.d.err = errImmutableInvalid
		return a
	}
	a.d.AddDateTime(strconv.Itoa(a.n), v)
	a.n++
	return a
}

// AddNull ...
func (a *Array) AddNull() *Array {
	if a.d.immutable || !a.d.valid {
		a.d.err = errImmutableInvalid
		return a
	}
	a.d.AddNull(strconv.Itoa(a.n))
	a.n++
	return a
}

// AddRegex ...
func (a *Array) AddRegex(v Regex) *Array {
	if a.d.immutable || !a.d.valid {
		a.d.err = errImmutableInvalid
		return a
	}
	a.d.AddRegex(strconv.Itoa(a.n), v)
	a.n++
	return a
}

// AddCode ...
func (a *Array) AddCode(v Code) *Array {
	if a.d.immutable || !a.d.valid {
		a.d.err = errImmutableInvalid
		return a
	}
	a.d.AddCode(strconv.Itoa(a.n), v)
	a.n++
	return a
}

// AddInt32 ...
func (a *Array) AddInt32(v int32) *Array {
	if a.d.immutable || !a.d.valid {
		a.d.err = errImmutableInvalid
		return a
	}
	a.d.AddInt32(strconv.Itoa(a.n), v)
	a.n++
	return a
}

// AddTimestamp ...
func (a *Array) AddTimestamp(v Timestamp) *Array {
	if a.d.immutable || !a.d.valid {
		a.d.err = errImmutableInvalid
		return a
	}
	a.d.AddTimestamp(strconv.Itoa(a.n), v)
	a.n++
	return a
}

// AddInt64 ...
func (a *Array) AddInt64(v int64) *Array {
	if a.d.immutable || !a.d.valid {
		a.d.err = errImmutableInvalid
		return a
	}
	a.d.AddInt64(strconv.Itoa(a.n), v)
	a.n++
	return a
}

// AddDecimal128 ...
func (a *Array) AddDecimal128(v Decimal128) *Array {
	if a.d.immutable || !a.d.valid {
		a.d.err = errImmutableInvalid
		return a
	}
	a.d.AddDecimal128(strconv.Itoa(a.n), v)
	a.n++
	return a
}

// AddMaxKey ...
func (a *Array) AddMaxKey() *Array {
	if a.d.immutable || !a.d.valid {
		a.d.err = errImmutableInvalid
		return a
	}
	a.d.AddMaxKey(strconv.Itoa(a.n))
	a.n++
	return a
}

// AddMinKey ...
func (a *Array) AddMinKey() *Array {
	if a.d.immutable || !a.d.valid {
		a.d.err = errImmutableInvalid
		return a
	}
	a.d.AddMinKey(strconv.Itoa(a.n))
	a.n++
	return a
}
