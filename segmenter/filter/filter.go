package filter

import (
	"unicode"
	"unicode/utf8"
)

type Func func([]byte) bool

// Contains returns a filter indicating that a segment (token) contains one
// or more runes that are in one or more of the ranges. Examples of ranges
// are things like unicode.Letter, unicode.White_Space, or unicode.Title,
// allowing testing for a wide variety of character or script types.
var Contains = func(ranges ...*unicode.RangeTable) Func {
	return func(token []byte) bool {
		return contains(token, ranges...)
	}
}

// Entirely returns a filter indicating that a segment (token) consists
// entirely of runes that are in one or more of the ranges. Examples of ranges
// are things like unicode.Letter, unicode.White_Space, or unicode.Title,
// allowing testing for a wide variety of character or script types.
var Entirely = func(ranges ...*unicode.RangeTable) Func {
	return func(token []byte) bool {
		return entirely(token, ranges...)
	}
}

func contains(token []byte, ranges ...*unicode.RangeTable) bool {
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

func entirely(token []byte, ranges ...*unicode.RangeTable) bool {
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

var IsWordlikeSlower = Contains(unicode.Letter, unicode.Number, unicode.Symbol)

// IsWordlike is a filter which return only tokens (segments) are "words" in
// the common sense, ignoring tokens that are whitespace and punctuation. It
// includes any token that contains a Letter, Number, or Symbol, as defined
// by Unicode. To use it, call segmenter.Filter(IsWordlike).
var IsWordlike = func(token []byte) bool {
	// Hotpath version, faster than Contains
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
