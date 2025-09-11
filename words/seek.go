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

const notfound = -1

// subsequent looks ahead in the buffer until it hits a rune in properties,
// ignoring runes with the _Ignore property per WB4
func subsequent(properties property, data []byte, atEOF bool) (advance int, more bool) {
	i := 0
	for i < len(data) {
		lookup, w := trie.lookup(data[i:])
		if w == 0 {
			if atEOF {
				// Nothing more to evaluate
				return notfound, false
			}
			// More to evaluate
			return 0, true
		}

		if lookup.is(_Ignore) {
			i += w
			continue
		}

		if lookup.is(properties) {
			// Found it
			return i, false
		}

		// If we see a non-ignored character that doesn't match,
		// the property is definitely not "immediately subsequent"
		return notfound, false
	}

	// If we reach here, we've only seen ignored characters or incomplete runes
	if atEOF {
		// Nothing more to evaluate
		return notfound, false
	}

	// Need more
	return 0, true
}
