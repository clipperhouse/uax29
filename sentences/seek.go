package sentences

import "unicode/utf8"

// seekPreviousIndex works backward until it hits a rune in the `seek` category,
// ignoring runes in the `ignore` category, and returns the index of the rune in the buffer.
// It returns -1 if `seek` rune is not found.
func previousIndex(seek uint16, data []byte) int {
	// Start at the end of the buffer and move backwards
	i := len(data)
	for i > 0 {
		_, w := utf8.DecodeLastRune(data[:i])
		i -= w

		if is(_Ignore, data[i:]) {
			continue
		}

		if is(seek, data[i:]) {
			return i
		}

		// If we get this far, it's not there
		break
	}

	return -1
}

// seekPrevious works backward in the buffer until it hits a rune in the `seek` category,
// ignoring runes in the `ignore` category. It returns true if `seek` is found.
func previous(seek uint16, data []byte) bool {
	return previousIndex(seek, data) >= 0
}

// seekForward looks ahead in the buffer until it hits a rune in the `seek` category,
// ignoring runes in the `ignore` category. It returns true if `seek` is found.
func forward(seek uint16, data []byte) bool {
	i := 0
	for i < len(data) {
		_, w := utf8.DecodeRune(data[i:])

		if is(_Ignore, data[i:]) {
			i += w
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
