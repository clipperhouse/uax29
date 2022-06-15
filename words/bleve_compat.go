package words

import (
	"unicode"

	"github.com/clipperhouse/uax29/iterators/filter"
	"golang.org/x/text/unicode/rangetable"
)

// BleveNumeric determines if a token is Numeric using the Bleve segmenter's.
// definition, see: https://github.com/blevesearch/segment/blob/master/segment_words.rl#L199-L207
// This API is experimental.
func BleveNumeric(token []byte) bool {
	var pos, w int
	var current property

	for pos < len(token) {
		last := current

		current, w = trie.lookup(token[pos:])

		if pos == 0 {
			if current.is(_Numeric | _ExtendNumLet) {
				pos += w
				continue
			} else {
				// definitely not Numeric, can move on
				return false
			}
		}

		// WB8.   Numeric × Numeric
		// https://unicode.org/reports/tr29/#WB8
		isWB8 := last.is(_Numeric) && current.is(_Numeric)
		if isWB8 {
			pos += w
			continue
		}

		// WB11.  Numeric (MidNum | MidNumLet | Single_Quote) × Numeric
		// https://unicode.org/reports/tr29/#WB11
		// Determine if WB11 can possibly apply
		maybeWB11 := last.is(_MidNum|_MidNumLetQ|_Ignore) && current.is(_Numeric)
		if maybeWB11 {
			i := previousIndex(_MidNum|_MidNumLetQ, token[:pos])
			if i > 0 && previous(_Numeric, token[:i]) {
				pos += w
				continue
			}
		}

		// WB12.  Numeric × (MidNum | MidNumLet | Single_Quote) Numeric
		// https://unicode.org/reports/tr29/#WB12
		// Optimization: determine if WB12 can possibly apply
		maybeWB12 := last.is(_Numeric|_Ignore) && current.is(_MidNum|_MidNumLetQ)
		if maybeWB12 {
			if subsequent(_Numeric, token[pos+w:]) && previous(_Numeric, token[:pos]) {
				pos += w
				continue
			}
		}

		// WB13a. (ALetter | Hebrew_Letter | Numeric | Katakana | ExtendNumLet) × ExtendNumLet
		// https://unicode.org/reports/tr29/#WB13a
		// Determine if WB13a can possibly apply
		maybeWB13a := last.is(_Numeric|_ExtendNumLet|_Ignore) && current.is(_ExtendNumLet)
		if maybeWB13a {
			if previous(_Numeric|_ExtendNumLet, token[:pos]) {
				pos += w
				continue
			}
		}

		// WB13b. ExtendNumLet × (ALetter | Hebrew_Letter | Numeric | Katakana)
		// https://unicode.org/reports/tr29/#WB13b
		// Determine if WB13b can possibly apply
		maybeWB13b := last.is(_ExtendNumLet|_Ignore) && current.is(_Numeric)
		if maybeWB13b {
			if previous(_ExtendNumLet, token[:pos]) {
				pos += w
				continue
			}
		}

		// if we get here, none of the above rules apply
		return false
	}

	return true
}

var ideoRange = rangetable.Merge(unicode.Ideographic, unicode.Katakana, unicode.Hiragana)
var ideographic = filter.Entirely(ideoRange) // would filter.Contains be better?

// BleveIdeographic determines if a token is comprised entirely of ideographs, by the
// Bleve segmenter's definition. It adds Katakana & Hiragana to unicode.Ideographic.
// This API is experimental.
func BleveIdeographic(token []byte) bool {
	return ideographic(token)
}

// On the complex topic of CJK & Unicode:
//  https://www.hieuthi.com/blog/2021/07/22/unicode-categories-cjk-ideographs.html
