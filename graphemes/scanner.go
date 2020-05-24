// Package graphemes provides a scanner for Unicode text segmentation grapheme cluster boundaries: https://unicode.org/reports/tr29/#Grapheme_Cluster_Boundaries
package graphemes

import (
	"bufio"
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

// is tests if lookup intersects categories
func is(categories, lookup uint16) bool {
	return (categories & lookup) != 0
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

		// Rules are usually of the form Cat1 × Cat2; "current" refers to the first category
		// to the right of the ×, from which we look back or forward

		// Decoding runes is a bit redundant, it happens in other places too
		// We do it here for clarity and to pick up errors early
		current, w := trie.lookup(data[pos:])

		_, pw := utf8.DecodeLastRune(data[:pos])
		last, _ := trie.lookup(data[pos-pw:])

		// https://unicode.org/reports/tr29/#GB3
		if is(_LF, current) && is(_CR, last) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB4
		if is(_Control|_CR|_LF, last) {
			break
		}

		// https://unicode.org/reports/tr29/#GB5
		if is(_Control|_CR|_LF, current) {
			break
		}

		// https://unicode.org/reports/tr29/#GB6
		if is(_L|_V|_LV|_LVT, current) && is(_L, last) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB7
		if is(_V|_T, current) && is(_LV|_V, last) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB8
		if is(_T, current) && is(_LVT|_T, last) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB9
		if is(_Extend|_ZWJ, current) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB9a
		if is(_SpacingMark, current) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB9b
		if is(_Prepend, last) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB11
		if is(_ExtendedPictographic, current) && is(_ZWJ, last) && previous(_ExtendedPictographic, data[:pos-pw]) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB12
		if is(_RegionalIndicator, current) {
			allRI := true

			// Buffer comprised entirely of an odd number of RI
			i := pos
			count := 0
			for i > 0 {
				_, w := utf8.DecodeLastRune(data[:i])
				i -= w

				lookup, _ := trie.lookup(data[i:])

				if !is(_RegionalIndicator, lookup) {
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
		if is(_RegionalIndicator, current) {
			odd := false
			// Last n runes represent an odd number of RI
			i := pos
			count := 0
			for i > 0 {
				_, w := utf8.DecodeLastRune(data[:i])
				i -= w

				lookup, _ := trie.lookup(data[i:])

				if !is(_RegionalIndicator, lookup) {
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
