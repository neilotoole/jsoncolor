package jsoncolor

import (
	"math"
	"math/rand"
	"strconv"
	"testing"
)

// TestAppendInt verifies that appendInt and appendUint produce exactly what
// strconv produces, across small values, every power of ten and its
// neighbours, the extremes of both types, and random values.
func TestAppendInt(t *testing.T) {
	var ints []int64
	var uints []uint64

	for i := int64(-1100); i <= 1100; i++ {
		ints = append(ints, i)
	}
	for i := uint64(0); i <= 1100; i++ {
		uints = append(uints, i)
	}
	p := uint64(1)
	for range 20 {
		for _, d := range []int64{-1, 0, 1} {
			uints = append(uints, p+uint64(d))
			ints = append(ints, int64(p)+d, -int64(p)+d)
		}
		p *= 10
	}
	ints = append(ints, math.MinInt64, math.MinInt64+1, math.MaxInt64, math.MaxInt64-1)
	uints = append(uints, math.MaxUint64, math.MaxUint64-1, math.MaxInt64, math.MaxInt64+1)

	rng := rand.New(rand.NewSource(1))
	for range 10000 {
		ints = append(ints, rng.Int63()>>uint(rng.Intn(63)), -(rng.Int63() >> uint(rng.Intn(63))))
		uints = append(uints, rng.Uint64()>>uint(rng.Intn(64)))
	}

	prefix := []byte("x")
	for _, n := range ints {
		want := strconv.AppendInt(prefix, n, 10)
		got := appendInt(prefix, n)
		if string(want) != string(got) {
			t.Fatalf("appendInt(%d) = %q, want %q", n, got, want)
		}
	}
	for _, u := range uints {
		want := strconv.AppendUint(prefix, u, 10)
		got := appendUint(prefix, u)
		if string(want) != string(got) {
			t.Fatalf("appendUint(%d) = %q, want %q", u, got, want)
		}
	}
}
