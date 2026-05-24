package jsoncolor

import (
	"bytes"
	"encoding"
	"encoding/base64"
	"errors"
	"math"
	"reflect"
	"sort"
	"strconv"
	"sync"
	"time"
	"unicode/utf8"
	"unsafe"
)

const hex = "0123456789abcdef"

func (e encoder) encodeNull(b []byte, _ unsafe.Pointer) ([]byte, error) {
	return e.clrs.appendNull(b), nil
}

func (e encoder) encodeBool(b []byte, p unsafe.Pointer) ([]byte, error) {
	return e.clrs.appendBool(b, *(*bool)(p)), nil
}

func (e encoder) encodeInt(b []byte, p unsafe.Pointer) ([]byte, error) {
	return e.clrs.appendInt64(b, int64(*(*int)(p))), nil
}

func (e encoder) encodeInt8(b []byte, p unsafe.Pointer) ([]byte, error) {
	return e.clrs.appendInt64(b, int64(*(*int8)(p))), nil
}

func (e encoder) encodeInt16(b []byte, p unsafe.Pointer) ([]byte, error) {
	return e.clrs.appendInt64(b, int64(*(*int16)(p))), nil
}

func (e encoder) encodeInt32(b []byte, p unsafe.Pointer) ([]byte, error) {
	return e.clrs.appendInt64(b, int64(*(*int32)(p))), nil
}

func (e encoder) encodeInt64(b []byte, p unsafe.Pointer) ([]byte, error) {
	return e.clrs.appendInt64(b, *(*int64)(p)), nil
}

func (e encoder) encodeUint(b []byte, p unsafe.Pointer) ([]byte, error) {
	return e.clrs.appendUint64(b, uint64(*(*uint)(p))), nil
}

func (e encoder) encodeUintptr(b []byte, p unsafe.Pointer) ([]byte, error) {
	return e.clrs.appendUint64(b, uint64(*(*uintptr)(p))), nil
}

func (e encoder) encodeUint8(b []byte, p unsafe.Pointer) ([]byte, error) {
	return e.clrs.appendUint64(b, uint64(*(*uint8)(p))), nil
}

func (e encoder) encodeUint16(b []byte, p unsafe.Pointer) ([]byte, error) {
	return e.clrs.appendUint64(b, uint64(*(*uint16)(p))), nil
}

func (e encoder) encodeUint32(b []byte, p unsafe.Pointer) ([]byte, error) {
	return e.clrs.appendUint64(b, uint64(*(*uint32)(p))), nil
}

func (e encoder) encodeUint64(b []byte, p unsafe.Pointer) ([]byte, error) {
	return e.clrs.appendUint64(b, *(*uint64)(p)), nil
}

func (e encoder) encodeFloat32(b []byte, p unsafe.Pointer) ([]byte, error) {
	if e.clrs == nil {
		return e.encodeFloat(b, float64(*(*float32)(p)), 32)
	}

	b = append(b, e.clrs.Number...)
	var err error
	b, err = e.encodeFloat(b, float64(*(*float32)(p)), 32)
	b = append(b, ansiReset...)
	return b, err
}

func (e encoder) encodeFloat64(b []byte, p unsafe.Pointer) ([]byte, error) {
	if e.clrs == nil {
		return e.encodeFloat(b, *(*float64)(p), 64)
	}

	b = append(b, e.clrs.Number...)
	var err error
	b, err = e.encodeFloat(b, *(*float64)(p), 64)
	b = append(b, ansiReset...)
	return b, err
}

func (e encoder) encodeFloat(b []byte, f float64, bits int) ([]byte, error) {
	switch {
	case math.IsNaN(f):
		return b, &UnsupportedValueError{Value: reflect.ValueOf(f), Str: "NaN"}
	case math.IsInf(f, 0):
		return b, &UnsupportedValueError{Value: reflect.ValueOf(f), Str: "inf"}
	}

	// Convert as if by ES6 number to string conversion.
	// This matches most other JSON generators.
	// See golang.org/issue/6384 and golang.org/issue/14135.
	// Like fmt %g, but the exponent cutoffs are different
	// and exponents themselves are not padded to two digits.
	abs := math.Abs(f)
	fmt := byte('f')
	// Note: Must use float32 comparisons for underlying float32 value to get precise cutoffs right.
	if abs != 0 {
		if bits == 64 && (abs < 1e-6 || abs >= 1e21) || bits == 32 && (float32(abs) < 1e-6 || float32(abs) >= 1e21) {
			fmt = 'e'
		}
	}

	b = strconv.AppendFloat(b, f, fmt, -1, bits)

	if fmt == 'e' {
		// clean up e-09 to e-9
		n := len(b)
		if n >= 4 && b[n-4] == 'e' && b[n-3] == '-' && b[n-2] == '0' {
			b[n-2] = b[n-1]
			b = b[:n-1]
		}
	}

	return b, nil
}

