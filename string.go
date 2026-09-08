package jsoncolor

import "math/bits"

// The string encoder scans its input eight bytes at a time for characters
// that need escaping, using the SWAR (SIMD within a register) technique: a
// uint64 holds eight bytes, and arithmetic on the whole word answers "is any
// byte below 0x20" or "is any byte equal to c" for all eight at once, with the
// answer for each byte in that byte's most significant bit. Most strings need
// no escaping at all, and for those the whole string is appended in one copy.
// The approach follows the segmentio/encoding upstream.

const (
	swarLSB = 0x0101010101010101 // the low bit of every byte
	swarMSB = 0x8080808080808080 // the high bit of every byte
)

// escapeIndex returns the index of the first byte of s that the string
// encoder must escape, or -1 if there is none. A byte needs escaping when it
// is outside [0x20, 0x7f], or is a double quote or backslash, or, when
// escapeHTML is set, is one of < > &.
//
// The 8-byte loop may report a byte later in a word as needing escaping when
// it does not (a borrow from a lower byte can set a higher byte's high bit),
// but it never misses one, and the lowest set bit is always a true positive.
// Because the scalar encoder re-checks every byte from the returned index on,
// the only requirement is that no byte before the returned index needs
// escaping, and that holds.
func escapeIndex(s string, escapeHTML bool) int {
	i := 0
	for ; i+8 <= len(s); i += 8 {
		n := load64(s, i)
		// n itself contributes the high bit of every byte >= 0x80.
		mask := n | swarBelow(n, 0x20) | swarContains(n, '"') | swarContains(n, '\\')
		if escapeHTML {
			mask |= swarContains(n, '<') | swarContains(n, '>') | swarContains(n, '&')
		}
		if mask&swarMSB != 0 {
			return i + bits.TrailingZeros64(mask&swarMSB)/8
		}
	}

	for ; i < len(s); i++ {
		c := s[i]
		if c < 0x20 || c > 0x7f || c == '"' || c == '\\' || (escapeHTML && (c == '<' || c == '>' || c == '&')) {
			return i
		}
	}

	return -1
}

// load64 returns bytes s[i:i+8] as a little-endian uint64, so that s[i] is
// the least significant byte regardless of the machine's byte order. The
// compiler recognizes this pattern and emits a single unaligned load.
func load64(s string, i int) uint64 {
	_ = s[i+7] // bounds check hint
	return uint64(s[i]) | uint64(s[i+1])<<8 | uint64(s[i+2])<<16 | uint64(s[i+3])<<24 |
		uint64(s[i+4])<<32 | uint64(s[i+5])<<40 | uint64(s[i+6])<<48 | uint64(s[i+7])<<56
}

// swarBelow returns a word whose high bit is set in every byte of n that is
// below b. The result is only meaningful for bytes of n below 0x80 and for
// b below 0x80; a borrow from a byte that is below b can also set the high
// bit of the next more significant byte.
func swarBelow(n uint64, b byte) uint64 {
	return n - swarExpand(b)
}

// swarContains returns a word whose high bit is set in every byte of n that
// equals b, under the same caveats as swarBelow.
func swarContains(n uint64, b byte) uint64 {
	return (n ^ swarExpand(b)) - swarLSB
}

// swarExpand replicates b into all eight bytes of a word.
func swarExpand(b byte) uint64 {
	return swarLSB * uint64(b)
}
