package gar

import (
	"bufio"
	"bytes"
	"errors"
	"html"
	"io"
	"strconv"
)

// Scanner walks the flat GAR XML files (<ROOT><TAG a="…" b="…" />…</ROOT>)
// element by element without building a DOM or reflecting into structs:
// AS_ADM_HIERARCHY alone is ~40 GB uncompressed, so encoding/xml is too slow.
// Attribute values are read lazily from the raw element bytes.
type Scanner struct {
	r       *bufio.Reader
	prefix  []byte
	elem    []byte
	buf     []byte
	needles map[string][]byte
	err     error
}

// NewScanner returns a Scanner yielding only <tag …/> elements.
func NewScanner(r io.Reader, tag string) *Scanner {
	return &Scanner{
		r:       bufio.NewReaderSize(r, 1<<20),
		prefix:  []byte("<" + tag + " "),
		needles: make(map[string][]byte),
	}
}

// Next advances to the next matching element. It returns false at EOF or on
// a read error, which Err reports.
func (s *Scanner) Next() bool {
	for {
		raw, err := s.readElement()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				s.err = err
			}
			return false
		}
		if i := bytes.IndexByte(raw, '<'); i >= 0 && bytes.HasPrefix(raw[i:], s.prefix) {
			s.elem = raw[i:]
			return true
		}
	}
}

// readElement returns the bytes up to and including the next '>' that is not
// inside a quoted attribute value. The slice is valid until the next call.
func (s *Scanner) readElement() ([]byte, error) {
	s.buf = s.buf[:0]
	for {
		chunk, err := s.r.ReadSlice('>')
		switch {
		case err == nil:
		case errors.Is(err, bufio.ErrBufferFull):
			s.buf = append(s.buf, chunk...)
			continue
		default:
			return nil, err
		}
		if len(s.buf) == 0 && bytes.Count(chunk, []byte{'"'})%2 == 0 {
			return chunk, nil
		}
		s.buf = append(s.buf, chunk...)
		if bytes.Count(s.buf, []byte{'"'})%2 == 0 {
			return s.buf, nil
		}
	}
}

// Err returns the first non-EOF read error.
func (s *Scanner) Err() error { return s.err }

// Raw returns the attribute's raw (still XML-escaped) value, or nil when the
// element has no such attribute. Valid until the next call to Next.
func (s *Scanner) Raw(name string) []byte {
	needle, ok := s.needles[name]
	if !ok {
		// The leading space keeps OBJECTID from matching inside PARENTOBJID.
		needle = []byte(" " + name + `="`)
		s.needles[name] = needle
	}
	i := bytes.Index(s.elem, needle)
	if i < 0 {
		return nil
	}
	v, _, ok := bytes.Cut(s.elem[i+len(needle):], []byte{'"'})
	if !ok {
		return nil
	}
	return v
}

// String returns the unescaped attribute value ("" when absent).
func (s *Scanner) String(name string) string {
	v := s.Raw(name)
	if bytes.IndexByte(v, '&') >= 0 {
		return html.UnescapeString(string(v))
	}
	return string(v)
}

// Int returns the attribute parsed as a base-10 integer, or 0 and false when
// it is absent or malformed.
func (s *Scanner) Int(name string) (int64, bool) {
	return ParseInt(s.Raw(name))
}

// Is reports whether the attribute equals value exactly.
func (s *Scanner) Is(name, value string) bool {
	return string(s.Raw(name)) == value
}

// ParseInt parses a non-negative base-10 integer without allocating.
func ParseInt(b []byte) (int64, bool) {
	if len(b) == 0 {
		return 0, false
	}
	if len(b) > 18 { // may overflow; let strconv range-check it
		n, err := strconv.ParseInt(string(b), 10, 64)
		return n, err == nil
	}
	var n int64
	for _, c := range b {
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int64(c-'0')
	}
	return n, true
}
