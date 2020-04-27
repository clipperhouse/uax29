// Package graphemes provides a scanner for Unicode text segmentation grapheme cluster boundaries: https://unicode.org/reports/tr29/#Grapheme_Cluster_Boundaries
package graphemes

import (
	"io"
	"unicode"

	"github.com/clipperhouse/uax29"
	"github.com/clipperhouse/uax29/emoji"
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
func NewScanner(r io.Reader) *uax29.Scanner {
	return uax29.NewScanner(r, BreakFunc)
}

var is = unicode.Is

// BreakFunc implements grapheme cluster boundaries according to https://unicode.org/reports/tr29/#Grapheme_Cluster_Boundaries.
// It is intended for use with uax29.Scanner.
var BreakFunc uax29.BreakFunc = func(buffer uax29.Runes, pos uax29.Pos) bool {
	// Rules: https://unicode.org/reports/tr29/#Grapheme_Cluster_Boundary_Rules

	sot := pos == 0 // "start of text"
	eof := len(buffer) == int(pos)

	// https://unicode.org/reports/tr29/#GB1
	if sot && !eof {
		return uax29.Accept
	}

	// https://unicode.org/reports/tr29/#GB2
	if eof {
		return uax29.Break
	}

	current := buffer[pos]
	previous := buffer[pos-1]

	// https://unicode.org/reports/tr29/#GB3
	if is(LF, current) && is(CR, previous) {
		return uax29.Accept
	}

	// https://unicode.org/reports/tr29/#GB4
	if is(_ControlǀCRǀLF, previous) {
		return uax29.Break
	}

	// https://unicode.org/reports/tr29/#GB5
	if is(_ControlǀCRǀLF, current) {
		return uax29.Break
	}

	// https://unicode.org/reports/tr29/#GB6
	if is(_LǀVǀLVǀLVT, current) && is(L, previous) {
		return uax29.Accept
	}

	// https://unicode.org/reports/tr29/#GB7
	if is(_VǀT, current) && is(_LVǀV, previous) {
		return uax29.Accept
	}

	// https://unicode.org/reports/tr29/#GB8
	if is(T, current) && is(_LVTǀT, previous) {
		return uax29.Accept
	}

	// https://unicode.org/reports/tr29/#GB9
	if is(_ExtendǀZWJ, current) {
		return uax29.Accept
	}

	// https://unicode.org/reports/tr29/#GB9a
	if is(SpacingMark, current) {
		return uax29.Accept
	}

	// https://unicode.org/reports/tr29/#GB9b
	if is(Prepend, previous) {
		return uax29.Accept
	}

	// https://unicode.org/reports/tr29/#GB11
	if is(emoji.Extended_Pictographic, current) && is(ZWJ, previous) &&
		buffer.SeekPrevious(pos-1, Extend, emoji.Extended_Pictographic) {
		return uax29.Accept
	}

	// https://unicode.org/reports/tr29/#GB12
	if is(Regional_Indicator, current) {
		// Buffer comprised entirely of an odd number of RI
		allRI := true
		count := 0
		for i := pos - 1; i >= 0; i-- {
			r := buffer[i]
			if !is(Regional_Indicator, r) {
				allRI = false
				break
			}
			count++
		}

		odd := count > 0 && count%2 == 1

		if allRI && odd {
			return uax29.Accept
		}
	}

	// https://unicode.org/reports/tr29/#GB13
	if is(Regional_Indicator, current) {
		// Last n runes represent an odd number of RI
		odd := false
		count := 0
		for i := pos - 1; i >= 0; i-- {
			r := buffer[i]
			if !is(Regional_Indicator, r) {
				odd = count > 0 && count%2 == 1
				break
			}
			count++
		}

		if odd {
			return uax29.Accept
		}
	}

	// https://unicode.org/reports/tr29/#WB999
	// If we fall through all the above rules, it's a break
	return uax29.Break
}