func (e encoder) encodeNumber(b []byte, p unsafe.Pointer) ([]byte, error) {
	n := *(*Number)(p)
	if n == "" {
		n = "0"
	}

	_, _, err := parseNumber(stringToBytes(string(n)))
	if err != nil {
		return b, err
	}

	if e.clrs == nil {
		return append(b, n...), nil
	}

	b = append(b, e.clrs.Number...)
	b = append(b, n...)
	b = append(b, ansiReset...)
	return b, nil
}

func (e encoder) encodeKey(b []byte, p unsafe.Pointer) ([]byte, error) {
	if e.clrs == nil {
		return e.doEncodeString(b, p)
	}

	b = append(b, e.clrs.Key...)
	var err error
	b, err = e.doEncodeString(b, p)
	b = append(b, ansiReset...)
	return b, err
}

func (e encoder) encodeString(b []byte, p unsafe.Pointer) ([]byte, error) {
	if e.clrs == nil {
		return e.doEncodeString(b, p)
	}

	b = append(b, e.clrs.String...)
	var err error
	b, err = e.doEncodeString(b, p)
	b = append(b, ansiReset...)
	return b, err
}

func (e encoder) doEncodeString(b []byte, p unsafe.Pointer) ([]byte, error) {
	s := *(*string)(p)
	i := 0
	j := 0
	escapeHTML := (e.flags & EscapeHTML) != 0

	b = append(b, '"')

	for j < len(s) {
		c := s[j]

		if c >= 0x20 && c <= 0x7f && c != '\\' && c != '"' && (!escapeHTML || (c != '<' && c != '>' && c != '&')) {
			// fast path: most of the time, printable ascii characters are used
			j++
			continue
		}

		switch c {
		case '\\', '"':
			b = append(b, s[i:j]...)
			b = append(b, '\\', c)
			i = j + 1
			j = i
			continue

		case '\n':
			b = append(b, s[i:j]...)
			b = append(b, '\\', 'n')
			i = j + 1
			j = i
			continue

		case '\r':
			b = append(b, s[i:j]...)
			b = append(b, '\\', 'r')
			i = j + 1
			j = i
			continue

		case '\t':
			b = append(b, s[i:j]...)
			b = append(b, '\\', 't')
			i = j + 1
			j = i
			continue

		case '<', '>', '&':
			b = append(b, s[i:j]...)
			b = append(b, `\u00`...)
			b = append(b, hex[c>>4], hex[c&0xF])
			i = j + 1
			j = i
			continue
		}

		// This encodes bytes < 0x20 except for \t, \n and \r.
		if c < 0x20 {
			b = append(b, s[i:j]...)
			b = append(b, `\u00`...)
			b = append(b, hex[c>>4], hex[c&0xF])
			i = j + 1
			j = i
			continue
		}

		r, size := utf8.DecodeRuneInString(s[j:])

		if r == utf8.RuneError && size == 1 {
			b = append(b, s[i:j]...)
			b = append(b, `\ufffd`...)
			i = j + size
			j = i
			continue
		}

		switch r {
		case '\u2028', '\u2029':
			// U+2028 is LINE SEPARATOR.
			// U+2029 is PARAGRAPH SEPARATOR.
			// They are both technically valid characters in JSON strings,
			// but don't work in JSONP, which has to be evaluated as JavaScript,
			// and can lead to security holes there. It is valid JSON to
			// escape them, so we do so unconditionally.
			// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
			b = append(b, s[i:j]...)
			b = append(b, `\u202`...)
			b = append(b, hex[r&0xF])
			i = j + size
			j = i
			continue
		}

		j += size
	}

	b = append(b, s[i:]...)
	b = append(b, '"')
	return b, nil
}

