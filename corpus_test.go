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

func shouldSkip(description string) bool {
	// we don't validate UTF-8 within string-like types
	patterns := []string{"invalid UTF-8", "bad UTF-8"}
	for _, p := range patterns {
		if strings.Contains(description, p) {
			return true
		}
	}
	return false
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
		t.Run("valid "+c.Description, func(t *testing.T) { testValidCase(t, c, cases.TestKey) })
	}
	for _, c := range cases.DecodeErrors {
		if shouldSkip(c.Description) {
			continue
		}
		t.Run("invalid "+c.Description, func(t *testing.T) { testErrorCase(t, c) })
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

func testErrorCase(t *testing.T, c errorCase) {
	t.Helper()
	doc, err := docFromHex(t, c.Bson)
	if err != nil {
		return
	}

	err = visitDoc(t, doc)
	if err != nil {
		return
	}

	t.Fatal("expected error, but got none")
}

func visitDoc(t *testing.T, d *Doc) error {
	iter := d.Iter()
	for iter.Next() {
		if iter.Err() != nil {
			return iter.Err()
		}
		switch iter.Type() {
		case TypeEmbeddedDocument:
			d, _ := iter.ValueUnsafe().Get().(*Doc)
			err := visitDoc(t, d)
			if err != nil {
				return err
			}
		case TypeArray:
			a, _ := iter.ValueUnsafe().Get().(*Array)
			err := visitArray(t, a)
			if err != nil {
				return err
			}
		case TypeCodeWithScope:
			cs, _ := iter.ValueUnsafe().Get().(CodeWithScope)
			err := visitDoc(t, cs.Scope)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func visitArray(t *testing.T, a *Array) error {
	iter := a.Iter()
	for iter.Next() {
		if iter.Err() != nil {
			return iter.Err()
		}
		switch iter.Type() {
		case TypeEmbeddedDocument:
			d, _ := iter.ValueUnsafe().Get().(*Doc)
			err := visitDoc(t, d)
			if err != nil {
				return err
			}
		case TypeArray:
			a, _ := iter.ValueUnsafe().Get().(*Array)
			err := visitArray(t, a)
			if err != nil {
				return err
			}
		case TypeCodeWithScope:
			cs, _ := iter.ValueUnsafe().Get().(CodeWithScope)
			err := visitDoc(t, cs.Scope)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func docFromHex(t *testing.T, s string) (*Doc, error) {
	t.Helper()
	raw, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("error decoding %s", s)
	}
	return fct.NewDocFromBytes(raw)
}

func BSONToBSON(t *testing.T, s string) string {
	t.Helper()
	cB, err := docFromHex(t, s)
	if err != nil {
		t.Fatal(err)
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
