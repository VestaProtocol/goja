package goja

import (
	"hash/maphash"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/dop251/goja/unistring"
)

type asciiString string

type asciiRuneReader struct {
	s   asciiString
	pos int
}

func (rr *asciiRuneReader) ReadRune() (r rune, size int, err error) {
	if rr.pos < len(rr.s) {
		r = rune(rr.s[rr.pos])
		size = 1
		rr.pos++
	} else {
		err = io.EOF
	}
	return
}

func (s asciiString) reader() io.RuneReader {
	return &asciiRuneReader{
		s: s,
	}
}

func (s asciiString) utf16Reader() io.RuneReader {
	return s.reader()
}

func (s asciiString) utf16Runes() []rune {
	runes := make([]rune, len(s))
	for i := 0; i < len(s); i++ {
		runes[i] = rune(s[i])
	}
	return runes
}

// ss must be trimmed
func stringToInt(ss string) (int64, error) {
	if ss == "" {
		return 0, nil
	}
	if ss == "-0" {
		return 0, strconv.ErrSyntax
	}
	if len(ss) > 2 {
		switch ss[:2] {
		case "0x", "0X":
			return strconv.ParseInt(ss[2:], 16, 64)
		case "0b", "0B":
			return strconv.ParseInt(ss[2:], 2, 64)
		case "0o", "0O":
			return strconv.ParseInt(ss[2:], 8, 64)
		}
	}
	return strconv.ParseInt(ss, 10, 64)
}

func (s asciiString) _toInt() (int64, error) {
	return stringToInt(strings.TrimSpace(string(s)))
}

func isRangeErr(err error) bool {
	if err, ok := err.(*strconv.NumError); ok {
		return err.Err == strconv.ErrRange
	}
	return false
}

func (s asciiString) ToInteger() int64 {
	if s == "" {
		return 0
	}
	if s == "Infinity" || s == "+Infinity" {
		return math.MaxInt64
	}
	if s == "-Infinity" {
		return math.MinInt64
	}
	i, err := s._toInt()
	if err != nil {
		return -1
	}
	return i
}

func (s asciiString) toString() valueString {
	return s
}

func (s asciiString) ToString() Value {
	return s
}

func (s asciiString) String() string {
	return string(s)
}

func (s asciiString) ToBoolean() bool {
	return s != ""
}

func (s asciiString) ToNumber() Value {
	if s == "" {
		return intToValue(0)
	}

	if i, err := s._toInt(); err == nil {
		return intToValue(i)
	}

	return Null()
}

func (s asciiString) ToObject(r *Runtime) *Object {
	return r._newString(s, r.global.StringPrototype)
}

func (s asciiString) SameAs(other Value) bool {
	return s.StrictEquals(other)
}

func (s asciiString) Equals(other Value) bool {
	if s.StrictEquals(other) {
		return true
	}

	if o, ok := other.(valueInt); ok {
		if o1, e := s._toInt(); e == nil {
			return o1 == int64(o)
		}
		return false
	}

	if _, ok := other.(valueBool); ok {
		return false
	}

	if o, ok := other.(*Object); ok {
		return s.Equals(o.toPrimitive())
	}
	return false
}

func (s asciiString) StrictEquals(other Value) bool {
	if otherStr, ok := other.(asciiString); ok {
		return s == otherStr
	}
	if otherStr, ok := other.(*importedString); ok {
		if otherStr.u == nil {
			return string(s) == otherStr.s
		}
	}
	return false
}

func (s asciiString) baseObject(r *Runtime) *Object {
	ss := r.stringSingleton
	ss.value = s
	ss.setLength()
	return ss.val
}

func (s asciiString) hash(hash *maphash.Hash) uint64 {
	_, _ = hash.WriteString(string(s))
	h := hash.Sum64()
	hash.Reset()
	return h
}

func (s asciiString) charAt(idx int) rune {
	return rune(s[idx])
}

func (s asciiString) length() int {
	return len(s)
}

func (s asciiString) concat(other valueString) valueString {
	a, u := devirtualizeString(other)
	if u != nil {
		b := make([]uint16, len(s)+len(u))
		b[0] = unistring.BOM
		for i := 0; i < len(s); i++ {
			b[i+1] = uint16(s[i])
		}
		copy(b[len(s)+1:], u[1:])
		return unicodeString(b)
	}
	return s + a
}

func (s asciiString) substring(start, end int) valueString {
	return s[start:end]
}

func (s asciiString) compareTo(other valueString) int {
	switch other := other.(type) {
	case asciiString:
		return strings.Compare(string(s), string(other))
	case unicodeString:
		return strings.Compare(string(s), other.String())
	case *importedString:
		return strings.Compare(string(s), other.s)
	default:
		panic(newTypeError("Internal bug: unknown string type: %T", other))
	}
}

func (s asciiString) index(substr valueString, start int) int {
	a, u := devirtualizeString(substr)
	if u == nil {
		p := strings.Index(string(s[start:]), string(a))
		if p >= 0 {
			return p + start
		}
	}
	return -1
}

func (s asciiString) lastIndex(substr valueString, pos int) int {
	a, u := devirtualizeString(substr)
	if u == nil {
		end := pos + len(a)
		var ss string
		if end > len(s) {
			ss = string(s)
		} else {
			ss = string(s[:end])
		}
		return strings.LastIndex(ss, string(a))
	}
	return -1
}

func (s asciiString) toLower() valueString {
	return asciiString(strings.ToLower(string(s)))
}

func (s asciiString) toUpper() valueString {
	return asciiString(strings.ToUpper(string(s)))
}

func (s asciiString) toTrimmedUTF8() string {
	return strings.TrimSpace(string(s))
}

func (s asciiString) string() unistring.String {
	return unistring.String(s)
}

func (s asciiString) Export() interface{} {
	return string(s)
}

func (s asciiString) ExportType() reflect.Type {
	return reflectTypeString
}
