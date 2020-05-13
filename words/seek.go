package words

import "unicode/utf8"

// previousIndex works backward until it hits a rune in categories,
// ignoring runes in the _Ignore category, and returns the index of the rune in the buffer.
// It returns -1 if a categories rune is not found.
func previousIndex(categories uint32, data []byte) int {
	// Start at the end of the buffer and move backwards
	i := len(data)
	for i > 0 {
		_, w := utf8.DecodeLastRune(data[:i])
		i -= w

		if is(_Ignore, data[i:]) {
			continue
		}

		if is(categories, data[i:]) {
			return i
		}

		// If we get this far, it's not there
		break
	}

	return -1
}

// previous works backward in the buffer until it hits a rune in categories,
// ignoring runes in the _Ignore category.
func previous(categories uint32, data []byte) bool {
	return previousIndex(categories, data) >= 0
}

// forward looks ahead in the buffer until it hits a rune in categories,
// ignoring runes in the _Ignore category.
func forward(categories uint32, data []byte) bool {
	i := 0
	for i < len(data) {
		_, w := utf8.DecodeRune(data[i:])

		if is(_Ignore, data[i:]) {
			i += w
			continue
		}

		if is(categories, data[i:]) {
			return true
		}

		// If we get this far, it's not there
		break
	}

	return false
}
