package seeker

import (
	"unicode/utf8"
)

// Seeker is a structure for looking forward or back in a byte slice, for categories
type Seeker struct {
	lookup func([]byte) (uint32, int)
	ignore uint32
}

// New creates a Seeker for looking forward or back in a byte slice, for categories
func New(lookup func([]byte) (uint32, int), ignore uint32) *Seeker {
	return &Seeker{
		lookup: lookup,
		ignore: ignore,
	}
}

// Is determines if the current rune matches categories
func (sk *Seeker) Is(categories uint32, s []byte) bool {
	lookup, _ := sk.lookup(s)
	return (lookup & categories) != 0
}

// PreviousIndex works backward until it hits a rune in the `seek` category,
// ignoring runes in the `ignore` category, and returns the index of the rune in the buffer.
// It returns -1 if `seek` rune is not found.
func (sk *Seeker) PreviousIndex(seek uint32, data []byte) int {
	// Start at the end of the buffer and move backwards
	i := len(data)
	for i > 0 {
		r, w := utf8.DecodeLastRune(data[:i])
		i -= w

		_ = r

		if sk.Is(sk.ignore, data[i:]) {
			continue
		}

		if sk.Is(seek, data[i:]) {
			return i
		}

		// If we get this far, it's not there
		break
	}

	return -1
}

// Previous works backward in the buffer until it hits a rune in the `seek` category,
// ignoring runes in the `ignore` category. It returns true if `seek` is found.
func (sk *Seeker) Previous(seek uint32, data []byte) bool {
	return sk.PreviousIndex(seek, data) >= 0
}

// Forward looks ahead in the buffer until it hits a rune in the `seek` category,
// ignoring runes in the `ignore` category. It returns true if `seek` is found.
func (sk *Seeker) Forward(seek uint32, data []byte) bool {
	i := 0
	for i < len(data) {
		_, w := utf8.DecodeRune(data[i:])

		if sk.Is(sk.ignore, data[i:]) {
			i += w
			continue
		}

		if sk.Is(seek, data[i:]) {
			return true
		}

		// If we get this far, it's not there
		break
	}

	return false
}
