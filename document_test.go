package bsony

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestRelease(t *testing.T) {
	fct := New()
	doc := fct.NewDoc()
	if !doc.Valid() {
		t.Error("new doc is invalid")
	}
	doc.Release()
	if doc.Valid() {
		t.Error("released doc is invalid")
	}
	if doc.buf != nil {
		t.Error("released doc buffer non-nil")
	}
	if doc.factory != nil {
		t.Error("released doc factory non-nil")
	}
	assertErr(t, doc.Err(), errBufferReleased)
}

func TestClone(t *testing.T) {
	fct := New()
	doc := fct.NewDoc()
	doc.AddDouble("d", 1.0)
	doc2 := doc.Clone()
	compareDocs(t, doc, doc2, "clone")
}

type AddTestCase struct {
	L string
	K string
	V interface{}
	D string
}

func TestAdd(t *testing.T) {
	testOID, _ := primitive.ObjectIDFromHex("56e1fc72e0c917e9c4714161")

	addCases := []AddTestCase{
		{"float64", "d", float64(1.0), "10000000016400000000000000F03F00"},
		{"float32", "d", float32(1.0), "10000000016400000000000000F03F00"},
		{"string", "a", "b", "0E00000002610002000000620000"},
		// Doc
		// Array
		// Binary
		{"undefined", "a", primitive.Undefined{}, "0800000006610000"},
		{"oid", "a", testOID, "1400000007610056E1FC72E0C917E9C471416100"},
		{"boolean", "b", true, "090000000862000100"},
		// Datetime
		// Null
		// Regex
		// DBPointer
		// Code
		// Symbol
		// Code w/ scope
		{"int32", "i", int32(-1), "0C000000106900FFFFFFFF00"},
		// Timestamp
		{"int64", "a", int64(1), "10000000126100010000000000000000"},
		// Decimal128
		// Minkey
		// Maxkey
	}

	fct := New()
	for _, c := range addCases {
		d := fct.NewDoc()
		d.Add(c.K, c.V)
		compareDocHex(t, d, c.D, c.L)
		d.Release()
	}
}
