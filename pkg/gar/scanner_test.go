package gar

import (
	"bufio"
	"strings"
	"testing"
)

func equal[T comparable](t *testing.T, got, want T, msg string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %v, want %v", msg, got, want)
	}
}

func TestScanner(t *testing.T) {
	const doc = "\xef\xbb\xbf<?xml version=\"1.0\" encoding=\"utf-8\"?><ITEMS>" +
		`<ITEM ID="1" OBJECTID="10" PARENTOBJID="5" ISACTIVE="1" PATH="1.5.10" />` +
		`<ITEM ID="2" OBJECTID="11" PARENTOBJID="10" NAME="a &gt; b &quot;c&quot;" ISACTIVE="0" />` +
		`<OTHER OBJECTID="99" />` +
		`<ITEM ID="3" OBJECTID="12" NAME="x>y" />` +
		`</ITEMS>`

	s := NewScanner(strings.NewReader(doc), "ITEM")

	equal(t, s.Next(), true, "first ITEM")
	id, ok := s.Int("OBJECTID")
	equal(t, ok, true, "OBJECTID parsed")
	equal(t, id, int64(10), "OBJECTID must not match inside PARENTOBJID")
	parent, _ := s.Int("PARENTOBJID")
	equal(t, parent, int64(5), "PARENTOBJID")
	equal(t, s.Is("ISACTIVE", "1"), true, "ISACTIVE")
	equal(t, s.String("PATH"), "1.5.10", "PATH")
	equal(t, s.Raw("MISSING") == nil, true, "missing attribute is nil")

	equal(t, s.Next(), true, "second ITEM")
	equal(t, s.String("NAME"), `a > b "c"`, "entities are unescaped")
	equal(t, s.Is("ISACTIVE", "1"), false, "ISACTIVE")

	equal(t, s.Next(), true, "OTHER is skipped")
	equal(t, s.String("NAME"), "x>y", "unescaped > inside a value")

	equal(t, s.Next(), false, "end of document")
	equal(t, s.Err(), nil, "no error at EOF")
}

func TestScannerLongElement(t *testing.T) {
	long := strings.Repeat("я", bufio.MaxScanTokenSize)
	doc := `<ROOT><ITEM NAME="` + long + `" ID="7" /></ROOT>`
	s := NewScanner(strings.NewReader(doc), "ITEM")
	s.r = bufio.NewReaderSize(strings.NewReader(doc), 16) // force ErrBufferFull

	equal(t, s.Next(), true, "long ITEM")
	equal(t, s.String("NAME"), long, "NAME across buffer refills")
	id, _ := s.Int("ID")
	equal(t, id, int64(7), "ID")
	equal(t, s.Next(), false, "end of document")
}

func TestParseInt(t *testing.T) {
	for in, want := range map[string]int64{"0": 0, "807356": 807356, "9223372036854775807": 1<<63 - 1} {
		got, ok := ParseInt([]byte(in))
		equal(t, ok, true, in)
		equal(t, got, want, in)
	}
	for _, in := range []string{"", "-1", "1a", "99999999999999999999"} {
		_, ok := ParseInt([]byte(in))
		equal(t, ok, false, in)
	}
}
