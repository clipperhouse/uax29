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
func is(categories uint16, s []byte) bool {
	lookup, _ := trie.lookup(s)
	return (lookup & categories) != 0
}

var _Ignore = _Extend

// SplitFunc is a bufio.SplitFunc implementation of grapheme cluster segmentation, for use with bufio.Scanner
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

		// https://unicode.org/reports/tr29/#SB1
		if sot && !eof {
			_, w := utf8.DecodeRune(data[current:])
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#SB2
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

		_, pw := utf8.DecodeLastRune(data[:current])
		last := current - pw

		// https://unicode.org/reports/tr29/#GB3
		if is(_LF, data[current:]) && is(_CR, data[last:]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB4
		if is(_Control|_CR|_LF, data[last:]) {
			break
		}

		// https://unicode.org/reports/tr29/#GB5
		if is(_Control|_CR|_LF, data[current:]) {
			break
		}

		// https://unicode.org/reports/tr29/#GB6
		if is(_L|_V|_LV|_LVT, data[current:]) && is(_L, data[last:]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB7
		if is(_V|_T, data[current:]) && is(_LV|_V, data[last:]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB8
		if is(_T, data[current:]) && is(_LVT|_T, data[last:]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB9
		if is(_Extend|_ZWJ, data[current:]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB9a
		if is(_SpacingMark, data[current:]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB9b
		if is(_Prepend, data[last:]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB11
		if is(_Extended_Pictographic, data[current:]) && is(_ZWJ, data[last:]) && previous(_Extended_Pictographic, data[:last]) {
			current += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB12
		if is(_Regional_Indicator, data[current:]) {
			allRI := true

			// Buffer comprised entirely of an odd number of RI, ignoring Extend|Format|ZWJ
			i := current
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
					current += w
					continue
				}
			}
		}

		// https://unicode.org/reports/tr29/#GB13
		if is(_Regional_Indicator, data[current:]) {
			odd := false
			// Last n runes represent an odd number of RI, ignoring Extend|Format|ZWJ
			i := current
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
				current += w
				continue
			}
		}

		// If we fall through all the above rules, it's a grapheme cluster break
		break
	}

	// Return token
	return current, data[:current], nil
}
