// Package words provides a scanner for Unicode text segmentation word boundaries: https://unicode.org/reports/tr29/#Word_Boundaries
package words

import (
	"bufio"
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
	return is2(categories, lookup)
}

// is2 tests if the first rune of s is in categories
func is2(categories, lookup uint32) bool {
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

		lookup, w := trie.lookup(data[current:])

		next := current + w

		_, pw := utf8.DecodeLastRune(data[:current])
		last := current - pw

		// https://unicode.org/reports/tr29/#WB3
		if is2(_LF, lookup) && is(_CR, data[last:]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB3a
		if is(_CR|_LF|_Newline, data[last:]) {
			break
		}

		// https://unicode.org/reports/tr29/#WB3b
		if is2(_CR|_LF|_Newline, lookup) {
			break
		}

		// https://unicode.org/reports/tr29/#WB3c
		if is2(_ExtendedPictographic, lookup) && is(_ZWJ, data[last:]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB3d
		if is2(_WSegSpace, lookup) && is(_WSegSpace, data[last:]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB4
		if is2(_Extend|_Format|_ZWJ, lookup) {
			current += w
			continue
		}

		// WB4 applies to subsequent rules; there is an implied "ignoring Extend & Format & ZWJ"
		// https://unicode.org/reports/tr29/#Grapheme_Cluster_and_Format_Rules
		// The previous/subsequent methods are shorthand for "seek a category but skip over Extend|Format|ZWJ on the way"

		// https://unicode.org/reports/tr29/#WB5
		if is2(_AHLetter, lookup) && previous(_AHLetter, data[:current]) {
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
		if is2(_MidLetter|_MidNumLetQ, lookup) && subsequent(_AHLetter, data[next:]) && previous(_AHLetter, data[:current]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB7
		if is2(_AHLetter, lookup) {
			pi := previousIndex(_MidLetter|_MidNumLetQ, data[:current])
			if pi >= 0 && previous(_AHLetter, data[:pi]) {
				current += w
				continue
			}
		}

		// https://unicode.org/reports/tr29/#WB7a
		if is2(_SingleQuote, lookup) && previous(_HebrewLetter, data[:current]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB7b
		if is2(_DoubleQuote, lookup) && subsequent(_HebrewLetter, data[next:]) && previous(_HebrewLetter, data[:current]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB7c
		if is2(_HebrewLetter, lookup) {
			pi := previousIndex(_DoubleQuote, data[:current])
			if pi >= 0 && previous(_HebrewLetter, data[:pi]) {
				current += w
				continue
			}
		}

		// https://unicode.org/reports/tr29/#WB8
		if is2(_Numeric, lookup) && previous(_Numeric, data[:current]) {
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
		if is2(_Numeric, lookup) && previous(_AHLetter, data[:current]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB10
		if is2(_AHLetter, lookup) && previous(_Numeric, data[:current]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB11
		if is2(_Numeric, lookup) {
			pi := previousIndex(_MidNum|_MidNumLetQ, data[:current])
			if pi >= 0 && previous(_Numeric, data[:pi]) {
				current += w
				continue
			}
		}

		// https://unicode.org/reports/tr29/#WB12
		if is2(_MidNum|_MidNumLet|_SingleQuote, lookup) && subsequent(_Numeric, data[next:]) && previous(_Numeric, data[:current]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB13
		if is2(_Katakana, lookup) && previous(_Katakana, data[:current]) {
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
		if is2(_ExtendNumLet, lookup) && previous(_AHLetter|_Numeric|_Katakana|_ExtendNumLet, data[:current]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB13b
		if is2(_AHLetter|_Numeric|_Katakana, lookup) && previous(_ExtendNumLet, data[:current]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB15
		if is2(_RegionalIndicator, lookup) {
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
		if is2(_RegionalIndicator, lookup) {
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
