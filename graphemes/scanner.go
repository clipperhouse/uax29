// Package graphemes provides a scanner for Unicode text segmentation grapheme cluster boundaries: https://unicode.org/reports/tr29/#Grapheme_Cluster_Boundaries
package graphemes

import (
	"bufio"
	"fmt"
	"io"
	"unicode/utf8"
)

// NewScanner tokenizes a reader into a stream of grapheme clusters according to Unicode Text Segmentation boundaries https://unicode.org/reports/tr29/#Grapheme_Cluster_Boundaries
// Iterate through the stream by calling Scan() until false.
//	text := "Good dog! 👍🏼🐶"
//	reader := strings.NewReader(text)
//
//	scanner := graphemes.NewScanner(reader)
//	for scanner.Scan() {
//		fmt.Printf("%s\n", scanner.Text())
//	}
//	if err := scanner.Err(); err != nil {
//		log.Fatal(err)
//	}
func NewScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	scanner.Split(SplitFunc)
	return scanner
}

var trie = newGraphemesTrie(0)

// is tests if the first rune of s is in categories
func is(categories uint32, s []byte) bool {
	lookup, _ := trie.lookup(s)
	return (lookup & categories) != 0
}

var _Ignore = _Extend

// SplitFunc is a bufio.SplitFunc implementation of grapheme cluster segmentation, for use with bufio.Scanner
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

		// https://unicode.org/reports/tr29/#SB1
		if sot && !eof {
			_, w := utf8.DecodeRune(data[pos:])
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#SB2
		if eof {
			break
		}

		current, w := utf8.DecodeRune(data[pos:])
		if current == utf8.RuneError {
			return 0, nil, fmt.Errorf("error decoding rune")
		}

		_, pw := utf8.DecodeLastRune(data[:pos])
		previous := data[pos-pw:]

		// https://unicode.org/reports/tr29/#GB3
		if is(_LF, data[pos:]) && is(_CR, previous) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB4
		if is(_Control|_CR|_LF, previous) {
			break
		}

		// https://unicode.org/reports/tr29/#GB5
		if is(_Control|_CR|_LF, data[pos:]) {
			break
		}

		// https://unicode.org/reports/tr29/#GB6
		if is(_L|_V|_LV|_LVT, data[pos:]) && is(_L, previous) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB7
		if is(_V|_T, data[pos:]) && is(_LV|_V, previous) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB8
		if is(_T, data[pos:]) && is(_LVT|_T, previous) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB9
		if is(_Extend|_ZWJ, data[pos:]) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB9a
		if is(_SpacingMark, data[pos:]) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB9b
		if is(_Prepend, previous) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB11
		if is(_Extended_Pictographic, data[pos:]) && is(_ZWJ, previous) &&
			seekPrevious(_Extended_Pictographic, data[:pos-pw]) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB12
		if is(_Regional_Indicator, data[pos:]) {
			allRI := true

			// Buffer comprised entirely of an odd number of RI, ignoring Extend|Format|ZWJ
			i := pos
			count := 0
			for i > 0 {
				_, w := utf8.DecodeLastRune(data[:i])
				i -= w
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

		// https://unicode.org/reports/tr29/#GB13
		if is(_Regional_Indicator, data[pos:]) {
			odd := false
			// Last n runes represent an odd number of RI, ignoring Extend|Format|ZWJ
			i := pos
			count := 0
			for i > 0 {
				_, w := utf8.DecodeLastRune(data[:i])
				i -= w
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

		// If we fall through all the above rules, it's a grapheme cluster break
		break
	}

	// Return token
	return pos, data[:pos], nil
}
