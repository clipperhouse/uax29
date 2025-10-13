package sentences

import (
	"github.com/clipperhouse/stringish"
	"github.com/clipperhouse/stringish/utf8"
)

const notfound = -1

// previousIndex works backward until it hits a rune in properties,
// ignoring runes in the _Ignore property (per SB5), and returns
// the index of the rune in data. It returns -1 if such a rune is not found.
func previousIndex[T stringish.Interface](properties property, data T) int {
	// Start at the end of the buffer and move backwards

	i := len(data)
	for i > 0 {
		_, w := utf8.DecodeLastRune(data[:i])
		i -= w

		lookup, _ := lookup(data[i:])

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
// ignoring runes with the _Ignore property per SB5
func previous[T stringish.Interface](properties property, data T) bool {
	return previousIndex(properties, data) != -1
}

// subsequent looks ahead in the buffer until it hits a rune in properties,
// ignoring runes with the _Ignore property per SB5
func subsequent[T stringish.Interface](properties property, data T, atEOF bool) (advance int, more bool) {
	i := 0
	for i < len(data) {
		lookup, w := lookup(data[i:])
		if w == 0 {
			if atEOF {
				// Nothing more to evaluate
				return notfound, false
			}
			// More to evaluate - return notfound to indicate no match found yet
			return notfound, true
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

	// Need more - return notfound to indicate no match found yet
	return notfound, true
}