func (e encoder) encodeToString(b []byte, p unsafe.Pointer, encode encodeFunc) ([]byte, error) {
	i := len(b)

	b, err := encode(e, b, p)
	if err != nil {
		return b, err
	}

	j := len(b)
	s := b[i:]

	if b, err = e.doEncodeString(b, unsafe.Pointer(&s)); err != nil {
		return b, err
	}

	n := copy(b[i:], b[j:])
	return b[:i+n], nil
}

func (e encoder) encodeBytes(b []byte, p unsafe.Pointer) ([]byte, error) {
	if e.clrs == nil {
		return e.doEncodeBytes(b, p)
	}

	b = append(b, e.clrs.Bytes...)
	var err error
	b, err = e.doEncodeBytes(b, p)
	return append(b, ansiReset...), err
}

func (e encoder) doEncodeBytes(b []byte, p unsafe.Pointer) ([]byte, error) {
	v := *(*[]byte)(p)
	if v == nil {
		return e.clrs.appendNull(b), nil
	}

	n := base64.StdEncoding.EncodedLen(len(v)) + 2

	if avail := cap(b) - len(b); avail < n {
		newB := make([]byte, cap(b)+(n-avail))
		copy(newB, b)
		b = newB[:len(b)]
	}

	i := len(b)
	j := len(b) + n

	b = b[:j]
	b[i] = '"'
	base64.StdEncoding.Encode(b[i+1:j-1], v)
	b[j-1] = '"'
	return b, nil
}

func (e encoder) encodeDuration(b []byte, p unsafe.Pointer) ([]byte, error) {
	// NOTE: The segmentj encoder does special handling for time.Duration (converts to string).
	//  The stdlib encoder does not. It just outputs the int64 value.
	//  We choose to follow the stdlib pattern, for fuller compatibility.

	b = e.clrs.appendInt64(b, int64(*(*time.Duration)(p)))
	return b, nil

	// NOTE: if we were to follow the segmentj pattern, we'd execute the code below.
	// if e.clrs == nil {
	// 	b = append(b, '"')
	//
	// 	b = appendDuration(b, *(*time.Duration)(p))
	// 	b = append(b, '"')
	// 	return b, nil
	// }
	//
	// b = append(b, e.clrs.Time...)
	// b = append(b, '"')
	// b = appendDuration(b, *(*time.Duration)(p))
	// b = append(b, '"')
	// b = append(b, ansiReset...)
	// return b, nil
}

func (e encoder) encodeTime(b []byte, p unsafe.Pointer) ([]byte, error) {
	if e.clrs == nil {
		t := *(*time.Time)(p)
		b = append(b, '"')
		b = t.AppendFormat(b, time.RFC3339Nano)
		b = append(b, '"')
		return b, nil
	}

	t := *(*time.Time)(p)
	b = append(b, e.clrs.Time...)
	b = append(b, '"')
	b = t.AppendFormat(b, time.RFC3339Nano)
	b = append(b, '"')
	b = append(b, ansiReset...)
	return b, nil
}

func (e encoder) encodeArray(b []byte, p unsafe.Pointer, n int, size uintptr, _ reflect.Type, encode encodeFunc) ([]byte, error) {
	start := len(b)
	var err error

	b = e.clrs.appendPunc(b, '[')

	if n > 0 {
		e.indentr.push()
		for i := 0; i < n; i++ {
			if i != 0 {
				b = e.clrs.appendPunc(b, ',')
			}

			b = e.indentr.appendByte(b, '\n')
			b = e.indentr.appendIndent(b)

			if b, err = encode(e, b, unsafe.Pointer(uintptr(p)+(uintptr(i)*size))); err != nil {
				return b[:start], err
			}
		}
		e.indentr.pop()
		b = e.indentr.appendByte(b, '\n')
		b = e.indentr.appendIndent(b)
	}

	b = e.clrs.appendPunc(b, ']')

	return b, nil
}

func (e encoder) encodeSlice(b []byte, p unsafe.Pointer, size uintptr, t reflect.Type, encode encodeFunc) ([]byte, error) {
	s := (*slice)(p)

	if s.data == nil && s.len == 0 && s.cap == 0 {
		return e.clrs.appendNull(b), nil
	}

	return e.encodeArray(b, s.data, s.len, size, t, encode)
}

