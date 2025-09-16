// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file is a modified version of unicode/utf8.go from the Go standard library.
// Modifications are licensed under the MIT License.
//
// MIT License for modifications:
// Copyright (c) 2025 Matt Sherman

// Package utf8 implements functions and constants to support text encoded in
// UTF-8. It includes functions to translate between runes and UTF-8 byte sequences.
// See https://en.wikipedia.org/wiki/UTF-8
package utf8

import "github.com/clipperhouse/uax29/v2/internal/stringish"

// The conditions RuneError==unicode.ReplacementChar and
// MaxRune==unicode.MaxRune are verified in the tests.
// Defining them locally avoids this package depending on package unicode.

// Numbers fundamental to the encoding.
const (
	RuneError = '\uFFFD'     // the "error" Rune or "Unicode replacement character"
	RuneSelf  = 0x80         // characters below RuneSelf are represented as themselves in a single byte.
	MaxRune   = '\U0010FFFF' // Maximum valid Unicode code point.
	UTFMax    = 4            // maximum number of bytes of a UTF-8 encoded Unicode character.
)

// Code points in the surrogate range are not valid for UTF-8.
const (
	surrogateMin = 0xD800
	surrogateMax = 0xDFFF
)

const (
	t1 = 0b00000000
	tx = 0b10000000
	t2 = 0b11000000
	t3 = 0b11100000
	t4 = 0b11110000
	t5 = 0b11111000

	maskx = 0b00111111
	mask2 = 0b00011111
	mask3 = 0b00001111
	mask4 = 0b00000111

	rune1Max = 1<<7 - 1
	rune2Max = 1<<11 - 1
	rune3Max = 1<<16 - 1

	// The default lowest and highest continuation byte.
	locb = 0b10000000
	hicb = 0b10111111

	// These names of these constants are chosen to give nice alignment in the
	// table below. The first nibble is an index into acceptRanges or F for
	// special one-byte cases. The second nibble is the Rune length or the
	// Status for the special one-byte case.
	xx = 0xF1 // invalid: size 1
	as = 0xF0 // ASCII: size 1
	s1 = 0x02 // accept 0, size 2
	s2 = 0x13 // accept 1, size 3
	s3 = 0x03 // accept 0, size 3
	s4 = 0x23 // accept 2, size 3
	s5 = 0x34 // accept 3, size 4
	s6 = 0x04 // accept 0, size 4
	s7 = 0x44 // accept 4, size 4
)

const (
	runeErrorByte0 = t3 | (RuneError >> 12)
	runeErrorByte1 = tx | (RuneError>>6)&maskx
	runeErrorByte2 = tx | RuneError&maskx
)

// first is information about the first byte in a UTF-8 sequence.
var first = [256]uint8{
	//   1   2   3   4   5   6   7   8   9   A   B   C   D   E   F
	as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, // 0x00-0x0F
	as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, // 0x10-0x1F
	as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, // 0x20-0x2F
	as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, // 0x30-0x3F
	as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, // 0x40-0x4F
	as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, // 0x50-0x5F
	as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, // 0x60-0x6F
	as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, // 0x70-0x7F
	//   1   2   3   4   5   6   7   8   9   A   B   C   D   E   F
	xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, // 0x80-0x8F
	xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, // 0x90-0x9F
	xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, // 0xA0-0xAF
	xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, // 0xB0-0xBF
	xx, xx, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, // 0xC0-0xCF
	s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, // 0xD0-0xDF
	s2, s3, s3, s3, s3, s3, s3, s3, s3, s3, s3, s3, s3, s4, s3, s3, // 0xE0-0xEF
	s5, s6, s6, s6, s7, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, // 0xF0-0xFF
}

// acceptRange gives the range of valid values for the second byte in a UTF-8
// sequence.
type acceptRange struct {
	lo uint8 // lowest value for second byte.
	hi uint8 // highest value for second byte.
}

// acceptRanges has size 16 to avoid bounds checks in the code that uses it.
var acceptRanges = [16]acceptRange{
	0: {locb, hicb},
	1: {0xA0, hicb},
	2: {locb, 0x9F},
	3: {0x90, hicb},
	4: {locb, 0x8F},
}

// FullRune reports whether s begins with a full UTF-8 encoding of a rune.
// An invalid encoding is considered a full Rune since it will convert as a width-1 error rune.
func FullRune[T stringish.Interface](s T) bool {
	n := len(s)
	if n == 0 {
		return false
	}

	// Get the first byte - works for both []byte and string
	b0 := s[0]
	x := first[b0]
	if n >= int(x&7) {
		return true // ASCII, invalid or valid.
	}
	// Must be short or invalid.
	accept := acceptRanges[x>>4]
	if n > 1 {
		b1 := s[1]
		if b1 < accept.lo || accept.hi < b1 {
			return true
		}
	}
	if n > 2 {
		b2 := s[2]
		if b2 < locb || hicb < b2 {
			return true
		}
	}
	return false
}

