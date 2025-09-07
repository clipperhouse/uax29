// Package filter provides methods for filtering via Scanners and Segmenters. A filter is
// defined as a func(text []byte) bool -- given a string, what is true about it?
//
// A filter can contain arbitrary logic. A common use of a filter is to determine
// what Unicode categories a string belongs to.
package filter

import (
	"unicode"
	"unicode/utf8"
)

type Func func([]byte) bool

// AlphaNumeric is a filter which returns only tokens
// that contain a Letter or Number, as defined by Unicode.
var AlphaNumeric Func = func(token []byte) bool {
	pos := 0
	for pos < len(token) {
		r, w := utf8.DecodeRune(token[pos:])
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			// we use these methods instead of unicode.In for
			// performance; these methods have ASCII fast paths
			return true
		}
		pos += w
	}

	return false
}
