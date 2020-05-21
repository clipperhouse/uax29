// Package words provides a scanner for Unicode text segmentation word boundaries: https://unicode.org/reports/tr29/#Word_Boundaries
package words

import (
	"bufio"
	"fmt"
	"io"
	"unicode/utf8"
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

// is tests if the first rune of s is in categories
func is(categories uint32, s []byte) bool {
	lookup, _ := trie.lookup(s)
	return (lookup & categories) != 0
}

var _AHLetter = _ALetter | _HebrewLetter
var _MidNumLetQ = _MidNumLet | _SingleQuote
var _Ignore = _Extend | _Format | _ZWJ

// SplitFunc is a bufio.SplitFunc implementation of word segmentation, for use with bufio.Scanner
func SplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	current := 0

	for {
		if current == len(data) && !atEOF {
			// Request more data
			return 0, nil, nil
		}

		sot := current == 0 // "start of text"
		eof := len(data) == current

		// https://unicode.org/reports/tr29/#WB1
		if sot && !eof {
			_, w := utf8.DecodeRune(data[current:])
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB2
		if eof {
			break
		}

		// Rules are usually of the form Cat1 × Cat2; "current" refers to the first category
		// to the right of the ×, from which we look back or forward

		// Decoding runes is a bit redundant, it happens in other places too
		// We do it here for clarity and to pick up errors early

		r, w := utf8.DecodeRune(data[current:])
		if r == utf8.RuneError {
			return 0, nil, fmt.Errorf("error decoding rune at byte 0x%x", data[current])
		}

		next := current + w

		_, pw := utf8.DecodeLastRune(data[:current])
		last := current - pw

		// https://unicode.org/reports/tr29/#WB3
		if is(_LF, data[current:]) && is(_CR, data[last:]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB3a
		if is(_CR|_LF|_Newline, data[last:]) {
			break
		}

		// https://unicode.org/reports/tr29/#WB3b
		if is(_CR|_LF|_Newline, data[current:]) {
			break
		}

		// https://unicode.org/reports/tr29/#WB3c
		if is(_ExtendedPictographic, data[current:]) && is(_ZWJ, data[last:]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB3d
		if is(_WSegSpace, data[current:]) && is(_WSegSpace, data[last:]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB4
		if is(_Extend|_Format|_ZWJ, data[current:]) {
			current += w
			continue
		}

		// WB4 applies to subsequent rules; there is an implied "ignoring Extend & Format & ZWJ"
		// https://unicode.org/reports/tr29/#Grapheme_Cluster_and_Format_Rules
		// The Seek* methods are shorthand for "seek a category but skip over Extend|Format|ZWJ on the way"

		// https://unicode.org/reports/tr29/#WB5
		if is(_AHLetter, data[current:]) && previous(_AHLetter, data[:current]) {
			current += w

			// Optimization: there's a likelihood of a run of AHLetter
			for current < len(data) {
				_, w := utf8.DecodeRune(data[current:])
				if is(_AHLetter, data[current:]) {
					current += w
					continue
				}
				break
			}

			continue
		}

		// https://unicode.org/reports/tr29/#WB6
		if is(_MidLetter|_MidNumLetQ, data[current:]) && forward(_AHLetter, data[next:]) && previous(_AHLetter, data[:current]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB7
		if is(_AHLetter, data[current:]) {
			pi := previousIndex(_MidLetter|_MidNumLetQ, data[:current])
			if pi >= 0 && previous(_AHLetter, data[:pi]) {
				current += w
				continue
			}
		}

		// https://unicode.org/reports/tr29/#WB7a
		if is(_SingleQuote, data[current:]) && previous(_HebrewLetter, data[:current]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB7b
		if is(_DoubleQuote, data[current:]) && forward(_HebrewLetter, data[next:]) && previous(_HebrewLetter, data[:current]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB7c
		if is(_HebrewLetter, data[current:]) {
			pi := previousIndex(_DoubleQuote, data[:current])
			if pi >= 0 && previous(_HebrewLetter, data[:pi]) {
				current += w
				continue
			}
		}

		// https://unicode.org/reports/tr29/#WB8
		if is(_Numeric, data[current:]) && previous(_Numeric, data[:current]) {
			current += w

			// Optimization: there's a likelihood of a run of Numeric
			for current < len(data) {
				_, w := utf8.DecodeRune(data[current:])
				if is(_Numeric, data[current:]) {
					current += w
					continue
				}
				break
			}

			continue
		}

		// https://unicode.org/reports/tr29/#WB9
		if is(_Numeric, data[current:]) && previous(_AHLetter, data[:current]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB10
		if is(_AHLetter, data[current:]) && previous(_Numeric, data[:current]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB11
		if is(_Numeric, data[current:]) {
			pi := previousIndex(_MidNum|_MidNumLetQ, data[:current])
			if pi >= 0 && previous(_Numeric, data[:pi]) {
				current += w
				continue
			}
		}

		// https://unicode.org/reports/tr29/#WB12
		if is(_MidNum|_MidNumLet|_SingleQuote, data[current:]) && forward(_Numeric, data[next:]) && previous(_Numeric, data[:current]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB13
		if is(_Katakana, data[current:]) && previous(_Katakana, data[:current]) {
			current += w

			// Optimization: there's a likelihood of a run of Katakana
			for current < len(data) {
				_, w := utf8.DecodeRune(data[current:])
				if is(_Katakana, data[current:]) {
					current += w
					continue
				}
				break
			}

			continue
		}

		// https://unicode.org/reports/tr29/#WB13a
		if is(_ExtendNumLet, data[current:]) && previous(_AHLetter|_Numeric|_Katakana|_ExtendNumLet, data[:current]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB13b
		if is(_AHLetter|_Numeric|_Katakana, data[current:]) && previous(_ExtendNumLet, data[:current]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB15
		if is(_RegionalIndicator, data[current:]) {
			allRI := true

			// Buffer comprised entirely of an odd number of RI, ignoring Extend|Format|ZWJ
			i := current
			count := 0
			for i > 0 {
				_, w := utf8.DecodeLastRune(data[:i])
				i -= w
				if is(_Ignore, data[i:]) {
					continue
				}
				if !is(_RegionalIndicator, data[i:]) {
					allRI = false
					break
				}
				count++
			}

			if allRI {
				odd := count > 0 && count%2 == 1
				if odd {
					current += w
					continue
				}
			}
		}

		// https://unicode.org/reports/tr29/#WB16
		if is(_RegionalIndicator, data[current:]) {
			odd := false
			// Last n runes represent an odd number of RI, ignoring Extend|Format|ZWJ
			i := current
			count := 0
			for i > 0 {
				_, w := utf8.DecodeLastRune(data[:i])
				i -= w
				if is(_Ignore, data[i:]) {
					continue
				}
				if !is(_RegionalIndicator, data[i:]) {
					odd = count > 0 && count%2 == 1
					break
				}
				count++
			}

			if odd {
				current += w
				continue
			}
		}

		// https://unicode.org/reports/tr29/#WB999
		// If we fall through all the above rules, it's a word break
		break
	}

	return current, data[:current], nil
}
