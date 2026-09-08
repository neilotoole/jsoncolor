package jsoncolor

import "encoding/binary"

// Integer formatting. strconv.AppendInt is general over bases and carries a
// 65-byte scratch buffer; JSON only ever needs base 10 and at most 20 digits,
// so a dedicated formatter that writes two digits per step from a lookup
// table is measurably faster on integer-heavy values. Output is identical to
// strconv's, which the tests verify.

// digitPairs holds the decimal representations of 0 through 99, two bytes
// each, so pair i is at digitPairs[2*i : 2*i+2].
const digitPairs = "" +
	"00010203040506070809" +
	"10111213141516171819" +
	"20212223242526272829" +
	"30313233343536373839" +
	"40414243444546474849" +
	"50515253545556575859" +
	"60616263646566676869" +
	"70717273747576777879" +
	"80818283848586878889" +
	"90919293949596979899"

// digitPairs16 holds the same pairs as 16-bit words in the machine's byte
// order, so that a pair is written with one 16-bit store.
var digitPairs16 = func() (t [100]uint16) {
	for i := range t {
		t[i] = binary.NativeEndian.Uint16([]byte(digitPairs[2*i : 2*i+2]))
	}
	return t
}()

// appendInt appends the decimal representation of n to b.
func appendInt(b []byte, n int64) []byte {
	u := uint64(n) //nolint:gosec // two's-complement reinterpretation is intended
	if n < 0 {
		b = append(b, '-')
		u = -u // wraps correctly for math.MinInt64
	}
	return appendUint(b, u)
}

// appendUint appends the decimal representation of u to b.
func appendUint(b []byte, u uint64) []byte {
	if u < 10 {
		return append(b, byte('0'+u))
	}
	if u < 100 {
		return append(b, digitPairs[2*u], digitPairs[2*u+1])
	}

	// Fill a scratch buffer from the right, two digits per step. A uint64
	// has at most 20 decimal digits. The last step always writes a pair; a
	// leading zero is then skipped rather than branched around.
	var buf [20]byte
	i := len(buf)
	for u >= 100 {
		q := u / 100
		i -= 2
		binary.NativeEndian.PutUint16(buf[i:], digitPairs16[u-q*100])
		u = q
	}
	i -= 2
	binary.NativeEndian.PutUint16(buf[i:], digitPairs16[u])
	if u < 10 {
		i++
	}
	return append(b, buf[i:]...)
}
