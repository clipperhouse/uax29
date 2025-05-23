package phrases

// subsequent looks ahead in the buffer until it hits a rune in properties,
// ignoring runes with the _Ignore property per WB4
func subsequent(properties property, data []byte, atEOF bool) (found bool, pos int, more bool) {
	i := 0
	for i < len(data) {
		lookup, w := trie.lookup(data[i:])
		if w == 0 {
			if atEOF {
				// Nothing more to evaluate
				return false, 0, false
			}
			// More to evaluate
			return false, 0, true
		}

		if lookup.is(_Ignore) {
			i += w
			continue
		}

		if lookup.is(properties) {
			// Found it
			return true, i + w, false
		}

		// If we get this far, it's not immediately subsequent
		return false, 0, false
	}

	// If not eof, we need more
	return false, 0, !atEOF
}
