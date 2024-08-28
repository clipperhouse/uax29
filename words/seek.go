package words

import "unicode/utf8"

// previousIndex works backward until it hits a rune in properties,
// ignoring runes with the _Ignore property (per WB4), and returns
// the index in data. It returns -1 if such a rune is not found.
func previousIndex(properties property, data []byte) int {
	// Start at the end of the buffer and move backwards
	i := len(data)
	for i > 0 {
		_, w := utf8.DecodeLastRune(data[:i])

		i -= w

		lookup, _ := trie.lookup(data[i:])
		// I think it's OK to elide width here; will fall through to break

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
// ignoring runes with the _Ignore property per WB4
func previous(properties property, data []byte) bool {
	return previousIndex(properties, data) != -1
}

// subsequent looks ahead in the buffer until it hits a rune in properties,
// ignoring runes with the _Ignore property per WB4
func subsequent(properties property, data []byte, atEOF bool) (found bool, more bool) {
	i := 0
	for i < len(data) {
		lookup, w := trie.lookup(data[i:])
		if w == 0 {
			if atEOF {
				// Nothing more to evaluate
				return false, false
			}
			// More to evaluate
			return false, true
		}

		if lookup.is(_Ignore) {
			i += w
			continue
		}

		if lookup.is(properties) {
			// Found it
			return true, false
		}

		// If we get this far, it's not there
		return false, false
	}

	if atEOF {
		// Nothing more to evaluate
		return false, false
	}
	// More to evaluate
	return false, true
}
