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

// is determines if lookup intersects propert(ies)
func (lookup property) is(properties property) bool {
	return (lookup & properties) != 0
}

var _AHLetter = _ALetter | _HebrewLetter
var _MidNumLetQ = _MidNumLet | _SingleQuote
var _Ignore = _Extend | _Format | _ZWJ

// SplitFunc is a bufio.SplitFunc implementation of word segmentation, for use with bufio.Scanner
func SplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	var pos, w int
	var current, last property

	for {
		if pos == len(data) && !atEOF {
			// Request more data
			return 0, nil, nil
		}

		sot := pos == 0 // "start of text"
		eof := len(data) == pos

		// https://unicode.org/reports/tr29/#WB1
		if sot && !eof {
			current, w = trie.lookup(data[pos:])
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB2
		if eof {
			break
		}

		// Rules are usually of the form Cat1 × Cat2; "current" refers to the first property
		// to the right of the ×, from which we look back or forward

		last = current

		current, w = trie.lookup(data[pos:])
		if w == 0 {
			return 0, nil, fmt.Errorf("error decoding rune at byte 0x%x", data[pos])
		}

		next := pos + w

		// https://unicode.org/reports/tr29/#WB3
		if current.is(_LF) && last.is(_CR) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB3a
		if last.is(_CR | _LF | _Newline) {
			break
		}

		// https://unicode.org/reports/tr29/#WB3b
		if current.is(_CR | _LF | _Newline) {
			break
		}

		// https://unicode.org/reports/tr29/#WB3c
		if current.is(_ExtendedPictographic) && last.is(_ZWJ) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB3d
		if current.is(_WSegSpace) && last.is(_WSegSpace) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB4
		if current.is(_Extend | _Format | _ZWJ) {
			pos += w
			continue
		}

		// WB4 applies to subsequent rules; there is an implied "ignoring Extend & Format & ZWJ"
		// https://unicode.org/reports/tr29/#Grapheme_Cluster_and_Format_Rules
		// The previous/subsequent methods are shorthand for "seek a property but skip over Extend|Format|ZWJ on the way"

		// https://unicode.org/reports/tr29/#WB5
		if current.is(_AHLetter) && previous(_AHLetter, data[:pos]) {
			pos += w

			// Optimization: there's a likelihood of a run of AHLetter
			for pos < len(data) {
				lookup, w := trie.lookup(data[pos:])
				if lookup.is(_AHLetter) {
					current = lookup
					pos += w
					continue
				}
				break
			}

			continue
		}

		// https://unicode.org/reports/tr29/#WB6
		if current.is(_MidLetter|_MidNumLetQ) && subsequent(_AHLetter, data[next:]) && previous(_AHLetter, data[:pos]) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB7
		if current.is(_AHLetter) {
			pi := previousIndex(_MidLetter|_MidNumLetQ, data[:pos])
			if pi >= 0 && previous(_AHLetter, data[:pi]) {
				pos += w
				continue
			}
		}

		// https://unicode.org/reports/tr29/#WB7a
		if current.is(_SingleQuote) && previous(_HebrewLetter, data[:pos]) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB7b
		if current.is(_DoubleQuote) && subsequent(_HebrewLetter, data[next:]) && previous(_HebrewLetter, data[:pos]) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB7c
		if current.is(_HebrewLetter) {
			pi := previousIndex(_DoubleQuote, data[:pos])
			if pi >= 0 && previous(_HebrewLetter, data[:pi]) {
				pos += w
				continue
			}
		}

		// https://unicode.org/reports/tr29/#WB8
		if current.is(_Numeric) && previous(_Numeric, data[:pos]) {
			pos += w

			// Optimization: there's a likelihood of a run of Numeric
			for pos < len(data) {
				lookup, w := trie.lookup(data[pos:])
				if lookup.is(_Numeric) {
					current = lookup
					pos += w
					continue
				}
				break
			}

			continue
		}

		// https://unicode.org/reports/tr29/#WB9
		if current.is(_Numeric) && previous(_AHLetter, data[:pos]) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB10
		if current.is(_AHLetter) && previous(_Numeric, data[:pos]) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB11
		if current.is(_Numeric) {
			pi := previousIndex(_MidNum|_MidNumLetQ, data[:pos])
			if pi >= 0 && previous(_Numeric, data[:pi]) {
				pos += w
				continue
			}
		}

		// https://unicode.org/reports/tr29/#WB12
		if current.is(_MidNum|_MidNumLet|_SingleQuote) && subsequent(_Numeric, data[next:]) && previous(_Numeric, data[:pos]) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB13
		if current.is(_Katakana) && previous(_Katakana, data[:pos]) {
			pos += w

			// Optimization: there's a likelihood of a run of Katakana
			for pos < len(data) {
				lookup, w := trie.lookup(data[pos:])
				if lookup.is(_Katakana) {
					current = lookup
					pos += w
					continue
				}
				break
			}

			continue
		}

		// https://unicode.org/reports/tr29/#WB13a
		if current.is(_ExtendNumLet) && previous(_AHLetter|_Numeric|_Katakana|_ExtendNumLet, data[:pos]) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB13b
		if current.is(_AHLetter|_Numeric|_Katakana) && previous(_ExtendNumLet, data[:pos]) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB15
		if current.is(_RegionalIndicator) {
			allRI := true

			// Buffer comprised entirely of an odd number of RI, ignoring Extend|Format|ZWJ
			i := pos
			count := 0
			for i > 0 {
				_, w := utf8.DecodeLastRune(data[:i])
				i -= w

				lookup, _ := trie.lookup(data[i:])

				if lookup.is(_Ignore) {
					continue
				}

				if !lookup.is(_RegionalIndicator) {
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
		if current.is(_RegionalIndicator) {
			odd := false
			// Last n runes represent an odd number of RI, ignoring Extend|Format|ZWJ
			i := pos
			count := 0
			for i > 0 {
				_, w := utf8.DecodeLastRune(data[:i])
				i -= w

				lookup, _ := trie.lookup(data[i:])

				if lookup.is(_Ignore) {
					continue
				}

				if !lookup.is(_RegionalIndicator) {
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
