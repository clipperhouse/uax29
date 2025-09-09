package words

import (
	"unicode/utf8"

	"github.com/clipperhouse/uax29/v2/internal/iterators"
)

// previousIndex works backward until it hits a rune in properties,
// ignoring runes with the _Ignore property (per WB4), and returns
// the index in data. It returns -1 if such a rune is not found.
func previousIndex[T iterators.Stringish](properties property, data T) int {
	// Create a trie instance for this type
	trie := &wordsTrie[T]{}

	// Start at the end of the buffer and move backwards
	i := len(data)
	for i > 0 {
		var w int
		switch any(data).(type) {
		case []byte:
			bytes := any(data).([]byte)
			_, w = utf8.DecodeLastRune(bytes[:i])
		case string:
			str := any(data).(string)
			_, w = utf8.DecodeLastRuneInString(str[:i])
		}

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
func previous[T iterators.Stringish](properties property, data T) bool {
	return previousIndex(properties, data) != -1
}

const notfound = -1

// subsequent looks ahead in the buffer until it hits a rune in properties,
// ignoring runes with the _Ignore property per WB4
func subsequent[T iterators.Stringish](properties property, data T, atEOF bool) (advance int, more bool) {
	// Create a trie instance for this type
	trie := &wordsTrie[T]{}

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
