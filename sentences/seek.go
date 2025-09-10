package sentences

import (
	"unicode/utf8"

	"github.com/clipperhouse/uax29/v2/internal/iterators"
)

func decodeLastRune[T iterators.Stringish](data T) (rune, int) {
	// This casting is a bit gross but it works
	// and is surprisingly fast
	switch any(data).(type) {
	case []byte:
		b := any(data).([]byte)
		return utf8.DecodeLastRune(b)
	case string:
		s := any(data).(string)
		return utf8.DecodeLastRuneInString(s)
	default:
		panic("unsupported type")
	}
}

// previousIndex works backward until it hits a rune in properties,
// ignoring runes in the _Ignore property (per SB5), and returns
// the index of the rune in data. It returns -1 if such a rune is not found.
func previousIndex[T iterators.Stringish](properties property, data T) int {
	// Start at the end of the buffer and move backwards

	i := len(data)
	for i > 0 {
		_, w := decodeLastRune(data[:i])
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
func previous[T iterators.Stringish](properties property, data T) bool {
	return previousIndex(properties, data) != -1
}

// subsequent looks ahead in the buffer until it hits a rune in properties,
// ignoring runes in the _Ignore property per SB5
func subsequent[T iterators.Stringish](properties property, data T, atEOF bool) (found bool, pos int, requestMore bool) {
	i := 0
	for i < len(data) {
		lookup, w := lookup(data[i:])
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
