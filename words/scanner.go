// Package words provides a scanner for Unicode text segmentation word boundaries: https://unicode.org/reports/tr29/#Word_Boundaries
package words

import (
	"io"
	"unicode"

	"github.com/clipperhouse/uax29"
	"github.com/clipperhouse/uax29/emoji"
)

// NewScanner tokenizes a reader into a stream of tokens according to Unicode Text Segmentation word boundaries https://unicode.org/reports/tr29/#Word_Boundaries
// Iterate through the stream by calling Scan() until false.
//	text := "This is an example."
//	reader := strings.NewReader(text)
//
//	scanner := words.NewScanner(reader)
//	for scanner.Scan() {
//		fmt.Printf("%q\n", scanner.Text())
//	}
//	if err := scanner.Err(); err != nil {
//		log.Fatal(err)
//	}
func NewScanner(r io.Reader) *uax29.Scanner {
	return uax29.NewScanner(r, BreakFunc)
}

var is = unicode.Is

// BreakFunc implements word boundaries according to https://unicode.org/reports/tr29/#Word_Boundaries.
// It is intended for use with uax29.Scanner.
var BreakFunc uax29.BreakFunc = func(buffer uax29.Runes, pos uax29.Pos) bool {
	// Rules: https://unicode.org/reports/tr29/#Word_Boundary_Rules

	sot := pos == 0 // "start of text"
	eof := len(buffer) == int(pos)

	// WB1
	if sot && !eof {
		return false
	}

	// WB2
	if eof {
		return true
	}

	current := buffer[pos]
	previous := buffer[pos-1]

	// WB3
	if is(LF, current) && is(CR, previous) {
		return false
	}

	// WB3a
	if is(_CRǀLFǀNewline, previous) {
		return true
	}

	// WB3b
	if is(_CRǀLFǀNewline, current) {
		return true
	}

	// WB3c
	if is(emoji.Extended_Pictographic, current) && is(ZWJ, previous) {
		return false
	}

	// WB3d
	if is(WSegSpace, current) && is(WSegSpace, previous) {
		return false
	}

	// WB4
	if is(_ExtendǀFormatǀZWJ, current) {
		return false
	}

	ignore := _ExtendǀFormatǀZWJ
	// WB5
	if is(AHLetter, current) && buffer.SeekPrevious(pos, ignore, AHLetter) {
		return false
	}

	// WB6
	if is(_MidLetterǀMidNumLetQ, current) && buffer.SeekForward(pos, ignore, AHLetter) && buffer.SeekPrevious(pos, ignore, AHLetter) {
		return false
	}

	// WB7
	if is(AHLetter, current) {
		previousIndex := buffer.SeekPreviousIndex(pos, ignore, _MidLetterǀMidNumLetQ)
		if previousIndex >= 0 && buffer.SeekPrevious(previousIndex, ignore, AHLetter) {
			return false
		}
	}

	// WB7a
	if is(Single_Quote, current) && buffer.SeekPrevious(pos, ignore, Hebrew_Letter) {
		return false
	}

	// WB7b
	if is(Double_Quote, current) && buffer.SeekForward(pos, ignore, Hebrew_Letter) &&
		buffer.SeekPrevious(pos, ignore, Hebrew_Letter) {
		return false
	}

	// WB7c
	if is(Hebrew_Letter, current) {
		previousIndex := buffer.SeekPreviousIndex(pos, ignore, Double_Quote)
		if previousIndex >= 0 && buffer.SeekPrevious(previousIndex, ignore, Hebrew_Letter) {
			return false
		}
	}

	// WB8
	if is(Numeric, current) && buffer.SeekPrevious(pos, ignore, Numeric) {
		return false
	}

	// WB9
	if is(Numeric, current) && buffer.SeekPrevious(pos, ignore, AHLetter) {
		return false
	}

	// WB10
	if is(AHLetter, current) && buffer.SeekPrevious(pos, ignore, Numeric) {
		return false
	}

	// WB11
	if is(Numeric, current) {
		previousIndex := buffer.SeekPreviousIndex(pos, ignore, _MidNumǀMidNumLetQ)
		if previousIndex >= 0 && buffer.SeekPrevious(previousIndex, ignore, Numeric) {
			return false
		}
	}

	// WB12
	if is(_MidNumǀMidNumLetQ, current) && buffer.SeekForward(pos, ignore, Numeric) &&
		buffer.SeekPrevious(pos, ignore, Numeric) {
		return false
	}

	// WB13
	if is(Katakana, current) && buffer.SeekPrevious(pos, ignore, Katakana) {
		return false
	}

	// WB13a
	if is(ExtendNumLet, current) && buffer.SeekPrevious(pos, ignore, _AHLetterǀNumericǀKatakanaǀExtendNumLet) {
		return false
	}

	// WB13b
	if is(_AHLetterǀNumericǀKatakana, current) && buffer.SeekPrevious(pos, ignore, ExtendNumLet) {
		return false
	}

	// WB15
	if is(Regional_Indicator, current) {
		ok := true

		// Buffer comprised entirely of an odd number of RI, ignoring Extend|Format|ZWJ
		count := 0
		for i := pos - 1; i >= 0; i-- {
			r := buffer[i]
			if is(_ExtendǀFormatǀZWJ, r) {
				continue
			}
			if !is(Regional_Indicator, r) {
				ok = false
				break
			}
			count++
		}

		// If we fall through, we've seen the whole buffer,
		// so it's all Regional_Indicator | Extend|Format|ZWJ
		if ok {
			odd := count > 0 && count%2 == 1
			if odd {
				return false
			}
		}
	}

	// WB16
	if is(Regional_Indicator, current) {
		ok := false
		// Last n runes represent an odd number of RI, ignoring Extend|Format|ZWJ
		count := 0
		for i := pos - 1; i >= 0; i-- {
			r := buffer[i]
			if is(_ExtendǀFormatǀZWJ, r) {
				continue
			}
			if !is(Regional_Indicator, r) {
				odd := count > 0 && count%2 == 1
				ok = odd
				break
			}
			count++
		}

		if ok {
			return false
		}
	}

	// WB999
	// If we fall through all the above rules, it's a word break
	return true
}
