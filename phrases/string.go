package phrases

import "github.com/clipperhouse/uax29/v2/internal/iterators"

// StringIterator is an iterator for phrases. Iterate while Next() is
// true, and access the phrase via Text().
type StringIterator struct {
	*iterators.StringIterator
}

// FromString returns an iterator for the phrases in the input string.
// Iterate while Next() is true, and access the phrase via Text().
func FromString(s string) *StringIterator {
	iter := &StringIterator{
		iterators.NewStringIterator(SplitFunc),
	}
	iter.SetText(s)
	return iter
}
