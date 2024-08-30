// Package transformer provides a few handy transformers, for use with Scanner and Segmenter.
//
// We use the golang.org/x/text/transform package. We can accept anything that conforms to the transform.Transformer interface.
package transformer

import (
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type lower struct{}

func (l lower) Transform(dst []byte, src []byte, atEOF bool) (nDst int, nSrc int, err error) {
	c := cases.Lower(language.Und)
	return c.Transform(dst, src, atEOF)
}
func (l lower) Reset() {
	// no-op for our purposes
}

type upper struct{}

func (u upper) Transform(dst []byte, src []byte, atEOF bool) (nDst int, nSrc int, err error) {
	c := cases.Upper(language.Und)
	return c.Transform(dst, src, atEOF)
}
func (l upper) Reset() {
	// no-op for our purposes
}

type diacritics struct{}

func (d diacritics) Transform(dst []byte, src []byte, atEOF bool) (nDst int, nSrc int, err error) {
	c := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC) // https://stackoverflow.com/q/24588295
	return c.Transform(dst, src, atEOF)
}
func (l diacritics) Reset() {
	// no-op for our purposes
}

// Lower transforms text to lowercase
var Lower transform.Transformer = lower{}

// Upper transforms text to uppercase
var Upper transform.Transformer = upper{}

// Diacritics 'flattens' characters with diacritics, such as accents. For example,
// açaí → acai. (It has the side effect of normalizing to NFC form, which should be fine.)
var Diacritics transform.Transformer = diacritics{}
