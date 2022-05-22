package filter

import (
	"unicode"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/segmenter/util"
)

type Func func([]byte) bool

// Contains returns a filter indicating that a segment (token) contains one
// or more runes that are in one or more of the ranges. Examples of ranges
// are things like unicode.Letter, unicode.Arabic, or unicode.Lower,
// allowing testing for a wide variety of character or script types.
var Contains = func(ranges ...*unicode.RangeTable) Func {
	return func(token []byte) bool {
		return util.Contains(token, ranges...)
	}
}

// Entirely returns a filter indicating that a segment (token) consists
// entirely of runes that are in one or more of the ranges. Examples of ranges
// are things like unicode.Letter, unicode.Arabic, or unicode.Lower,
// allowing testing for a wide variety of character or script types.
var Entirely = func(ranges ...*unicode.RangeTable) Func {
	return func(token []byte) bool {
		return util.Entirely(token, ranges...)
	}
}

// Wordlike is a filter which return only tokens (segments) are "words" in
// the common sense, ignoring tokens that are whitespace and punctuation. It
// includes any token that contains a Letter, Number, or Symbol, as defined
// by Unicode. To use it, call segmenter.Filter(Wordlike).
var Wordlike = func(token []byte) bool {
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