func (e encoder) encodeMap(b []byte, p unsafe.Pointer, t reflect.Type, encodeKey, encodeValue encodeFunc, sortKeys sortFunc) ([]byte, error) {
	m := reflect.NewAt(t, p).Elem()
	if m.IsNil() {
		return e.clrs.appendNull(b), nil
	}

	keys := m.MapKeys()
	if sortKeys != nil && (e.flags&SortMapKeys) != 0 {
		sortKeys(keys)
	}

	start := len(b)
	var err error
	b = e.clrs.appendPunc(b, '{')

	if len(keys) != 0 {
		b = e.indentr.appendByte(b, '\n')

		e.indentr.push()
		for i := range keys {
			k := keys[i]
			v := m.MapIndex(k)

			if i != 0 {
				b = e.clrs.appendPunc(b, ',')
				b = e.indentr.appendByte(b, '\n')
			}

			b = e.indentr.appendIndent(b)
			if b, err = encodeKey(e, b, (*iface)(unsafe.Pointer(&k)).ptr); err != nil {
				return b[:start], err
			}

			b = e.clrs.appendPunc(b, ':')
			b = e.indentr.appendByte(b, ' ')

			if b, err = encodeValue(e, b, (*iface)(unsafe.Pointer(&v)).ptr); err != nil {
				return b[:start], err
			}
		}
		b = e.indentr.appendByte(b, '\n')
		e.indentr.pop()
		b = e.indentr.appendIndent(b)
	}

	b = e.clrs.appendPunc(b, '}')
	return b, nil
}

type element struct {
	key string
	val interface{}
	raw RawMessage
}

type mapslice struct {
	elements []element
}

func (m *mapslice) Len() int           { return len(m.elements) }
func (m *mapslice) Less(i, j int) bool { return m.elements[i].key < m.elements[j].key }
func (m *mapslice) Swap(i, j int)      { m.elements[i], m.elements[j] = m.elements[j], m.elements[i] }

var mapslicePool = sync.Pool{
	New: func() interface{} { return new(mapslice) },
}

func (e encoder) encodeMapStringInterface(b []byte, p unsafe.Pointer) ([]byte, error) {
	m := *(*map[string]interface{})(p)
	if m == nil {
		return e.clrs.appendNull(b), nil
	}

	if (e.flags & SortMapKeys) == 0 {
		// Optimized code path when the program does not need the map keys to be
		// sorted.
		b = e.clrs.appendPunc(b, '{')

		if len(m) != 0 {
			b = e.indentr.appendByte(b, '\n')

			var err error
			i := 0

			e.indentr.push()
			for k, v := range m {
				if i != 0 {
					b = e.clrs.appendPunc(b, ',')
					b = e.indentr.appendByte(b, '\n')
				}

				b = e.indentr.appendIndent(b)

				b, err = e.encodeKey(b, unsafe.Pointer(&k))
				if err != nil {
					return b, err
				}

				b = e.clrs.appendPunc(b, ':')
				b = e.indentr.appendByte(b, ' ')

				b, err = Append(b, v, e.flags, e.clrs, e.indentr)
				if err != nil {
					return b, err
				}

				i++
			}
			b = e.indentr.appendByte(b, '\n')
			e.indentr.pop()
			b = e.indentr.appendIndent(b)
		}

		b = e.clrs.appendPunc(b, '}')
		return b, nil
	}

	s := mapslicePool.Get().(*mapslice) //nolint:errcheck
	if cap(s.elements) < len(m) {
		s.elements = make([]element, 0, align(10, uintptr(len(m))))
	}
	for key, val := range m {
		s.elements = append(s.elements, element{key: key, val: val})
	}
	sort.Sort(s)

	start := len(b)
	var err error
	b = e.clrs.appendPunc(b, '{')

	if len(s.elements) > 0 {
		b = e.indentr.appendByte(b, '\n')

		e.indentr.push()
		for i := range s.elements {
			elem := s.elements[i]
			if i != 0 {
				b = e.clrs.appendPunc(b, ',')
				b = e.indentr.appendByte(b, '\n')
			}

			b = e.indentr.appendIndent(b)

			b, _ = e.encodeKey(b, unsafe.Pointer(&elem.key))
			b = e.clrs.appendPunc(b, ':')
			b = e.indentr.appendByte(b, ' ')

			b, err = Append(b, elem.val, e.flags, e.clrs, e.indentr)
			if err != nil {
				break
			}
		}
		b = e.indentr.appendByte(b, '\n')
		e.indentr.pop()
		b = e.indentr.appendIndent(b)
	}

	for i := range s.elements {
		s.elements[i] = element{}
	}

	s.elements = s.elements[:0]
	mapslicePool.Put(s)

	if err != nil {
		return b[:start], err
	}

	b = e.clrs.appendPunc(b, '}')
	return b, nil
}

