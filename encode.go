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

// encodeRawMessage encodes a RawMessage to bytes. Unfortunately, this
// implementation has a deficiency: it uses Unmarshal to build an
// object from the RawMessage, which in the case of a struct, results
// in a map being constructed, and thus the order of the keys is not
// guaranteed to be maintained. A superior implementation would decode and
// then re-encode (with color/indentation) the basic JSON tokens on the fly.
// Note also that if TrustRawMessage is set, and the RawMessage is
// invalid JSON (cannot be parsed by Unmarshal), then this function
// falls back to encodeRawMessageNoParseTrusted, which seems to exhibit the
// correct behavior. It's a bit of a mess, but seems to do the trick.
func (e encoder) encodeRawMessage(b []byte, p unsafe.Pointer) ([]byte, error) {
	v := *(*RawMessage)(p)

	if v == nil {
		return e.clrs.appendNull(b), nil
	}

	var s []byte

	if (e.flags & TrustRawMessage) != 0 {
		s = v
	} else {
		var err error
		s, _, err = parseValue(v)
		if err != nil {
			return b, &UnsupportedValueError{Value: reflect.ValueOf(v), Str: err.Error()}
		}
	}

	var x interface{}
	if err := Unmarshal(s, &x); err != nil {
		return e.encodeRawMessageNoParseTrusted(b, p)
	}

	return Append(b, x, e.flags, e.clrs, e.indentr)
}

// encodeRawMessageNoParseTrusted is a fallback method that is
// used by encodeRawMessage if it fails to parse a trusted RawMessage.
// The (invalid) JSON produced by this method is not colorized.
// This method may have wonky logic or even bugs in it; little effort
// has been expended on it because it's a rarely visited edge case.
func (e encoder) encodeRawMessageNoParseTrusted(b []byte, p unsafe.Pointer) ([]byte, error) {
	v := *(*RawMessage)(p)

	if v == nil {
		return e.clrs.appendNull(b), nil
	}

	var s []byte

	if (e.flags & TrustRawMessage) != 0 {
		s = v
	} else {
		var err error
		s, _, err = parseValue(v)
		if err != nil {
			return b, &UnsupportedValueError{Value: reflect.ValueOf(v), Str: err.Error()}
		}
	}

	if e.indentr == nil {
		if (e.flags & EscapeHTML) != 0 {
			return appendCompactEscapeHTML(b, s), nil
		}

		return append(b, s...), nil
	}

	// In order to get the tests inherited from the original segmentio
	// encoder to work, we need to support indentation.

	// This below is sloppy, but seems to work.
	if (e.flags & EscapeHTML) != 0 {
		s = appendCompactEscapeHTML(nil, s)
	}

	// The "prefix" arg to Indent is the current indentation.
	pre := e.indentr.appendIndent(nil)

	buf := &bytes.Buffer{}
	// And now we just make use of the existing Indent function.
	err := Indent(buf, s, string(pre), e.indentr.indent)
	if err != nil {
		return b, err
	}

	s = buf.Bytes()

	return append(b, s...), nil
}

// encodeJSONMarshaler suffers from the same defect as encodeRawMessage; it
// can result in keys being reordered.
func (e encoder) encodeJSONMarshaler(b []byte, p unsafe.Pointer, t reflect.Type, pointer bool) ([]byte, error) {
	v := reflect.NewAt(t, p)

	if !pointer {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return e.clrs.appendNull(b), nil
		}
	}

	j, err := v.Interface().(Marshaler).MarshalJSON()
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
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return e.clrs.appendNull(b), nil
		}
	}

	s, err := v.Interface().(encoding.TextMarshaler).MarshalText()
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

// indenter is used to indent JSON. The push and pop methods
// change indentation level. The appendIndent method appends the
// computed indentation. The appendByte method appends a byte. All
// methods are safe to use with a nil receiver.
type indenter struct {
	disabled bool
	prefix   string
	indent   string
	depth    int
}

// newIndenter returns a new indenter instance. If prefix and
// indent are both empty, the indenter is effectively disabled,
// and the appendIndent and appendByte methods are no-op.
func newIndenter(prefix, indent string) *indenter {
	return &indenter{
		disabled: prefix == "" && indent == "",
		prefix:   prefix,
		indent:   indent,
	}
}

// push increases the indentation level.
func (in *indenter) push() {
	if in != nil {
		in.depth++
	}
}

// pop decreases the indentation level.
func (in *indenter) pop() {
	if in != nil {
		in.depth--
	}
}

// appendByte appends a to b if the indenter is non-nil and enabled.
// Otherwise b is returned unmodified.
func (in *indenter) appendByte(b []byte, a byte) []byte {
	if in == nil || in.disabled {
		return b
	}

	return append(b, a)
}

// appendIndent writes indentation to b, returning the resulting slice.
// If the indenter is nil or disabled b is returned unchanged.
func (in *indenter) appendIndent(b []byte) []byte {
	if in == nil || in.disabled {
		return b
	}

	b = append(b, in.prefix...)
	for i := 0; i < in.depth; i++ {
		b = append(b, in.indent...)
	}
	return b
}
