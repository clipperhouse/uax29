package util

import (
	"unicode"
	"unicode/utf8"
)

// Contains indicates whether the token contains any runes that are in
// one or more of the given ranges. If the token is empty, or no ranges
// are given, it will return false.
func Contains(token []byte, ranges ...*unicode.RangeTable) bool {
	if len(token) == 0 || len(ranges) == 0 {
		return false
	}

	pos := 0
	for pos < len(token) {
		r, w := utf8.DecodeRune(token[pos:])
		if unicode.In(r, ranges...) {
			return true
		}
		pos += w
	}

	return false
}

// Entirely indicates whether the token consists entirely of runes
// that are in one or more of the given ranges. If the token is empty,
// or no ranges are given, it will return false.
func Entirely(token []byte, ranges ...*unicode.RangeTable) bool {
	if len(token) == 0 || len(ranges) == 0 {
		return false
	}

	pos := 0
	for pos < len(token) {
		r, w := utf8.DecodeRune(token[pos:])
		if !unicode.In(r, ranges...) {
			return false
		}
		pos += w
	}

	return true
}
