// Package words provides a scanner for Unicode text segmentation word boundaries: https://unicode.org/reports/tr29/#Word_Boundaries
package words

import (
	"bufio"
	"io"
	"unicode"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/emoji"
	"github.com/clipperhouse/uax29/seek"
)

// NewScanner tokenizes a reader into a stream of tokens according to Unicode Text Segmentation word boundaries https://unicode.org/reports/tr29/#Word_Boundaries.
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
func NewScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	scanner.Split(SplitFunc)
	return scanner
}

var is = unicode.Is

// SplitFunc is a bufio.SplitFunc implementation of word segmentation, for use with bufio.Scanner
func SplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	pos := 0

	for {
		sot := pos == 0 // "start of text"
		eof := len(data) == pos && atEOF

		// https://unicode.org/reports/tr29/#WB1
		if sot && !eof {
			_, w := utf8.DecodeRune(data[pos:])
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB2
		if eof {
			break
		}

		current, w := utf8.DecodeRune(data[pos:])
		previous, _ := utf8.DecodeLastRune(data[:pos])

		// https://unicode.org/reports/tr29/#WB3
		if is(LF, current) && is(CR, previous) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB3a
		if is(_CRǀLFǀNewline, previous) {
			break
		}

		// https://unicode.org/reports/tr29/#WB3b
		if is(_CRǀLFǀNewline, current) {
			break
		}

		// https://unicode.org/reports/tr29/#WB3c
		if is(emoji.Extended_Pictographic, current) && is(ZWJ, previous) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB3d
		if is(WSegSpace, current) && is(WSegSpace, previous) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB4
		if is(_ExtendǀFormatǀZWJ, current) {
			pos += w
			continue
		}

		// WB4 applies to subsequent rules; there is an implied "ignoring Extend & Format & ZWJ"
		// https://unicode.org/reports/tr29/#Grapheme_Cluster_and_Format_Rules
		// The Seek* methods are shorthand for "seek a category but skip over Extend|Format|ZWJ on the way"

		ignore := _ExtendǀFormatǀZWJ

		// https://unicode.org/reports/tr29/#WB5
		if is(AHLetter, current) && seek.Previous(data[:pos], ignore, AHLetter) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB6
		if is(_MidLetterǀMidNumLetQ, current) && seek.Forward(data[pos+w:], ignore, AHLetter) && seek.Previous(data[:pos], ignore, AHLetter) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB7
		if is(AHLetter, current) {
			previousIndex := seek.PreviousIndex(data[:pos], ignore, _MidLetterǀMidNumLetQ)
			if previousIndex >= 0 && seek.Previous(data[:previousIndex], ignore, AHLetter) {
				pos += w
				continue
			}
		}

		// https://unicode.org/reports/tr29/#WB7a
		if is(Single_Quote, current) && seek.Previous(data[:pos], ignore, Hebrew_Letter) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB7b
		if is(Double_Quote, current) && seek.Forward(data[pos+w:], ignore, Hebrew_Letter) && seek.Previous(data[:pos], ignore, Hebrew_Letter) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB7c
		if is(Hebrew_Letter, current) {
			previousIndex := seek.PreviousIndex(data[:pos], ignore, Double_Quote)
			if previousIndex >= 0 && seek.Previous(data[:previousIndex], ignore, Hebrew_Letter) {
				pos += w
				continue
			}
		}

		// https://unicode.org/reports/tr29/#WB8
		if is(Numeric, current) && seek.Previous(data[:pos], ignore, Numeric) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB9
		if is(Numeric, current) && seek.Previous(data[:pos], ignore, AHLetter) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB10
		if is(AHLetter, current) && seek.Previous(data[:pos], ignore, Numeric) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB11
		if is(Numeric, current) {
			previousIndex := seek.PreviousIndex(data[:pos], ignore, _MidNumǀMidNumLetQ)
			if previousIndex >= 0 && seek.Previous(data[:previousIndex], ignore, Numeric) {
				pos += w
				continue
			}
		}

		// https://unicode.org/reports/tr29/#WB12
		if is(_MidNumǀMidNumLetQ, current) && seek.Forward(data[pos+w:], ignore, Numeric) && seek.Previous(data[:pos], ignore, Numeric) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB13
		if is(Katakana, current) && seek.Previous(data[:pos], ignore, Katakana) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB13a
		if is(ExtendNumLet, current) && seek.Previous(data[:pos], ignore, _AHLetterǀNumericǀKatakanaǀExtendNumLet) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB13b
		if is(_AHLetterǀNumericǀKatakana, current) && seek.Previous(data[:pos], ignore, ExtendNumLet) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB15
		if is(Regional_Indicator, current) {
			allRI := true

			// Buffer comprised entirely of an odd number of RI, ignoring Extend|Format|ZWJ
			i := pos
			count := 0
			for i >= 0 {
				r, w := utf8.DecodeLastRune(data[:i])
				i -= w
				if is(ignore, r) {
					continue
				}
				if !is(Regional_Indicator, r) {
					allRI = false
					break
				}
				count++
			}

			if allRI {
				odd := count > 0 && count%2 == 1
				if odd {
					pos += w
					continue
				}
			}
		}

		// https://unicode.org/reports/tr29/#WB16
		if is(Regional_Indicator, current) {
			odd := false
			// Last n runes represent an odd number of RI, ignoring Extend|Format|ZWJ
			i := pos
			count := 0
			for i >= 0 {
				r, w := utf8.DecodeLastRune(data[:i])
				i -= w
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
				pos += w
				continue
			}
		}

		break
	}

	// https://unicode.org/reports/tr29/#WB999
	// If we fall through all the above rules, it's a word break
	return pos, data[:pos], nil
}
