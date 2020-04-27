package uax29

import "unicode"

var is = unicode.Is

// Runes is a slice of runes, handy to use as a buffer
type Runes []rune

// Pos is a cursor for Runes
type Pos int

// SeekPreviousIndex works backward until it hits a rune satisfying one of the range tables,
// ignoring Extend, and returns the index of the rune in the buffer
// See: https://unicode.org/reports/tr29/#Grapheme_Cluster_and_Format_Rules (driven by SB5)
func (buffer Runes) SeekPreviousIndex(pos Pos, ignore, seek *unicode.RangeTable) Pos {
	// Start at the end of the buffer and move backwards
	for i := pos - 1; i >= 0; i-- {
		r := buffer[i]

		if is(ignore, r) {
			continue
		}

		if is(seek, r) {
			return i
		}

		// If we get this far, it's not there
		break
	}

	return -1
}

// SeekPrevious works backward ahead until it hits a rune satisfying one of the range tables,
// ignoring the ignore, reporting success
// Logic is here: https://unicode.org/reports/tr29/#Grapheme_Cluster_and_Format_Rules (driven by SB5)
func (buffer Runes) SeekPrevious(pos Pos, ignore, seek *unicode.RangeTable) bool {
	return buffer.SeekPreviousIndex(pos, ignore, seek) >= 0
}

// SeekForward looks ahead until it hits a rune satisfying one of the range tables,
// ignoring ignore
// See: https://unicode.org/reports/tr29/#Grapheme_Cluster_and_Format_Rules (driven by SB5)
func (buffer Runes) SeekForward(pos Pos, ignore, seek *unicode.RangeTable) bool {
	for i := int(pos) + 1; i < len(buffer); i++ {
		r := buffer[i]

		if is(ignore, r) {
			continue
		}

		if is(seek, r) {
			return true
		}

		// If we get this far, it's not there
		break
	}

	return false
}
