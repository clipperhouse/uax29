// Package words provides a scanner for Unicode text segmentation word boundaries: https://unicode.org/reports/tr29/#Word_Boundaries
package words

import (
	"bufio"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/seeker"
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

var trie = newWordsTrie(0)
var seek = seeker.New(trie.lookup, _Ignore)

// is tests if the first rune of s is in categories
func is(categories uint32, s []byte) bool {
	lookup, _ := trie.lookup(s)
	return (lookup & categories) != 0
}

var _AHLetter = _ALetter | _Hebrew_Letter
var _MidNumLetQ = _MidNumLet | _Single_Quote
var _Ignore = _Extend | _Format | _ZWJ

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
		if is(_LF, data[pos:]) && is(_CR, previous) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB3a
		if is(_CR|_LF|_Newline, previous) {
			break
		}

		// https://unicode.org/reports/tr29/#WB3b
		if is(_CR|_LF|_Newline, data[pos:]) {
			break
		}

		// https://unicode.org/reports/tr29/#WB3c
		if is(_Extended_Pictographic, data[pos:]) && is(_ZWJ, previous) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB3d
		if is(_WSegSpace, data[pos:]) && is(_WSegSpace, previous) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB4
		if is(_Extend|_Format|_ZWJ, data[pos:]) {
			pos += w
			continue
		}

		// WB4 applies to subsequent rules; there is an implied "ignoring Extend & Format & ZWJ"
		// https://unicode.org/reports/tr29/#Grapheme_Cluster_and_Format_Rules
		// The Seek* methods are shorthand for "seek a category but skip over Extend|Format|ZWJ on the way"

		// https://unicode.org/reports/tr29/#WB5
		if is(_AHLetter, data[pos:]) &&
			seek.Previous(_AHLetter, data[:pos]) {
			pos += w

			// Optimization: there's a likelihood of a run of AHLetter
			for pos < len(data) {
				_, w := utf8.DecodeRune(data[pos:])
				if is(_AHLetter, data[pos:]) {
					pos += w
					continue
				}
				break
			}

			continue
		}

		// https://unicode.org/reports/tr29/#WB6
		if is(_MidLetter|_MidNumLetQ, data[pos:]) &&
			seek.Forward(_AHLetter, data[pos+w:]) &&
			seek.Previous(_AHLetter, data[:pos]) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB7
		if is(_AHLetter, data[pos:]) {
			previousIndex := seek.PreviousIndex(_MidLetter|_MidNumLetQ, data[:pos])
			if previousIndex >= 0 && seek.Previous(_AHLetter, data[:previousIndex]) {
				pos += w
				continue
			}
		}

		// https://unicode.org/reports/tr29/#WB7a
		if is(_Single_Quote, data[pos:]) &&
			seek.Previous(_Hebrew_Letter, data[:pos]) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB7b
		if is(_Double_Quote, data[pos:]) &&
			seek.Forward(_Hebrew_Letter, data[pos+w:]) &&
			seek.Previous(_Hebrew_Letter, data[:pos]) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB7c
		if is(_Hebrew_Letter, data[pos:]) {
			previousIndex := seek.PreviousIndex(_Double_Quote, data[:pos])
			if previousIndex >= 0 && seek.Previous(_Hebrew_Letter, data[:previousIndex]) {
				pos += w
				continue
			}
		}

		// https://unicode.org/reports/tr29/#WB8
		if is(_Numeric, data[pos:]) && seek.Previous(_Numeric, data[:pos]) {
			pos += w

			// Optimization: there's a likelihood of a run of Numeric
			for pos < len(data) {
				_, w := utf8.DecodeRune(data[pos:])
				if is(_Numeric, data[pos:]) {
					pos += w
					continue
				}
				break
			}

			continue
		}

		// https://unicode.org/reports/tr29/#WB9
		if is(_Numeric, data[pos:]) && seek.Previous(_AHLetter, data[:pos]) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB10
		if is(_AHLetter, data[pos:]) && seek.Previous(_Numeric, data[:pos]) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB11
		if is(_Numeric, data[pos:]) {
			previousIndex := seek.PreviousIndex(_MidNum|_MidNumLetQ, data[:pos])
			if previousIndex >= 0 && seek.Previous(_Numeric, data[:previousIndex]) {
				pos += w
				continue
			}
		}

		// https://unicode.org/reports/tr29/#WB12
		if is(_MidNum|_MidNumLet|_Single_Quote, data[pos:]) &&
			seek.Forward(_Numeric, data[pos+w:]) &&
			seek.Previous(_Numeric, data[:pos]) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB13
		if is(_Katakana, data[pos:]) &&
			seek.Previous(_Katakana, data[:pos]) {
			pos += w

			// Optimization: there's a likelihood of a run of Katakana
			for pos < len(data) {
				_, w := utf8.DecodeRune(data[pos:])
				if is(_Katakana, data[pos:]) {
					pos += w
					continue
				}
				break
			}

			continue
		}

		// https://unicode.org/reports/tr29/#WB13a
		if is(_ExtendNumLet, data[pos:]) &&
			seek.Previous(_ALetter|_Hebrew_Letter|_Numeric|_Katakana|_ExtendNumLet, data[:pos]) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB13b
		if is(_ALetter|_Hebrew_Letter|_Numeric|_Katakana, data[pos:]) &&
			seek.Previous(_ExtendNumLet, data[:pos]) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB15
		if is(_Regional_Indicator, data[pos:]) {
			allRI := true

			// Buffer comprised entirely of an odd number of RI, ignoring Extend|Format|ZWJ
			i := pos
			count := 0
			for i > 0 {
				_, w := utf8.DecodeLastRune(data[:i])
				i -= w
				if is(_Ignore, data[i:]) {
					continue
				}
				if !is(_Regional_Indicator, data[i:]) {
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
		if is(_Regional_Indicator, data[pos:]) {
			odd := false
			// Last n runes represent an odd number of RI, ignoring Extend|Format|ZWJ
			i := pos
			count := 0
			for i > 0 {
				_, w := utf8.DecodeLastRune(data[:i])
				i -= w
				if is(_Ignore, data[i:]) {
					continue
				}
				if !is(_Regional_Indicator, data[i:]) {
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
