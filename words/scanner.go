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
func BreakFunc(buffer uax29.Runes, pos uax29.Pos) bool {
	// Rules: https://unicode.org/reports/tr29/#Word_Boundary_Rules

	sot := pos == 0 // "start of text"
	eof := len(buffer) == int(pos)

	// https://unicode.org/reports/tr29/#WB1
	if sot && !eof {
		return uax29.Accept
	}

	// https://unicode.org/reports/tr29/#WB2
	if eof {
		return uax29.Break
	}

	current := buffer[pos]
	previous := buffer[pos-1]

	// https://unicode.org/reports/tr29/#WB3
	if is(LF, current) && is(CR, previous) {
		return uax29.Accept
	}

	// https://unicode.org/reports/tr29/#WB3a
	if is(_CRǀLFǀNewline, previous) {
		return uax29.Break
	}

	// https://unicode.org/reports/tr29/#WB3b
	if is(_CRǀLFǀNewline, current) {
		return uax29.Break
	}

	// https://unicode.org/reports/tr29/#WB3c
	if is(emoji.Extended_Pictographic, current) && is(ZWJ, previous) {
		return uax29.Accept
	}

	// https://unicode.org/reports/tr29/#WB3d
	if is(WSegSpace, current) && is(WSegSpace, previous) {
		return uax29.Accept
	}

	// https://unicode.org/reports/tr29/#WB4
	if is(_ExtendǀFormatǀZWJ, current) {
		return uax29.Accept
	}

	// WB4 applies to subsequent rules; there is an implied "ignoring Extend & Format & ZWJ"
	// https://unicode.org/reports/tr29/#Grapheme_Cluster_and_Format_Rules
	// The Seek* methods are shorthand for "seek a category but skip over Extend|Format|ZWJ on the way"

	ignore := _ExtendǀFormatǀZWJ

	// https://unicode.org/reports/tr29/#WB5
	if is(AHLetter, current) && buffer.SeekPrevious(pos, ignore, AHLetter) {
		return uax29.Accept
	}

	// https://unicode.org/reports/tr29/#WB6
	if is(_MidLetterǀMidNumLetQ, current) && buffer.SeekForward(pos, ignore, AHLetter) && buffer.SeekPrevious(pos, ignore, AHLetter) {
		return uax29.Accept
	}

	// https://unicode.org/reports/tr29/#WB7
	if is(AHLetter, current) {
		previousIndex := buffer.SeekPreviousIndex(pos, ignore, _MidLetterǀMidNumLetQ)
		if previousIndex >= 0 && buffer.SeekPrevious(previousIndex, ignore, AHLetter) {
			return uax29.Accept
		}
	}

	// https://unicode.org/reports/tr29/#WB7a
	if is(Single_Quote, current) && buffer.SeekPrevious(pos, ignore, Hebrew_Letter) {
		return uax29.Accept
	}

	// https://unicode.org/reports/tr29/#WB7b
	if is(Double_Quote, current) && buffer.SeekForward(pos, ignore, Hebrew_Letter) &&
		buffer.SeekPrevious(pos, ignore, Hebrew_Letter) {
		return uax29.Accept
	}

	// https://unicode.org/reports/tr29/#WB7c
	if is(Hebrew_Letter, current) {
		previousIndex := buffer.SeekPreviousIndex(pos, ignore, Double_Quote)
		if previousIndex >= 0 && buffer.SeekPrevious(previousIndex, ignore, Hebrew_Letter) {
			return uax29.Accept
		}
	}

	// https://unicode.org/reports/tr29/#WB8
	if is(Numeric, current) && buffer.SeekPrevious(pos, ignore, Numeric) {
		return uax29.Accept
	}

	// https://unicode.org/reports/tr29/#WB9
	if is(Numeric, current) && buffer.SeekPrevious(pos, ignore, AHLetter) {
		return uax29.Accept
	}

	// https://unicode.org/reports/tr29/#WB10
	if is(AHLetter, current) && buffer.SeekPrevious(pos, ignore, Numeric) {
		return uax29.Accept
	}

	// https://unicode.org/reports/tr29/#WB11
	if is(Numeric, current) {
		previousIndex := buffer.SeekPreviousIndex(pos, ignore, _MidNumǀMidNumLetQ)
		if previousIndex >= 0 && buffer.SeekPrevious(previousIndex, ignore, Numeric) {
			return uax29.Accept
		}
	}

	// https://unicode.org/reports/tr29/#WB12
	if is(_MidNumǀMidNumLetQ, current) && buffer.SeekForward(pos, ignore, Numeric) &&
		buffer.SeekPrevious(pos, ignore, Numeric) {
		return uax29.Accept
	}

	// https://unicode.org/reports/tr29/#WB13
	if is(Katakana, current) && buffer.SeekPrevious(pos, ignore, Katakana) {
		return uax29.Accept
	}

	// https://unicode.org/reports/tr29/#WB13a
	if is(ExtendNumLet, current) && buffer.SeekPrevious(pos, ignore, _AHLetterǀNumericǀKatakanaǀExtendNumLet) {
		return uax29.Accept
	}

	// https://unicode.org/reports/tr29/#WB13b
	if is(_AHLetterǀNumericǀKatakana, current) && buffer.SeekPrevious(pos, ignore, ExtendNumLet) {
		return uax29.Accept
	}

	// https://unicode.org/reports/tr29/#WB15
	if is(Regional_Indicator, current) {
		allRI := true

		// Buffer comprised entirely of an odd number of RI, ignoring Extend|Format|ZWJ
		count := 0
		for i := pos - 1; i >= 0; i-- {
			r := buffer[i]
			if is(ignore, r) {
				continue
			}
			if !is(Regional_Indicator, r) {
				allRI = false
				break
			}
			count++
		}

		odd := count > 0 && count%2 == 1

		if allRI && odd {
			return uax29.Accept
		}
	}

	// https://unicode.org/reports/tr29/#WB16
	if is(Regional_Indicator, current) {
		odd := false
		// Last n runes represent an odd number of RI, ignoring Extend|Format|ZWJ
		count := 0
		for i := pos - 1; i >= 0; i-- {
			r := buffer[i]
			if is(ignore, r) {
				continue
			}
			if !is(Regional_Indicator, r) {
				odd = count > 0 && count%2 == 1
				break
			}
			count++
		}

		if odd {
			return uax29.Accept
		}
	}

	// https://unicode.org/reports/tr29/#WB999
	// If we fall through all the above rules, it's a word break
	return uax29.Break
}

// SplitFunc is a bufio.SplitFunc implementation of word segmentation, for use with bufio.Scanner
func SplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	return splitFunc(data, atEOF)
}

var splitFunc = uax29.NewSplitFunc(BreakFunc)
