package graphemes

import "unicode/utf8"

// previous works backward in the buffer until it hits a rune in categories,
// ignoring runes in the _Ignore category.
func previous(categories uint16, data []byte) bool {
	// Start at the end of the buffer and move backwards
	i := len(data)
	for i > 0 {
		_, w := utf8.DecodeLastRune(data[:i])
		i -= w

		if is(_Ignore, data[i:]) {
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
