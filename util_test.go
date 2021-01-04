package bsony

import (
	"encoding/hex"
	"strings"
	"testing"
)

func compareDocs(t *testing.T, got, want *Doc, label string) {
	t.Helper()
	gotHex := strings.ToLower(hex.EncodeToString(got.buf))
	wantHex := strings.ToLower(hex.EncodeToString(want.buf))
	if gotHex != wantHex {
		t.Errorf("%s: docs not equal.\nGot:  %s\nWant: %s", label, gotHex, wantHex)
	}
}

func compareDocHex(t *testing.T, d *Doc, want string, label string) {
	t.Helper()
	got := strings.ToLower(hex.EncodeToString(d.buf))
	want = strings.ToLower(want)
	if got != want {
		t.Errorf("%s: encoded doc incorrect.\nGot:  %s\nWant: %s", label, got, want)
	}
	return
}

func compareArrayHex(t *testing.T, a *Array, want string, label string) {
	t.Helper()
	got := strings.ToLower(hex.EncodeToString(a.d.buf))
	want = strings.ToLower(want)
	if got != want {
		t.Errorf("%s: encoded array incorrect.\nGot:  %s\nWant: %s", label, got, want)
	}
	return
}

func assertErr(t *testing.T, got error, want error) {
	t.Helper()
	if want == nil {
		if got != nil {
			t.Errorf("expected no error, got '%v'", got)
		}
	} else {
		if got == nil {
			t.Errorf("expected an error, got none")
		} else if got != want {
			t.Errorf("expected '%v', got '%v'", want, got)
		}
	}
}