func (e encoder) encodeMapStringRawMessage(b []byte, p unsafe.Pointer) ([]byte, error) {
	m := *(*map[string]RawMessage)(p)
	if m == nil {
		return e.clrs.appendNull(b), nil
	}

	if (e.flags & SortMapKeys) == 0 {
		// Optimized code path when the program does not need the map keys to be
		// sorted.
		b = e.clrs.appendPunc(b, '{')

		if len(m) != 0 {
			b = e.indentr.appendByte(b, '\n')

			var err error
			i := 0

			e.indentr.push()
			for k := range m {
				if i != 0 {
					b = e.clrs.appendPunc(b, ',')
					b = e.indentr.appendByte(b, '\n')
				}

				b = e.indentr.appendIndent(b)

				b, _ = e.encodeKey(b, unsafe.Pointer(&k))

				b = e.clrs.appendPunc(b, ':')
				b = e.indentr.appendByte(b, ' ')

				v := m[k]
				b, err = e.encodeRawMessage(b, unsafe.Pointer(&v))
				if err != nil {
					break
				}

				i++
			}
			b = e.indentr.appendByte(b, '\n')
			e.indentr.pop()
			b = e.indentr.appendIndent(b)
		}

		b = e.clrs.appendPunc(b, '}')
		return b, nil
	}

	s := mapslicePool.Get().(*mapslice) //nolint:errcheck
	if cap(s.elements) < len(m) {
		s.elements = make([]element, 0, align(10, uintptr(len(m))))
	}
	for key, raw := range m {
		s.elements = append(s.elements, element{key: key, raw: raw})
	}
	sort.Sort(s)

	start := len(b)
	var err error
	b = e.clrs.appendPunc(b, '{')

	if len(s.elements) > 0 {
		b = e.indentr.appendByte(b, '\n')

		e.indentr.push()

		for i := range s.elements {
			if i != 0 {
				b = e.clrs.appendPunc(b, ',')
				b = e.indentr.appendByte(b, '\n')
			}

			b = e.indentr.appendIndent(b)

			elem := s.elements[i]
			b, _ = e.encodeKey(b, unsafe.Pointer(&elem.key))
			b = e.clrs.appendPunc(b, ':')
			b = e.indentr.appendByte(b, ' ')

			b, err = e.encodeRawMessage(b, unsafe.Pointer(&elem.raw))
			if err != nil {
				break
			}
		}
		b = e.indentr.appendByte(b, '\n')
		e.indentr.pop()
		b = e.indentr.appendIndent(b)
	}

	for i := range s.elements {
		s.elements[i] = element{}
	}

	s.elements = s.elements[:0]
	mapslicePool.Put(s)

	if err != nil {
		return b[:start], err
	}

	b = e.clrs.appendPunc(b, '}')
	return b, nil
}

func (e encoder) encodeStruct(b []byte, p unsafe.Pointer, st *structType) ([]byte, error) {
	var err error
	var k string
	var n int
	start := len(b)

	b = e.clrs.appendPunc(b, '{')

	if len(st.fields) > 0 {
		b = e.indentr.appendByte(b, '\n')
	}

	e.indentr.push()

	for i := range st.fields {
		f := &st.fields[i]
		v := unsafe.Pointer(uintptr(p) + f.offset)

		if f.omitempty && f.empty(v) {
			continue
		}

		if n != 0 {
			b = e.clrs.appendPunc(b, ',')
			b = e.indentr.appendByte(b, '\n')
		}

		if (e.flags & EscapeHTML) != 0 {
			k = f.html
		} else {
			k = f.json
		}

		lengthBeforeKey := len(b)
		b = e.indentr.appendIndent(b)

		if e.clrs == nil {
			b = append(b, k...)
		} else {
			b = append(b, e.clrs.Key...)
			b = append(b, k...)
			b = append(b, ansiReset...)
		}

		b = e.clrs.appendPunc(b, ':')

		b = e.indentr.appendByte(b, ' ')

		if b, err = f.codec.encode(e, b, v); err != nil {
			if errors.Is(err, rollback{}) {
				b = b[:lengthBeforeKey]
				continue
			}
			return b[:start], err
		}

		n++
	}

	if n > 0 {
		b = e.indentr.appendByte(b, '\n')
	}

	e.indentr.pop()
	b = e.indentr.appendIndent(b)

	b = e.clrs.appendPunc(b, '}')
	return b, nil
}

