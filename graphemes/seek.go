package graphemes

import "unicode/utf8"

// seekPrevious works backward in the buffer until it hits a rune in the `seek` category,
// ignoring runes in the `ignore` category. It returns true if `seek` is found.
func seekPrevious(seek uint32, data []byte) bool {
	// Start at the end of the buffer and move backwards
	i := len(data)
	for i > 0 {
		r, w := utf8.DecodeLastRune(data[:i])
		i -= w

		_ = r

		if is(_Ignore, data[i:]) {
			continue
		}

		if is(seek, data[i:]) {
			return true
		}

		// If we get this far, it's not there
		break
	}

	return false
}
