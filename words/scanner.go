// Package words provides a scanner for Unicode text segmentation word boundaries: https://unicode.org/reports/tr29/#Word_Boundaries
package words

import (
	"bufio"
	"fmt"
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

var trie = newWordsTrie(0)

// Is tests if the first rune of s is in categories
func Is(categories uint32, s []byte) bool {
	iotas, _ := trie.lookup(s)
	return (iotas & categories) != 0
}

// SplitFunc is a bufio.SplitFunc implementation of word segmentation, for use with bufio.Scanner
func SplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	pos := 0

	for {
		if pos == len(data) && !atEOF {
			// Request more data
			return 0, nil, nil
		}

		sot := pos == 0 // "start of text"
		eof := len(data) == pos

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
		if current == utf8.RuneError {
			return 0, nil, fmt.Errorf("error decoding rune")
		}

		_, pw := utf8.DecodeLastRune(data[:pos])
		previous := data[pos-pw:]

		// https://unicode.org/reports/tr29/#WB3
		if Is(bLF, data[pos:]) && Is(bCR, previous) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB3a
		if Is(bCR|bLF|bNewline, previous) {
			break
		}

		// https://unicode.org/reports/tr29/#WB3b
		if Is(bCR|bLF|bNewline, data[pos:]) {
			break
		}

		// https://unicode.org/reports/tr29/#WB3c
		if is(emoji.Extended_Pictographic, current) && Is(bZWJ, previous) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB3d
		if Is(bWSegSpace, data[pos:]) && Is(bWSegSpace, previous) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB4
		if Is(bExtend|bFormat|bZWJ, data[pos:]) {
			pos += w
			continue
		}

		// WB4 applies to subsequent rules; there is an implied "ignoring Extend & Format & ZWJ"
		// https://unicode.org/reports/tr29/#Grapheme_Cluster_and_Format_Rules
		// The Seek* methods are shorthand for "seek a category but skip over Extend|Format|ZWJ on the way"

		ignore := _ExtendǀFormatǀZWJ

		// https://unicode.org/reports/tr29/#WB5
		if Is(bALetter|bHebrew_Letter, data[pos:]) && seek.Previous(data[:pos], ignore, AHLetter) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB6
		if Is(bMidLetter|bMidNumLet|bSingle_Quote, data[pos:]) && seek.Forward(data[pos+w:], ignore, AHLetter) && seek.Previous(data[:pos], ignore, AHLetter) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB7
		if Is(bALetter|bHebrew_Letter, data[pos:]) {
			previousIndex := seek.PreviousIndex(data[:pos], ignore, _MidLetterǀMidNumLetQ)
			if previousIndex >= 0 && seek.Previous(data[:previousIndex], ignore, AHLetter) {
				pos += w
				continue
			}
		}

		// https://unicode.org/reports/tr29/#WB7a
		if Is(bSingle_Quote, data[pos:]) && seek.Previous(data[:pos], ignore, Hebrew_Letter) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB7b
		if Is(bDouble_Quote, data[pos:]) && seek.Forward(data[pos+w:], ignore, Hebrew_Letter) && seek.Previous(data[:pos], ignore, Hebrew_Letter) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB7c
		if Is(bHebrew_Letter, data[pos:]) {
			previousIndex := seek.PreviousIndex(data[:pos], ignore, Double_Quote)
			if previousIndex >= 0 && seek.Previous(data[:previousIndex], ignore, Hebrew_Letter) {
				pos += w
				continue
			}
		}

		// https://unicode.org/reports/tr29/#WB8
		if Is(bNumeric, data[pos:]) && seek.Previous(data[:pos], ignore, Numeric) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB9
		if Is(bNumeric, data[pos:]) && seek.Previous(data[:pos], ignore, AHLetter) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB10
		if Is(bALetter|bHebrew_Letter, data[pos:]) && seek.Previous(data[:pos], ignore, Numeric) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB11
		if Is(bNumeric, data[pos:]) {
			previousIndex := seek.PreviousIndex(data[:pos], ignore, _MidNumǀMidNumLetQ)
			if previousIndex >= 0 && seek.Previous(data[:previousIndex], ignore, Numeric) {
				pos += w
				continue
			}
		}

		// https://unicode.org/reports/tr29/#WB12
		if Is(bMidNum|bMidNumLet|bSingle_Quote, data[pos:]) && seek.Forward(data[pos+w:], ignore, Numeric) && seek.Previous(data[:pos], ignore, Numeric) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB13
		if Is(bKatakana, data[pos:]) && seek.Previous(data[:pos], ignore, Katakana) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB13a
		if Is(bExtendNumLet, data[pos:]) && seek.Previous(data[:pos], ignore, _AHLetterǀNumericǀKatakanaǀExtendNumLet) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB13b
		if Is(bALetter|bHebrew_Letter|bNumeric|bKatakana, data[pos:]) && seek.Previous(data[:pos], ignore, ExtendNumLet) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB15
		if Is(bRegional_Indicator, data[pos:]) {
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
		if Is(bRegional_Indicator, data[pos:]) {
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

		// https://unicode.org/reports/tr29/#WB999
		// If we fall through all the above rules, it's a word break
		break
	}

	return pos, data[:pos], nil
}
