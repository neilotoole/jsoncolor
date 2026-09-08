package jsoncolor

import (
	"sort"
	"sync"
	"unsafe"
)

// Specialized encoders for string-keyed maps whose values are strings, bools
// or string slices. Without them these maps take the reflect-based generic
// map path, which allocates a reflect.Value per key and per value. Here the
// map is ranged over directly, and the members collected for sorted output
// live in a pooled slice, so encoding allocates nothing.
//
// One generic body serves all three value types. The value encoder is an
// ordinary encodeFunc; it receives the address of a scratch copy of the value
// held in the pooled slice, so the address never causes a heap allocation.

// mapEntry is one member of a string-keyed map, collected for sorted output.
type mapEntry[V any] struct {
	key string
	val V
}

// mapEntries is the pooled working storage for one encode of a map[string]V:
// the members when the output is sorted, and a scratch slot through which
// each value is handed to the value encoder.
type mapEntries[V any] struct {
	entries []mapEntry[V]
	scratch V
}

func (m *mapEntries[V]) Len() int           { return len(m.entries) }
func (m *mapEntries[V]) Less(i, j int) bool { return m.entries[i].key < m.entries[j].key }
func (m *mapEntries[V]) Swap(i, j int)      { m.entries[i], m.entries[j] = m.entries[j], m.entries[i] }

// mapEntriesPool pools mapEntries for one value type.
type mapEntriesPool[V any] struct {
	pool sync.Pool
}

func (p *mapEntriesPool[V]) get() *mapEntries[V] {
	if m, ok := p.pool.Get().(*mapEntries[V]); ok {
		return m
	}
	return new(mapEntries[V])
}

// put clears m, so the pool holds no references to the caller's data, and
// returns it to the pool.
func (p *mapEntriesPool[V]) put(m *mapEntries[V]) {
	clear(m.entries)
	m.entries = m.entries[:0]
	var zero V
	m.scratch = zero
	p.pool.Put(m)
}

var (
	mapStringStringPool      mapEntriesPool[string]
	mapStringBoolPool        mapEntriesPool[bool]
	mapStringStringSlicePool mapEntriesPool[[]string]
)

// encodeMapStringValues encodes m, applying encodeValue to each value. It is
// the shared body of the specialized string-keyed map encoders. encodeValue
// receives the address of a scratch copy of the value.
func encodeMapStringValues[V any](e encoder, b []byte, m map[string]V, pool *mapEntriesPool[V], encodeValue encodeFunc) ([]byte, error) {
	if m == nil {
		return e.clrs.appendNull(b), nil
	}
	if e.clrs == nil {
		return encodeMapStringValuesFast(e, b, m, pool, encodeValue)
	}

	start := len(b)
	var err error
	s := pool.get()

	b = e.clrs.appendPunc(b, '{')

	if (e.flags & SortMapKeys) == 0 {
		if len(m) != 0 {
			b = e.indentr.appendByte(b, '\n')
			e.indentr.push()

			i := 0
			for k, v := range m {
				if i != 0 {
					b = e.clrs.appendPunc(b, ',')
					b = e.indentr.appendByte(b, '\n')
				}
				b = e.indentr.appendIndent(b)
				b = e.appendKey(b, k)
				b = e.clrs.appendPunc(b, ':')
				b = e.indentr.appendByte(b, ' ')

				s.scratch = v
				if b, err = encodeValue(e, b, unsafe.Pointer(&s.scratch)); err != nil {
					break
				}
				i++
			}

			e.indentr.pop()
			if err == nil {
				b = e.indentr.appendByte(b, '\n')
				b = e.indentr.appendIndent(b)
			}
		}
	} else {
		if cap(s.entries) < len(m) {
			s.entries = make([]mapEntry[V], 0, align(10, uintptr(len(m))))
		}
		for k, v := range m {
			s.entries = append(s.entries, mapEntry[V]{key: k, val: v})
		}
		sort.Sort(s)

		if len(s.entries) != 0 {
			b = e.indentr.appendByte(b, '\n')
			e.indentr.push()

			for i := range s.entries {
				if i != 0 {
					b = e.clrs.appendPunc(b, ',')
					b = e.indentr.appendByte(b, '\n')
				}
				b = e.indentr.appendIndent(b)
				b = e.appendKey(b, s.entries[i].key)
				b = e.clrs.appendPunc(b, ':')
				b = e.indentr.appendByte(b, ' ')

				if b, err = encodeValue(e, b, unsafe.Pointer(&s.entries[i].val)); err != nil {
					break
				}
			}

			e.indentr.pop()
			if err == nil {
				b = e.indentr.appendByte(b, '\n')
				b = e.indentr.appendIndent(b)
			}
		}
	}

	pool.put(s)

	if err != nil {
		return b[:start], err
	}
	return e.clrs.appendPunc(b, '}'), nil
}

// encodeMapStringValuesFast is the colorless path for encodeMapStringValues.
// The caller must have checked that m is non-nil and e.clrs is nil;
// indentation is decided here once.
func encodeMapStringValuesFast[V any](e encoder, b []byte, m map[string]V, pool *mapEntriesPool[V], encodeValue encodeFunc) ([]byte, error) {
	ind := e.indentr.enabled()
	start := len(b)
	var err error
	s := pool.get()

	b = append(b, '{')
	if ind {
		e.indentr.depth++
	}

	n := 0
	if (e.flags & SortMapKeys) == 0 {
		for k, v := range m {
			b = e.appendFastMember(b, n, ind, k)
			s.scratch = v
			if b, err = encodeValue(e, b, unsafe.Pointer(&s.scratch)); err != nil {
				break
			}
			n++
		}
	} else {
		if cap(s.entries) < len(m) {
			s.entries = make([]mapEntry[V], 0, align(10, uintptr(len(m))))
		}
		for k, v := range m {
			s.entries = append(s.entries, mapEntry[V]{key: k, val: v})
		}
		sort.Sort(s)

		for i := range s.entries {
			b = e.appendFastMember(b, n, ind, s.entries[i].key)
			if b, err = encodeValue(e, b, unsafe.Pointer(&s.entries[i].val)); err != nil {
				break
			}
			n++
		}
	}

	pool.put(s)

	if err != nil {
		return b[:start], err
	}

	if ind {
		e.indentr.depth--
		if n > 0 {
			b = append(b, '\n')
			b = e.indentr.appendIndentFast(b)
		}
	}

	return append(b, '}'), nil
}

func (e encoder) encodeMapStringString(b []byte, p unsafe.Pointer) ([]byte, error) {
	return encodeMapStringValues(e, b, *(*map[string]string)(p), &mapStringStringPool, encoder.encodeString)
}

func (e encoder) encodeMapStringBool(b []byte, p unsafe.Pointer) ([]byte, error) {
	return encodeMapStringValues(e, b, *(*map[string]bool)(p), &mapStringBoolPool, encoder.encodeBool)
}

// constructMapStringStringSliceEncodeFunc returns the encoder for
// map[string][]string, given the codec for []string.
func constructMapStringStringSliceEncodeFunc(encodeValue encodeFunc) encodeFunc {
	return func(e encoder, b []byte, p unsafe.Pointer) ([]byte, error) {
		return encodeMapStringValues(e, b, *(*map[string][]string)(p), &mapStringStringSlicePool, encodeValue)
	}
}
