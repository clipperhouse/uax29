package filter

import (
	"unicode"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/iterators/util"
)

type Predicate func([]byte) bool

// Contains returns a filter (predicate) indicating that a segment (token) contains one
// or more runes that are in one or more of the given ranges. Examples of ranges
// are things like unicode.Letter, unicode.Arabic, or unicode.Lower,
// allowing testing for a wide variety of character or script types.
//
// Intended for use with segmenter.Filter or scanner.Filter.
//
// If the given token is empty, or no ranges are given, it will return false.
func Contains(ranges ...*unicode.RangeTable) Predicate {
	return func(token []byte) bool {
		return util.Contains(token, ranges...)
	}
}

// Entirely returns a filter (predicate) indicating that a segment (token)
// consists entirely of runes that are in one or more of the given ranges.
// Examples of ranges are things like unicode.Letter, unicode.Arabic,
// or unicode.Lower, allowing testing for a wide variety of character
// or script types.
//
// Intended for use with segmenter.Filter or scanner.Filter.
//
// If the given token is empty, or no ranges are given, it will return false.
func Entirely(ranges ...*unicode.RangeTable) Predicate {
	return func(token []byte) bool {
		return util.Entirely(token, ranges...)
	}
}

// Wordlike is a filter which returns only tokens (segments) that are "words"
// in the common sense, excluding tokens that are whitespace or punctuation.
// It includes any token that contains a Letter, Number, or Symbol, as defined
// by Unicode. To use it, call Filter(Wordlike) on a Segmenter or Scanner.
var Wordlike Predicate = func(token []byte) bool {
	// Hotpath version, faster than using Contains with Rangetables
	pos := 0
	for pos < len(token) {
		r, w := utf8.DecodeRune(token[pos:])
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSymbol(r) {
			return true
		}
		pos += w
	}

	return false
}
