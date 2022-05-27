// Package transformer provides a few handy transformers, for use with Scanner and Segmenter. See https://pkg.go.dev/golang.org/x/text
// for lots of others.
package transformer

import (
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Lower transforms text to lowercase
var Lower transform.Transformer = cases.Lower(language.Und)

// Upper transforms text to uppercase
var Upper transform.Transformer = cases.Upper(language.Und)

// Diacritics 'flattens' characters with diacritics, such as accents. For example,
// açaí → acai. (It has the side effect of normalizing to NFC form, which should be fine.)
var Diacritics transform.Transformer = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC) // https://stackoverflow.com/q/24588295
