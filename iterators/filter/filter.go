// Package filter provides methods for filtering via Scanners and Segmenters. A filter is
// defined as a func(text []byte) bool -- given a string, what is true about it?
//
// A filter can contain arbitrary logic. A common use of a filter is to determine
// what Unicode categories a string belongs to.
package filter

import (
	"unicode"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/iterators/util"
	"golang.org/x/text/unicode/rangetable"
)

type Func func([]byte) bool
type Predicate = Func // for backwards compat, this was renamed

// Contains returns a filter indicating that a token contains one
// or more runes that are in one or more of the given ranges.
// Examples of ranges are things like unicode.Letter, unicode.Arabic,
// or unicode.Lower, allowing testing for a variety of character
// or script types.
func Contains(ranges ...*unicode.RangeTable) Func {
	merged := rangetable.Merge(ranges...)
	return func(token []byte) bool {
		return util.Contains(token, merged)
	}
}

// Entirely returns a filter indicating that a token consists
// entirely of runes that are in one or more of the given ranges.
// Examples of ranges are things like unicode.Letter, unicode.Arabic,
// or unicode.Lower, allowing testing for a variety of character
// or script types.
func Entirely(ranges ...*unicode.RangeTable) Func {
	merged := rangetable.Merge(ranges...)
	return func(token []byte) bool {
		return util.Entirely(token, merged)
	}
}

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

// Wordlike is a filter which returns only tokens that contain
// a Letter, Number, or Symbol, as defined by Unicode.
var Wordlike Func = func(token []byte) bool {
	pos := 0
	for pos < len(token) {
		r, w := utf8.DecodeRune(token[pos:])
		// we use these methods instead of unicode.In for
		// performance; these methods have ASCII fast paths
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSymbol(r) {
			return true
		}
		pos += w
	}

	return false
}
