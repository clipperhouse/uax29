package words

import (
	"unicode"

	"github.com/clipperhouse/uax29/iterators/filter"
	"golang.org/x/text/unicode/rangetable"
)

// Numeric determines if a token is Numeric, in the narrow sense of being
// attributable to the WB11/12 rules in https://unicode.org/reports/tr29/#WB11.
// It does not attempt to determine validity (parseability) of the number; for
// that, use something like https://pkg.go.dev/strconv#ParseFloat.
func Numeric(token []byte) bool {
	// is the token "attributable" to WB11/12?

	var pos, w int
	var current property

	for pos < len(token) {
		last := current

		current, w = trie.lookup(token[pos:])

		if pos == 0 && current.is(_Numeric) {
			pos += w
			continue
		}

		// Optimization: determine if WB11 can possibly apply
		maybeWB11 := current.is(_Numeric) && last.is(_MidNum|_MidNumLetQ|_Ignore)

		// https://unicode.org/reports/tr29/#WB11
		if maybeWB11 {
			i := previousIndex(_MidNum|_MidNumLetQ, token[:pos])
			if i > 0 && previous(_Numeric, token[:i]) {
				pos += w
				continue
			}
		}

		// Optimization: determine if WB12 can possibly apply
		maybeWB12 := current.is(_MidNum|_MidNumLetQ) && last.is(_Numeric|_Ignore)

		// https://unicode.org/reports/tr29/#WB12
		if maybeWB12 {
			if subsequent(_Numeric, token[pos+w:]) && previous(_Numeric, token[:pos]) {
				pos += w
				continue
			}
		}

		// if we get here, it's something other than WB11 or WB12
		return false
	}

	return true
}

var ideoRange = rangetable.Merge(unicode.Ideographic, unicode.Katakana, unicode.Hiragana)
var ideographic = filter.Entirely(ideoRange) // would filter.Contains be better?

// Ideographic determines if a token is comprised entirely of ideographs, broadly defined.
// In particular, it adds Katakana & Hiragana, which is Bleve's definition.
func Ideographic(token []byte) bool {
	return ideographic(token)
}

// On the complex topic of CJK & Unicode:
//  https://www.hieuthi.com/blog/2021/07/22/unicode-categories-cjk-ideographs.html
