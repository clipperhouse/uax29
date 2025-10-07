package graphemes

import (
	"bufio"

	"github.com/clipperhouse/uax29/v2/internal/stringish"
)

// is determines if lookup intersects propert(ies)
func (lookup property) is(properties property) bool {
	return (lookup & properties) != 0
}

// mask returns all-ones (0xFFFF) if p is non-zero, all-zeros otherwise, without branching
func mask(p property) property {
	// Branchless: (p | -p) has bit 15 set iff p != 0
	// Right shift by 15 gives 1 if non-zero, 0 if zero
	// Negate to get 0xFFFF or 0x0000
	return -property((p | -p) >> 15)
}

const _Ignore = _Extend

// SplitFunc is a bufio.SplitFunc implementation of Unicode grapheme cluster segmentation, for use with bufio.Scanner.
//
// See https://unicode.org/reports/tr29/#Grapheme_Cluster_Boundaries.
var SplitFunc bufio.SplitFunc = splitFunc[[]byte]

func splitFunc[T stringish.Interface](data T, atEOF bool) (advance int, token T, err error) {
	var empty T
	if len(data) == 0 {
		return 0, empty, nil
	}

	// These vars are stateful across loop iterations
	var pos int
	var lastExIgnore property = 0     // "last excluding ignored categories"
	var lastLastExIgnore property = 0 // "last one before that"
	var regionalIndicatorCount int

	// Rules are usually of the form Cat1 × Cat2; "current" refers to the first property
	// to the right of the ×, from which we look back or forward

	current, w := lookup(data[pos:])
	if w == 0 {
		if !atEOF {
			// Rune extends past current data, request more
			return 0, empty, nil
		}
		pos = len(data)
		return pos, data[:pos], nil
	}

	// https://unicode.org/reports/tr29/#GB1
	// Start of text always advances
	pos += w

	for {
		eot := pos == len(data) // "end of text"

		if eot {
			if !atEOF {
				// Token extends past current data, request more
				return 0, empty, nil
			}

			// https://unicode.org/reports/tr29/#GB2
			break
		}

		/*
			We've switched the evaluation order of GB1↓ and GB2↑. It's ok:
			because we've checked for len(data) at the top of this function,
			sot and eot are mutually exclusive, order doesn't matter.
		*/

		// Rules are usually of the form Cat1 × Cat2; "current" refers to the first property
		// to the right of the ×, from which we look back or forward

		// Remember previous properties to avoid lookups/lookbacks
		last := current
		if !last.is(_Ignore) {
			lastLastExIgnore = lastExIgnore
			lastExIgnore = last
		}

		current, w = lookup(data[pos:])
		if w == 0 {
			if atEOF {
				// Just return the bytes, we can't do anything with them
				pos = len(data)
				break
			}
			// Rune extends past current data, request more
			return 0, empty, nil
		}

		// Optimization: no rule can possibly apply
		if current|last == 0 { // i.e. both are zero
			break
		}

		// https://unicode.org/reports/tr29/#GB3
		// Branchless: only check _LF if last has _CR
		if current.is(_LF & mask(last&_CR)) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB4
		// https://unicode.org/reports/tr29/#GB5
		if (current | last).is(_Control | _CR | _LF) {
			break
		}

		// https://unicode.org/reports/tr29/#GB6
		// https://unicode.org/reports/tr29/#GB7
		// https://unicode.org/reports/tr29/#GB8
		// Combined Hangul syllable rules using bitwise ops to avoid branches
		// Compute which properties current is allowed to have, based on last
		allowed := ((_L | _V | _LV | _LVT) & mask(last&_L)) |
			((_V | _T) & mask(last&(_LV|_V))) |
			(_T & mask(last&(_LVT|_T)))
		if current.is(allowed) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB9
		// https://unicode.org/reports/tr29/#GB9a
		// Combined: don't break after Extend, ZWJ, or SpacingMark
		if current.is(_Extend | _ZWJ | _SpacingMark) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB9b
		if last.is(_Prepend) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB9c
		// TODO(clipperhouse):
		// It appears to be added in Unicode 15.1.0:
		// https://unicode.org/versions/Unicode15.1.0/#Migration
		// This package currently supports Unicode 15.0.0, so
		// out of scope for now

		// https://unicode.org/reports/tr29/#GB11
		// Branchless: check _ExtendedPictographic only if last has _ZWJ and lastLastExIgnore has _ExtendedPictographic
		if current.is(_ExtendedPictographic & mask(last&_ZWJ) & mask(lastLastExIgnore&_ExtendedPictographic)) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#GB12
		// https://unicode.org/reports/tr29/#GB13
		if (current & last).is(_RegionalIndicator) {
			regionalIndicatorCount++

			odd := regionalIndicatorCount%2 == 1
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
