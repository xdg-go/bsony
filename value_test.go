package bsony

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"
)

// 0x01 => 8,
// 0x02 => 5,
// 0x03 => 5,
// 0x04 => 5,
// 0x05 => 5,
// 0x06 => 0,
// 0x07 => 12,
// 0x08 => 1,
// 0x09 => 8,
// 0x0A => 0,
// 0x0B => 2,
// 0x0C => 17,
// 0x0D => 5,
// 0x0E => 5,
// 0x0F => 11,
// 0x10 => 4,
// 0x11 => 8,
// 0x12 => 8,
// 0x13 => 16,
// 0x7F => 0,
// 0xFF => 0,

// Every type except null, undefined, minkey, and maxkey requires a fixed or
// minimum number of bytes.
func TestNewUnsafeValue_TooShort(t *testing.T) {
	fct := New()

	cases := []struct {
		minLen int
		types  []Type
	}{
		{
			minLen: 1,
			types:  []Type{TypeBoolean},
		},
		{
			minLen: 2,
			types:  []Type{TypeRegex},
		},
		{
			minLen: 4,
			types:  []Type{TypeInt32},
		},
		{
			minLen: 5,
			types:  []Type{TypeString, TypeEmbeddedDocument, TypeArray, TypeBinary, TypeSymbol, TypeJavaScript},
		},
		{
			minLen: 8,
			types:  []Type{TypeDouble, TypeInt64, TypeDateTime, TypeTimestamp},
		},
		{
			minLen: 12,
			types:  []Type{TypeObjectID},
		},
		{
			minLen: 14,
			types:  []Type{TypeCodeWithScope},
		},
		{
			minLen: 16,
			types:  []Type{TypeDecimal128},
		},
		{
			minLen: 17,
			types:  []Type{TypeDBPointer},
		},
	}

	for _, c := range cases {
		for _, bt := range c.types {
			// nil buffer
			uv := newValueUnsafe(fct, nil, bt)
			if uv.Err() != errShortDoc {
				t.Errorf("for type %s with nil buffer, expected '%v', got '%v'", bt, errShortDoc, uv.Err())
			}
			uv.Release()

			// short buffer
			buf := make([]byte, c.minLen-1)
			uv = newValueUnsafe(fct, buf, bt)
			if uv.Err() != errShortDoc {
				t.Errorf("for type %s, expected '%v', got '%v'", bt, errShortDoc, uv.Err())
			}
			uv.Release()
		}
	}
}

// Some types have a leading length, though they vary
// whether or not the length includes itself.  This test covers types that have
// no other internal structure.
func TestNewUnsafeValue_BadLeadingLength(t *testing.T) {
	fct := New()

	cases := []struct {
		bt                 Type
		nullLocations      []int
		lengthIncludesSelf bool
		minLen             int
		unTerminated       bool
	}{
		{
			bt:                 TypeString,
			nullLocations:      []int{4},
			lengthIncludesSelf: false,
			minLen:             5,
		},
		{
			bt:                 TypeEmbeddedDocument,
			nullLocations:      []int{4},
			lengthIncludesSelf: true,
			minLen:             5,
		},
		{
			bt:                 TypeArray,
			nullLocations:      []int{4},
			lengthIncludesSelf: true,
			minLen:             5,
		},
		{
			bt:                 TypeBinary,
			nullLocations:      []int{4},
			lengthIncludesSelf: false,
			minLen:             5,
			unTerminated:       true,
		},
		{
			bt:                 TypeJavaScript,
			nullLocations:      []int{4},
			lengthIncludesSelf: false,
			minLen:             5,
		},
		{
			bt:                 TypeSymbol,
			nullLocations:      []int{4},
			lengthIncludesSelf: false,
			minLen:             5,
		},
	}

	for _, c := range cases {
		// First test case: a minimum length buffer with a leading length that
		// exceeds that.
		buf := make([]byte, c.minLen)
		l := c.minLen + 1
		if !c.lengthIncludesSelf {
			l -= 4
		}
		writeInt32(buf, 0, int32(l))

		uv := newValueUnsafe(fct, buf, c.bt)
		if uv.Err() != errShortDoc {
			t.Errorf("for type %s, expected '%v', got '%v'", c.bt, errShortDoc, uv.Err())
		}
		uv.Release()

		// Second test case: a buffer that exceeds the minimum length, with
		// non-zero sentinal values afterwards, and a leading length that
		// exceeds the minimum length to catch the sentinal.
		buf = bytes.Repeat([]byte{0xff}, c.minLen)
		writeInt32(buf, 0, int32(l))
		for _, n := range c.nullLocations {
			buf[n] = 0
		}
		uv = newValueUnsafe(fct, buf, c.bt)
		if uv.Err() != errShortDoc {
			t.Errorf("for type %s, expected '%v', got '%v'", c.bt, errShortDoc, uv.Err())
		}
		uv.Release()
	}
}

// Code with scope has several ways the internal structure can be invalid
func TestNewUnsafeValue_BadCodeWithScope(t *testing.T) {
	fct := New()

	cases := []struct {
		label  string
		src    string
		errStr string
	}{
		{
			label:  "length exceeds buffer",
			src:    "0f000000 01000000 00 05000000 00",
			errStr: errShortDoc.Error(),
		},
		{
			label:  "length exceeds null terminator",
			src:    "0f000000 01000000 00 05000000 00 ff",
			errStr: "code with scope: scope missing null terminator",
		},
		{
			label:  "zero string length",
			src:    "0e000000 00000000 05000000 00 00",
			errStr: "invalid, non-positive string length",
		},
		{
			label:  "zero string length",
			src:    "0e000000 0a000000 00 05000000 00",
			errStr: "string length too long",
		},
		{
			label:  "zero string length",
			src:    "0e000000 01000000 00 06000000 00",
			errStr: "scope size invalid",
		},
	}

	for _, c := range cases {
		stripped := strings.ReplaceAll(c.src, " ", "")
		buf, err := hex.DecodeString(stripped)
		if err != nil {
			t.Fatal(err)
		}
		uv := newValueUnsafe(fct, buf, TypeCodeWithScope)
		if uv.Err() == nil || !strings.Contains(uv.Err().Error(), c.errStr) {
			t.Errorf("expected '%v', got '%v'", c.errStr, uv.Err())
		}
		uv.Release()
	}
}