type rollback struct{}

func (rollback) Error() string { return "rollback" }

func (e encoder) encodeEmbeddedStructPointer(b []byte, p unsafe.Pointer, _ reflect.Type, _ bool, offset uintptr, encode encodeFunc) ([]byte, error) {
	p = *(*unsafe.Pointer)(p)
	if p == nil {
		return b, rollback{}
	}
	return encode(e, b, unsafe.Pointer(uintptr(p)+offset))
}

func (e encoder) encodePointer(b []byte, p unsafe.Pointer, _ reflect.Type, encode encodeFunc) ([]byte, error) {
	if p = *(*unsafe.Pointer)(p); p != nil {
		return encode(e, b, p)
	}
	return e.encodeNull(b, nil)
}

func (e encoder) encodeInterface(b []byte, p unsafe.Pointer) ([]byte, error) {
	return Append(b, *(*interface{})(p), e.flags, e.clrs, e.indentr)
}

func (e encoder) encodeMaybeEmptyInterface(b []byte, p unsafe.Pointer, t reflect.Type) ([]byte, error) {
	return Append(b, reflect.NewAt(t, p).Elem().Interface(), e.flags, e.clrs, e.indentr)
}

func (e encoder) encodeUnsupportedTypeError(b []byte, _ unsafe.Pointer, t reflect.Type) ([]byte, error) {
	return b, &UnsupportedTypeError{Type: t}
}

// encodeRawMessage encodes a RawMessage to b, applying colorization and
// indentation as configured on the encoder.
//
// The message is re-encoded by walking its JSON tokens on the fly (via
// [Tokenizer]) and emitting them in source order. This is important for two
// reasons: it preserves the original ordering of object keys (an earlier
// implementation round-tripped the message through a map[string]interface{},
// which reordered keys nondeterministically; see issue #19), and it allows the
// individual tokens (keys, strings, numbers, etc.) to be colorized.
//
// When TrustRawMessage is set the message bytes are emitted without an upfront
// validity check. If such a message turns out to be malformed and cannot be
// tokenized, encodeRawMessage honors the "trust" contract by emitting the raw
// bytes verbatim (uncolorized) rather than returning an error.
func (e encoder) encodeRawMessage(b []byte, p unsafe.Pointer) ([]byte, error) {
	v := *(*RawMessage)(p)

	if v == nil {
		return e.clrs.appendNull(b), nil
	}

	trusted := (e.flags & TrustRawMessage) != 0

	s := []byte(v)
	if !trusted {
		var err error
		s, _, err = parseValue(v)
		if err != nil {
			return b, &UnsupportedValueError{Value: reflect.ValueOf(v), Str: err.Error()}
		}
	}

	b, err := e.appendRawMessageTokens(b, s)
	if err != nil && trusted {
		// The message was trusted but is not well-formed JSON. Per the
		// TrustRawMessage contract, emit the bytes verbatim without validation
		// or colorization.
		return e.appendRawMessageVerbatim(b, s)
	}
	return b, err
}

// appendRawMessageVerbatim appends a (possibly malformed) trusted RawMessage to
// b without colorization. It applies HTML escaping when EscapeHTML is set, and
// best-effort indentation when an indenter is configured.
func (e encoder) appendRawMessageVerbatim(b, s []byte) ([]byte, error) {
	if e.indentr == nil || e.indentr.disabled {
		if (e.flags & EscapeHTML) != 0 {
			return appendCompactEscapeHTML(b, s), nil
		}
		return append(b, s...), nil
	}

	if (e.flags & EscapeHTML) != 0 {
		s = appendCompactEscapeHTML(nil, s)
	}

	// Indent reproduces the standard library's indentation. The "prefix" arg is
	// the current indentation. If s is malformed, Indent fails; fall back to
	// appending the bytes unchanged.
	pre := e.indentr.appendIndent(nil)
	buf := &bytes.Buffer{}
	if err := Indent(buf, s, string(pre), e.indentr.indent); err != nil {
		return append(b, s...), nil //nolint:nilerr
	}

	return append(b, buf.Bytes()...), nil
}

// rawFrame tracks the state of one open container (object or array) while
// re-encoding a RawMessage. count is the number of elements (object members or
// array elements) emitted so far in this container.
type rawFrame struct {
	isObject bool
	count    int
}

