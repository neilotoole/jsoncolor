package jsoncolor

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

	// Fill a scratch buffer from the right, two digits at a time. A uint64
	// has at most 20 decimal digits.
	var buf [20]byte
	i := len(buf)
	for u >= 100 {
		q := u / 100
		j := 2 * (u - q*100)
		i -= 2
		buf[i] = digitPairs[j]
		buf[i+1] = digitPairs[j+1]
		u = q
	}
	if u >= 10 {
		i -= 2
		buf[i] = digitPairs[2*u]
		buf[i+1] = digitPairs[2*u+1]
	} else {
		i--
		buf[i] = byte('0' + u)
	}
	return append(b, buf[i:]...)
}
