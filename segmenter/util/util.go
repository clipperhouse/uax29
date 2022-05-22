package util

import (
	"unicode"
	"unicode/utf8"
)

func Contains(token []byte, ranges ...*unicode.RangeTable) bool {
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

func Entirely(token []byte, ranges ...*unicode.RangeTable) bool {
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