// DecodeRune unpacks the first UTF-8 encoding in data and returns the rune and
// its width in bytes. If data is empty it returns ([RuneError], 0). Otherwise, if
// the encoding is invalid, it returns (RuneError, 1). Both are impossible
// results for correct, non-empty UTF-8.
//
// An encoding is invalid if it is incorrect UTF-8, encodes a rune that is
// out of range, or is not the shortest possible UTF-8 encoding for the
// value. No other validation is performed.
func DecodeRune[T stringish.Interface](s T) (r rune, size int) {
	n := len(s)
	if n < 1 {
		return RuneError, 0
	}

	// Get the first byte - works for both []byte and string
	b0 := s[0]
	x := first[b0]
	if x >= as {
		// The following code simulates an additional check for x == xx and
		// handling the ASCII and invalid cases accordingly. This mask-and-or
		// approach prevents an additional branch.
		mask := rune(x) << 31 >> 31 // Create 0x0000 or 0xFFFF.
		return rune(b0)&^mask | RuneError&mask, 1
	}
	sz := int(x & 7)
	accept := acceptRanges[x>>4]
	if n < sz {
		return RuneError, 1
	}

	// Get subsequent bytes as needed
	var b1, b2, b3 byte
	if sz > 1 {
		b1 = s[1]
	}
	if sz > 2 {
		b2 = s[2]
	}
	if sz > 3 {
		b3 = s[3]
	}

	if b1 < accept.lo || accept.hi < b1 {
		return RuneError, 1
	}
	if sz <= 2 { // <= instead of == to help the compiler eliminate some bounds checks
		return rune(b0&mask2)<<6 | rune(b1&maskx), 2
	}
	if b2 < locb || hicb < b2 {
		return RuneError, 1
	}
	if sz <= 3 {
		return rune(b0&mask3)<<12 | rune(b1&maskx)<<6 | rune(b2&maskx), 3
	}
	if b3 < locb || hicb < b3 {
		return RuneError, 1
	}
	return rune(b0&mask4)<<18 | rune(b1&maskx)<<12 | rune(b2&maskx)<<6 | rune(b3&maskx), 4
}

// DecodeLastRune unpacks the last UTF-8 encoding in data and returns the rune and
// its width in bytes. If data is empty it returns ([RuneError], 0). Otherwise, if
// the encoding is invalid, it returns (RuneError, 1). Both are impossible
// results for correct, non-empty UTF-8.
//
// An encoding is invalid if it is incorrect UTF-8, encodes a rune that is
// out of range, or is not the shortest possible UTF-8 encoding for the
// value. No other validation is performed.
func DecodeLastRune[T stringish.Interface](s T) (r rune, size int) {
	end := len(s)
	if end == 0 {
		return RuneError, 0
	}
	start := end - 1

	// Get the last byte - works for both []byte and string
	lastByte := s[start]
	r = rune(lastByte)
	if r < RuneSelf {
		return r, 1
	}
	// guard against O(n^2) behavior when traversing
	// backwards through strings with long sequences of
	// invalid UTF-8.
	lim := end - UTFMax
	if lim < 0 {
		lim = 0
	}
	for start--; start >= lim; start-- {
		b := s[start]
		if RuneStart(b) {
			break
		}
	}
	if start < 0 {
		start = 0
	}

	// Use the generic DecodeRune function
	r, size = DecodeRune(s[start:end])
	if start+size != end {
		return RuneError, 1
	}
	return r, size
}

// RuneLen returns the number of bytes in the UTF-8 encoding of the rune.
// It returns -1 if the rune is not a valid value to encode in UTF-8.
func RuneLen(r rune) int {
	switch {
	case r < 0:
		return -1
	case r <= rune1Max:
		return 1
	case r <= rune2Max:
		return 2
	case surrogateMin <= r && r <= surrogateMax:
		return -1
	case r <= rune3Max:
		return 3
	case r <= MaxRune:
		return 4
	}
	return -1
}

// EncodeRune writes into p (which must be large enough) the UTF-8 encoding of the rune.
// If the rune is out of range, it writes the encoding of [RuneError].
// It returns the number of bytes written.
func EncodeRune(p []byte, r rune) int {
	// This function is inlineable for fast handling of ASCII.
	if uint32(r) <= rune1Max {
		p[0] = byte(r)
		return 1
	}
	return encodeRuneNonASCII(p, r)
}

func encodeRuneNonASCII(p []byte, r rune) int {
	// Negative values are erroneous. Making it unsigned addresses the problem.
	switch i := uint32(r); {
	case i <= rune2Max:
		_ = p[1] // eliminate bounds checks
		p[0] = t2 | byte(r>>6)
		p[1] = tx | byte(r)&maskx
		return 2
	case i < surrogateMin, surrogateMax < i && i <= rune3Max:
		_ = p[2] // eliminate bounds checks
		p[0] = t3 | byte(r>>12)
		p[1] = tx | byte(r>>6)&maskx
		p[2] = tx | byte(r)&maskx
		return 3
	case i > rune3Max && i <= MaxRune:
		_ = p[3] // eliminate bounds checks
		p[0] = t4 | byte(r>>18)
		p[1] = tx | byte(r>>12)&maskx
		p[2] = tx | byte(r>>6)&maskx
		p[3] = tx | byte(r)&maskx
		return 4
	default:
		_ = p[2] // eliminate bounds checks
		p[0] = runeErrorByte0
		p[1] = runeErrorByte1
		p[2] = runeErrorByte2
		return 3
	}
}

