// Package graphemes provides a scanner for Unicode text segmentation grapheme cluster boundaries: https://unicode.org/reports/tr29/#Grapheme_Cluster_Boundaries
package graphemes

import (
	"bufio"
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/emoji"
	"github.com/clipperhouse/uax29/seek"
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

var is = unicode.Is

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

		previous, wp := utf8.DecodeLastRune(data[:pos])

		// https://unicode.org/reports/tr29/#GB3
		if is(LF, current) && is(CR, previous) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB4
		if is(_ControlǀCRǀLF, previous) {
			break
		}

		// https://unicode.org/reports/tr29/#GB5
		if is(_ControlǀCRǀLF, current) {
			break
		}

		// https://unicode.org/reports/tr29/#GB6
		if is(_LǀVǀLVǀLVT, current) && is(L, previous) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB7
		if is(_VǀT, current) && is(_LVǀV, previous) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB8
		if is(T, current) && is(_LVTǀT, previous) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB9
		if is(_ExtendǀZWJ, current) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB9a
		if is(SpacingMark, current) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB9b
		if is(Prepend, previous) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB11
		if is(emoji.Extended_Pictographic, current) && is(ZWJ, previous) && seek.Previous(data[:pos-wp], Extend, emoji.Extended_Pictographic) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB12
		if is(Regional_Indicator, current) {
			allRI := true

			// Buffer comprised entirely of an odd number of RI, ignoring Extend|Format|ZWJ
			i := pos
			count := 0
			for i >= 0 {
				r, w := utf8.DecodeLastRune(data[:i])
				i -= w
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

		// https://unicode.org/reports/tr29/#GB13
		if is(Regional_Indicator, current) {
			odd := false
			// Last n runes represent an odd number of RI, ignoring Extend|Format|ZWJ
			i := pos
			count := 0
			for i >= 0 {
				r, w := utf8.DecodeLastRune(data[:i])
				i -= w
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

		// If we fall through all the above rules, it's a grapheme cluster break
		break
	}

	// Return token
	return pos, data[:pos], nil
}