// appendRawMessageTokens re-encodes the JSON in s, appending the result to b.
// It preserves the source ordering of object keys while applying the
// encoder's colorization and indentation. The output is byte-for-byte
// deterministic for a given input.
//
// The layout (placement of newlines, indentation, commas, and the space after
// a colon) mirrors that produced by encodeMap and encodeArray, so that a
// RawMessage is rendered identically to an equivalent natively-encoded value.
func (e encoder) appendRawMessageTokens(b, s []byte) ([]byte, error) {
	start := len(b)

	stack := make([]rawFrame, 0, 8)

	tok := NewTokenizer(s)
	for tok.Next() {
		d := tok.Delim

		switch d {
		case ':':
			b = e.clrs.appendPunc(b, ':')
			b = e.indentr.appendByte(b, ' ')
			continue
		case ',':
			// Element separators are emitted as part of the prefix of the
			// element that follows; nothing to do here.
			continue
		case '}', ']':
			// Close the current container. If it had any elements, the closing
			// delimiter goes on its own line at the parent's indentation.
			top := len(stack) - 1
			had := top >= 0 && stack[top].count > 0
			if top >= 0 {
				stack = stack[:top]
			}
			e.indentr.pop()
			if had {
				b = e.indentr.appendByte(b, '\n')
				b = e.indentr.appendIndent(b)
			}
			b = e.clrs.appendPunc(b, tok.Value[0])
			continue
		}

		// At this point the token starts a value: either a scalar
		// (string/number/bool/null), an object key, or an opening delimiter.
		// Emit the leading comma/newline/indentation if it begins a new line
		// (an object key or an array element), then the token itself.
		isKey := d == 0 && tok.IsKey
		b = e.appendRawMessageItemPrefix(b, stack, isKey)

		switch d {
		case '{', '[':
			b = e.clrs.appendPunc(b, tok.Value[0])
			e.indentr.push()
			stack = append(stack, rawFrame{isObject: d == '{'})
		default:
			b = e.appendRawMessageScalar(b, tok.Value, isKey)
		}
	}

	if tok.Err != nil {
		return b[:start], tok.Err
	}

	return b, nil
}

// appendRawMessageItemPrefix emits the punctuation and whitespace that precede
// an item that begins a new line: object keys and array elements. Object
// member values (which follow a colon) and the top-level value are emitted
// inline, so they receive no prefix. The first item in a container is preceded
// by a newline+indent; subsequent items are additionally separated from the
// previous item by a comma.
func (e encoder) appendRawMessageItemPrefix(b []byte, stack []rawFrame, isKey bool) []byte {
	top := len(stack) - 1
	if top < 0 {
		// Top-level value; no surrounding container.
		return b
	}

	frame := &stack[top]
	if frame.isObject && !isKey {
		// An object member value: emitted inline after the colon.
		return b
	}

	if frame.count > 0 {
		b = e.clrs.appendPunc(b, ',')
	}
	frame.count++

	b = e.indentr.appendByte(b, '\n')
	b = e.indentr.appendIndent(b)
	return b
}

// appendRawMessageScalar appends a single colorized scalar token to b. The
// token bytes v are the raw JSON representation (e.g. a quoted string,
// number, true/false, or null). isKey reports whether the token is an object
// key, which is colorized using the Key color rather than the value colors.
func (e encoder) appendRawMessageScalar(b []byte, v RawValue, isKey bool) []byte {
	escapeHTML := (e.flags & EscapeHTML) != 0

	if e.clrs == nil {
		if escapeHTML && v.String() {
			return appendCompactEscapeHTML(b, v)
		}
		return append(b, v...)
	}

	var clr Color
	switch {
	case isKey:
		clr = e.clrs.Key
	case v.String():
		clr = e.clrs.String
	case v.Number():
		clr = e.clrs.Number
	case v.True(), v.False():
		clr = e.clrs.Bool
	case v.Null():
		clr = e.clrs.Null
	}

	b = append(b, clr...)
	if escapeHTML && v.String() {
		b = appendCompactEscapeHTML(b, v)
	} else {
		b = append(b, v...)
	}
	return append(b, ansiReset...)
}

