// Copyright 2018 by David A. Golden. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package bsony

import (
	"testing"
)

func TestNewDoc(t *testing.T) {
	fct := New()
	doc := fct.NewDoc()
	compareDocHex(t, doc, "0500000000", "new document")
	doc.Release()
}

func TestNewDocFromBytes(t *testing.T) {
	cases := []struct {
		in    []byte
		err   error
		label string
	}{
		{
			[]byte{5, 0, 0, 0, 0},
			nil,
			"valid",
		},
		{
			[]byte{},
			errShortDoc,
			"short",
		},
		{
			[]byte{5, 0, 0, 0, 0, 0},
			errInvalidLength,
			"bad length",
		},
		{
			[]byte{5, 0, 0, 0, 1},
			errMissingTerminator,
			"unterminated",
		},
	}

	fct := New()
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			doc, err := fct.NewDocFromBytes(c.in)
			assertErr(t, err, c.err)
			if doc != nil {
				doc.Release()
			}
		})
	}
}

// XXX eventually add cases for initial values in array
func TestNewArray(t *testing.T) {
	fct := New()
	ary := fct.NewArray()
	compareArrayHex(t, ary, "0500000000", "new array")
	ary.Release()
}
