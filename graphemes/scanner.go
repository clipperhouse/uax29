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

// is determines if lookup intersects propert(ies)
func (lookup property) is(properties property) bool {
	return (lookup & properties) != 0
}

var _Ignore = _Extend

// SplitFunc is a bufio.SplitFunc implementation of grapheme cluster segmentation, for use with bufio.Scanner
func SplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if len(data) == 0 {
		return 0, nil, nil
	}

	// These vars are stateful across loop iterations
	var pos, w int
	var current property

	for {
		sot := pos == 0         // "start of text"
		eot := pos == len(data) // "end of text"

		if eot && !atEOF {
			// Token extends past current data, request more
			return 0, nil, nil
		}

		// https://unicode.org/reports/tr29/#SB1
		if sot {
			current, w = trie.lookup(data[pos:])
			if w == 0 {
				// Rune extends past current data, request more
				return 0, nil, nil
			}
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#SB2
		if eot {
			break
		}

		// Rules are usually of the form Cat1 × Cat2; "current" refers to the first property
		// to the right of the ×, from which we look back or forward

		last := current
		lastw := w

		current, w = trie.lookup(data[pos:])
		if w == 0 {
			// Rune extends past current data, request more
			return 0, nil, nil
		}

		// https://unicode.org/reports/tr29/#GB3
		if current.is(_LF) && last.is(_CR) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB4
		if last.is(_Control | _CR | _LF) {
			break
		}

		// https://unicode.org/reports/tr29/#GB5
		if current.is(_Control | _CR | _LF) {
			break
		}

		// https://unicode.org/reports/tr29/#GB6
		if current.is(_L|_V|_LV|_LVT) && last.is(_L) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB7
		if current.is(_V|_T) && last.is(_LV|_V) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB8
		if current.is(_T) && last.is(_LVT|_T) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB9
		if current.is(_Extend | _ZWJ) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB9a
		if current.is(_SpacingMark) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB9b
		if last.is(_Prepend) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB11
		if current.is(_ExtendedPictographic) && last.is(_ZWJ) && previous(_ExtendedPictographic, data[:pos-lastw]) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB12 and
		// https://unicode.org/reports/tr29/#GB13
		if current.is(_RegionalIndicator) && last.is(_RegionalIndicator) {
			i := pos
			count := 0

			for i > 0 {
				_, w := utf8.DecodeLastRune(data[:i])
				i -= w

				lookup, _ := trie.lookup(data[i:])

				if !lookup.is(_RegionalIndicator) {
					// It's GB13
					break
				}

				count++
			}

			// If i == 0, we fell through and hit sot (start of text), so GB12 applies
			// If i > 0, we hit a non-RI, so GB13 applies

			oddRI := count%2 == 1
			if oddRI {
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
