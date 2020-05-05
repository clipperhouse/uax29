package seek

import (
	"unicode"
	"unicode/utf8"
)

var is = unicode.Is

// PreviousIndex works backward until it hits a rune in the `seek` category,
// ignoring runes in the `ignore` category, and returns the index of the rune in the buffer.
// It returns -1 if `seek` rune is not found.
func PreviousIndex(data []byte, ignore, seek *unicode.RangeTable) int {
	// Start at the end of the buffer and move backwards
	i := len(data)
	for i >= 0 {
		r, w := utf8.DecodeLastRune(data[:i])
		i -= w

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

// Previous works backward in the buffer until it hits a rune in the `seek` category,
// ignoring runes in the `ignore` category. It returns true if `seek` is found.
func Previous(data []byte, ignore, seek *unicode.RangeTable) bool {
	return PreviousIndex(data, ignore, seek) >= 0
}

// Forward looks ahead in the buffer until it hits a rune in the `seek` category,
// ignoring runes in the `ignore` category. It returns true if `seek` is found.
func Forward(data []byte, ignore, seek *unicode.RangeTable) bool {
	i := 0
	for i < len(data) {
		r, w := utf8.DecodeRune(data[i:])
		i += w

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