// encodeJSONMarshaler suffers from the same defect as encodeRawMessage; it
// can result in keys being reordered.
func (e encoder) encodeJSONMarshaler(b []byte, p unsafe.Pointer, t reflect.Type, pointer bool) ([]byte, error) {
	v := reflect.NewAt(t, p)

	if !pointer {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			return e.clrs.appendNull(b), nil
		}
	}

	m, _ := v.Interface().(Marshaler)
	j, err := m.MarshalJSON()
	if err != nil {
		return b, err
	}

	// We effectively delegate to the encodeRawMessage method.
	return Append(b, RawMessage(j), e.flags, e.clrs, e.indentr)
}

func (e encoder) encodeTextMarshaler(b []byte, p unsafe.Pointer, t reflect.Type, pointer bool) ([]byte, error) {
	v := reflect.NewAt(t, p)

	if !pointer {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			return e.clrs.appendNull(b), nil
		}
	}

	tm, _ := v.Interface().(encoding.TextMarshaler)
	s, err := tm.MarshalText()
	if err != nil {
		return b, err
	}

	if e.clrs == nil {
		return e.doEncodeString(b, unsafe.Pointer(&s))
	}

	b = append(b, e.clrs.TextMarshaler...)
	b, err = e.doEncodeString(b, unsafe.Pointer(&s))
	b = append(b, ansiReset...)
	return b, err
}

func appendCompactEscapeHTML(dst, src []byte) []byte {
	start := 0
	escape := false
	inString := false

	for i, c := range src {
		if !inString {
			switch c {
			case '"': // enter string
				inString = true
			case ' ', '\n', '\r', '\t': // skip space
				if start < i {
					dst = append(dst, src[start:i]...)
				}
				start = i + 1
			}
			continue
		}

		if escape {
			escape = false
			continue
		}

		if c == '\\' {
			escape = true
			continue
		}

		if c == '"' {
			inString = false
			continue
		}

		if c == '<' || c == '>' || c == '&' {
			if start < i {
				dst = append(dst, src[start:i]...)
			}
			dst = append(dst, `\u00`...)
			dst = append(dst, hex[c>>4], hex[c&0xF])
			start = i + 1
			continue
		}

		// Convert U+2028 and U+2029 (E2 80 A8 and E2 80 A9).
		if c == 0xE2 && i+2 < len(src) && src[i+1] == 0x80 && src[i+2]&^1 == 0xA8 {
			if start < i {
				dst = append(dst, src[start:i]...)
			}
			dst = append(dst, `\u202`...)
			dst = append(dst, hex[src[i+2]&0xF])
			start = i + 3
			continue
		}
	}

	if start < len(src) {
		dst = append(dst, src[start:]...)
	}

	return dst
}

// Indenter is used to indent JSON, controlling the prefix and indent
// strings applied at each nesting level. Construct an Indenter with
// [NewIndenter], and pass it (or nil for no indentation) to [Append].
//
// A nil *Indenter is valid and disables indentation, producing compact
// output. All Indenter methods are safe to use with a nil receiver.
type Indenter struct {
	disabled bool
	prefix   string
	indent   string
	depth    int
}

// NewIndenter returns a new Indenter instance for use with [Append]. The
// prefix is prepended to each indented line, and indent is repeated once
// per nesting level, matching the semantics of [Encoder.SetIndent] and
// [encoding/json.Encoder.SetIndent]. If prefix and indent are both empty,
// the Indenter is effectively disabled, and the resulting JSON is compact
// (equivalent to passing a nil *Indenter to Append).
func NewIndenter(prefix, indent string) *Indenter {
	return &Indenter{
		disabled: prefix == "" && indent == "",
		prefix:   prefix,
		indent:   indent,
	}
}

// push increases the indentation level.
func (in *Indenter) push() {
	if in != nil {
		in.depth++
	}
}

// pop decreases the indentation level.
func (in *Indenter) pop() {
	if in != nil {
		in.depth--
	}
}

// appendByte appends a to b if the Indenter is non-nil and enabled.
// Otherwise b is returned unmodified.
func (in *Indenter) appendByte(b []byte, a byte) []byte {
	if in == nil || in.disabled {
		return b
	}

	return append(b, a)
}

// appendIndent writes indentation to b, returning the resulting slice.
// If the Indenter is nil or disabled b is returned unchanged.
func (in *Indenter) appendIndent(b []byte) []byte {
	if in == nil || in.disabled {
		return b
	}

	b = append(b, in.prefix...)
	for i := 0; i < in.depth; i++ {
		b = append(b, in.indent...)
	}
	return b
}