// AppendRune appends the UTF-8 encoding of r to the end of p and
// returns the extended buffer. If the rune is out of range,
// it appends the encoding of [RuneError].
func AppendRune(p []byte, r rune) []byte {
	// This function is inlineable for fast handling of ASCII.
	if uint32(r) <= rune1Max {
		return append(p, byte(r))
	}
	return appendRuneNonASCII(p, r)
}

func appendRuneNonASCII(p []byte, r rune) []byte {
	// Negative values are erroneous. Making it unsigned addresses the problem.
	switch i := uint32(r); {
	case i <= rune2Max:
		return append(p, t2|byte(r>>6), tx|byte(r)&maskx)
	case i < surrogateMin, surrogateMax < i && i <= rune3Max:
		return append(p, t3|byte(r>>12), tx|byte(r>>6)&maskx, tx|byte(r)&maskx)
	case i > rune3Max && i <= MaxRune:
		return append(p, t4|byte(r>>18), tx|byte(r>>12)&maskx, tx|byte(r>>6)&maskx, tx|byte(r)&maskx)
	default:
		return append(p, runeErrorByte0, runeErrorByte1, runeErrorByte2)
	}
}

// RuneCount returns the number of runes in s. Erroneous and short
// encodings are treated as single runes of width 1 byte.
func RuneCount[T stringish.Interface](s T) int {
	n := len(s)
	var count int
	for i := 0; i < n; i++ {
		c := s[i]
		if c >= RuneSelf {
			// non-ASCII slow path - use the generic DecodeRune function
			_, size := DecodeRune(s[i:])
			if size > 0 {
				count++
				i += size - 1 // -1 because the loop will increment
			} else {
				count++
			}
		} else {
			count++
		}
	}
	return count
}

// RuneStart reports whether the byte could be the first byte of an encoded,
// possibly invalid rune. Second and subsequent bytes always have the top two
// bits set to 10.
func RuneStart(b byte) bool { return b&0xC0 != 0x80 }

// Valid reports whether s consists entirely of valid UTF-8-encoded runes.
func Valid[T stringish.Interface](s T) bool {
	n := len(s)
	if n == 0 {
		return true
	}

	// Fast path. Check for and skip 8 bytes of ASCII characters per iteration.
	for i := 0; i+8 <= n; i += 8 {
		// Get 8 bytes - works for both []byte and string
		var bytes [8]byte
		copy(bytes[:], s[i:i+8])

		// Combining two 32 bit loads allows the same code to be used
		// for 32 and 64 bit platforms.
		// The compiler can generate a 32bit load for first32 and second32
		// on many platforms. See test/codegen/memcombine.go.
		first32 := uint32(bytes[0]) | uint32(bytes[1])<<8 | uint32(bytes[2])<<16 | uint32(bytes[3])<<24
		second32 := uint32(bytes[4]) | uint32(bytes[5])<<8 | uint32(bytes[6])<<16 | uint32(bytes[7])<<24
		if (first32|second32)&0x80808080 != 0 {
			// Found a non ASCII byte (>= RuneSelf).
			// Process remaining bytes individually
			for j := i; j < n; {
				b := s[j]
				if b < RuneSelf {
					j++
					continue
				}
				x := first[b]
				if x == xx {
					return false // Illegal starter byte.
				}
				size := int(x & 7)
				if j+size > n {
					return false // Short or invalid.
				}
				accept := acceptRanges[x>>4]
				if j+1 < n {
					c := s[j+1]
					if c < accept.lo || accept.hi < c {
						return false
					}
				}
				if size >= 3 && j+2 < n {
					c := s[j+2]
					if c < locb || hicb < c {
						return false
					}
				}
				if size == 4 && j+3 < n {
					c := s[j+3]
					if c < locb || hicb < c {
						return false
					}
				}
				j += size
			}
			return true
		}
	}

	// Process remaining bytes individually
	for i := 0; i < n; {
		b := s[i]
		if b < RuneSelf {
			i++
			continue
		}
		x := first[b]
		if x == xx {
			return false // Illegal starter byte.
		}
		size := int(x & 7)
		if i+size > n {
			return false // Short or invalid.
		}
		accept := acceptRanges[x>>4]
		if i+1 < n {
			c := s[i+1]
			if c < accept.lo || accept.hi < c {
				return false
			}
		}
		if size >= 3 && i+2 < n {
			c := s[i+2]
			if c < locb || hicb < c {
				return false
			}
		}
		if size == 4 && i+3 < n {
			c := s[i+3]
			if c < locb || hicb < c {
				return false
			}
		}
		i += size
	}
	return true
}

// ValidRune reports whether r can be legally encoded as UTF-8.
// Code points that are out of range or a surrogate half are illegal.
func ValidRune(r rune) bool {
	switch {
	case 0 <= r && r < surrogateMin:
		return true
	case surrogateMax < r && r <= MaxRune:
		return true
	}
	return false
}
