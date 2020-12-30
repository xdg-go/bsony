package bsony

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"path"
	"strings"
	"testing"
)

var testDir = "testdata/bson-corpus"
var fct = New()

type validCase struct {
	Description       string
	CanonicalBSON     string `json:"canonical_bson"`
	CanonicalExtJSON  string `json:"canonical_extjson"`
	RelaxedExtJSON    string `json:"relaxed_extjson"`
	DegenerateBSON    string `json:"degenerate_bson"`
	DegenerateExtJSON string `json:"degenerate_extjson"`
	ConvertedBSON     string `json:"converted_bson"`
	ConvertedExtJSON  string `json:"converted_extjson"`
}

type errorCase struct {
	Description string
	Bson        string
}

type corpusData struct {
	Description  string
	BsonType     string `json:"bson_type"`
	TestKey      string `json:"test_key"`
	Valid        []validCase
	DecodeErrors []errorCase
}

func TestCorpus(t *testing.T) {
	files, err := ioutil.ReadDir(testDir)
	if err != nil {
		t.Fatal("couldn't read corpus directory")
	}
	for _, f := range files {
		if f.IsDir() || path.Ext(f.Name()) != ".json" {
			continue
		}
		file := path.Join(testDir, f.Name())
		t.Run(f.Name(), func(t *testing.T) { testCorpusFile(t, file) })
	}
}

func testCorpusFile(t *testing.T, file string) {
	t.Helper()
	guts, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatalf("couldn't read %s", file)
	}
	cases := &corpusData{}
	json.Unmarshal(guts, cases)
	for _, c := range cases.Valid {
		t.Run(c.Description, func(t *testing.T) { testValidCase(t, c, cases.TestKey) })
	}
}

func testValidCase(t *testing.T, c validCase, k string) {
	t.Helper()
	cB := strings.ToLower(c.CanonicalBSON)
	cB2 := strings.ToLower(BSONToBSON(t, cB))
	if cB != cB2 {
		t.Errorf("native_to_bson( bson_to_native(cB) ) != cB\n Got: %s\nWant: %s", cB2, cB)
	}
}

func BSONToBSON(t *testing.T, s string) string {
	t.Helper()
	raw, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("error decoding %s", s)
	}
	cB, err := fct.NewDocFromBytes(raw)
	if err != nil {
		t.Fatalf("invalid BSON from %s", s)
	}
	defer func() { cB.Release() }()
	cB2 := fct.NewDoc()
	defer func() { cB2.Release() }()
	iter := cB.Iter()
	for iter.Next() {
		cB2.Add(iter.Key(), iter.Get())
	}
	if cB2.Err() != nil {
		t.Fatalf("error copying BSON doc: %v", cB2.Err())
	}
	buf := make([]byte, cB2.Len())
	cB2.CopyTo(buf)
	return hex.EncodeToString(buf)
}
