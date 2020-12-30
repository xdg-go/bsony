package bsony

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
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

const (
	rfc3339Milli = "2006-01-02T15:04:05.999Z07:00"
)

func TestAdd(t *testing.T) {
	fct := New()
	testArray := fct.NewArray()
	testBinary := primitive.Binary{Subtype: 1, Data: []byte{255, 255}}
	testDoc := fct.NewDoc()
	testOID, _ := primitive.ObjectIDFromHex("56e1fc72e0c917e9c4714161")
	testRegex := primitive.Regex{Pattern: "abc", Options: "im"}
	testTime, _ := time.Parse(rfc3339Milli, "2012-12-24T12:15:30.501Z")
	testDBPointer := primitive.DBPointer{DB: "b", Pointer: testOID}
	testCodeScope := CodeWithScope{Code: "abcd", Scope: fct.NewDoc()}
	testPrimitiveCodeScope := primitive.CodeWithScope{Code: "abcd", Scope: &bson.D{}}
	testDecimal128, _ := primitive.ParseDecimal128("0")

	addCases := []AddTestCase{
		{"float64", "d", float64(1.0), "10000000016400000000000000F03F00"},
		{"float32", "d", float32(1.0), "10000000016400000000000000F03F00"},
		{"string", "a", "b", "0E00000002610002000000620000"},
		{"doc (pointer)", "x", testDoc, "0D000000037800050000000000"},
		{"doc (value)", "x", *testDoc, "0D000000037800050000000000"},
		{"array (pointer)", "x", testArray, "0D000000047800050000000000"},
		{"array (value)", "x", *testArray, "0D000000047800050000000000"},
		{"binary (pointer)", "x", &testBinary, "0F0000000578000200000001FFFF00"},
		{"binary (value)", "x", testBinary, "0F0000000578000200000001FFFF00"},
		{"undefined", "a", primitive.Undefined{}, "0800000006610000"},
		{"oid", "a", testOID, "1400000007610056E1FC72E0C917E9C471416100"},
		{"boolean", "b", true, "090000000862000100"},
		{"datetime (pointer)", "a", &testTime, "10000000096100C5D8D6CC3B01000000"},
		{"datetime (value)", "a", testTime, "10000000096100C5D8D6CC3B01000000"},
		{"datetime (primitive)", "a", primitive.NewDateTimeFromTime(testTime), "10000000096100C5D8D6CC3B01000000"},
		{"null", "a", nil, "080000000a610000"},
		{"regex (pointer)", "a", &testRegex, "0F0000000B610061626300696D0000"},
		{"regex (value)", "a", testRegex, "0F0000000B610061626300696D0000"},
		{"DBPointer (pointer)", "a", &testDBPointer, "1A0000000C610002000000620056E1FC72E0C917E9C471416100"},
		{"DBPointer (value)", "a", testDBPointer, "1A0000000C610002000000620056E1FC72E0C917E9C471416100"},
		{"JavaScript", "a", primitive.JavaScript("b"), "0E0000000D610002000000620000"},
		{"Symbol", "a", primitive.Symbol("b"), "0E0000000E610002000000620000"},
		{"Code with Scope (bsony)", "a", testCodeScope, "1A0000000F610012000000050000006162636400050000000000"},
		{"Code with Scope (primitive)", "a", testPrimitiveCodeScope, "1A0000000F610012000000050000006162636400050000000000"},
		{"CodeWithScope (nil scope)", "a", CodeWithScope{Code: "abcd"}, "1A0000000F610012000000050000006162636400050000000000"},
		{"int32", "i", int32(-1), "0C000000106900FFFFFFFF00"},
		{"timestamp", "a", primitive.Timestamp{T: 123456789, I: 42}, "100000001161002A00000015CD5B0700"},
		{"int64", "a", int64(1), "10000000126100010000000000000000"},
		{"decimal128 (pointer)", "d", &testDecimal128, "180000001364000000000000000000000000000000403000"},
		{"decimal128 (value)", "d", testDecimal128, "180000001364000000000000000000000000000000403000"},
		{"minkey", "a", primitive.MinKey{}, "08000000FF610000"},
		{"maxkey", "a", primitive.MaxKey{}, "080000007F610000"},
	}

	for _, c := range addCases {
		d := fct.NewDoc()
		d.Add(c.K, c.V)
		compareDocHex(t, d, c.D, c.L)
		d.Release()
	}
}
