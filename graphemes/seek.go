package graphemes

import "unicode/utf8"

// previous works backward in the buffer until it hits a rune in properties,
// ignoring runes in the _Ignore property.
func previous(properties property, data []byte) bool {
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
			return true
		}

		// If we get this far, it's not there
		break
	}

	return false
}
