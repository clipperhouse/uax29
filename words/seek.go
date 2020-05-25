package words

import "unicode/utf8"

// previousIndex works backward until it hits a rune in properties,
// ignoring runes in the _Ignore property, and returns the index of the rune in the buffer.
// It returns -1 if a properties rune is not found.
func previousIndex(properties property, data []byte) int {
	// Start at the end of the buffer and move backwards
	i := len(data)
	for i > 0 {
		_, w := utf8.DecodeLastRune(data[:i])
		i -= w

		lookup, _ := trie.lookup(data[i:])

		if lookup.is(_Ignore) {
			continue
		}

		if lookup.is(properties) {
			return i
		}

		// If we get this far, it's not there
		break
	}

	return -1
}

// previous works backward in the buffer until it hits a rune in properties,
// ignoring runes in the _Ignore property.
func previous(properties property, data []byte) bool {
	return previousIndex(properties, data) != -1
}

// subsequent looks ahead in the buffer until it hits a rune in properties,
// ignoring runes in the _Ignore property.
func subsequent(properties property, data []byte) bool {
	i := 0
	for i < len(data) {
		lookup, w := trie.lookup(data[i:])

		if lookup.is(_Ignore) {
			i += w
			continue
		}

		if lookup.is(properties) {
			return true
		}

		// If we get this far, it's not there
		break
	}

	return false
}
