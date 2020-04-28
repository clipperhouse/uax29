package uax29

import "unicode"

var is = unicode.Is

// Runes is a slice of runes, for use as a buffer
type Runes []rune

// Pos is a cursor for a Runes buffer
type Pos int

// SeekPreviousIndex works backward until it hits a rune in the `seek` category,
// ignoring runes in the `ignore` category, and returns the index of the rune in the buffer.
// It returns -1 if `seek` rune is not found.
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

// SeekPrevious works backward in the buffer until it hits a rune in the `seek` category,
// ignoring runes in the `ignore` category. It returns true if `seek` is found.
func (buffer Runes) SeekPrevious(pos Pos, ignore, seek *unicode.RangeTable) bool {
	return buffer.SeekPreviousIndex(pos, ignore, seek) >= 0
}

// SeekForward looks ahead in the buffer until it hits a rune in the `seek` category,
// ignoring runes in the `ignore` category. It returns true if `seek` is found.
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
