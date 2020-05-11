// Package graphemes provides a scanner for Unicode text segmentation grapheme cluster boundaries: https://unicode.org/reports/tr29/#Grapheme_Cluster_Boundaries
package graphemes

import (
	"bufio"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/seeker"
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
var seek = seeker.New(trie.lookup, bExtend)

// Is tests if the first rune of s is in categories
func Is(categories uint32, s []byte) bool {
	lookup, _ := trie.lookup(s)
	return (lookup & categories) != 0
}

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
		if Is(bLF, data[pos:]) && Is(bCR, previous) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB4
		if Is(bControl|bCR|bLF, previous) {
			break
		}

		// https://unicode.org/reports/tr29/#GB5
		if Is(bControl|bCR|bLF, data[pos:]) {
			break
		}

		// https://unicode.org/reports/tr29/#GB6
		if Is(bL|bV|bLV|bLVT, data[pos:]) && Is(bL, previous) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB7
		if Is(bV|bT, data[pos:]) && Is(bLV|bV, previous) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB8
		if Is(bT, data[pos:]) && Is(bLVT|bT, previous) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB9
		if Is(bExtend|bZWJ, data[pos:]) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB9a
		if Is(bSpacingMark, data[pos:]) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB9b
		if Is(bPrepend, previous) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB11
		if Is(bExtended_Pictographic, data[pos:]) && Is(bZWJ, previous) &&
			seek.Previous(bExtended_Pictographic, data[:pos-pw]) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB12
		if Is(bRegional_Indicator, data[pos:]) {
			allRI := true

			// Buffer comprised entirely of an odd number of RI, ignoring Extend|Format|ZWJ
			i := pos
			count := 0
			for i > 0 {
				_, w := utf8.DecodeLastRune(data[:i])
				i -= w
				if !Is(bRegional_Indicator, data[i:]) {
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
		if Is(bRegional_Indicator, data[pos:]) {
			odd := false
			// Last n runes represent an odd number of RI, ignoring Extend|Format|ZWJ
			i := pos
			count := 0
			for i > 0 {
				_, w := utf8.DecodeLastRune(data[:i])
				i -= w
				if !Is(bRegional_Indicator, data[i:]) {
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
