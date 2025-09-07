package graphemes

import "github.com/clipperhouse/uax29/internal/iterators"

// StringIterator is an iterator for grapheme clusters. Iterate while Next() is
// true, and access the grapheme via Text().
type StringIterator struct {
	*iterators.StringIterator
}

// FromString returns an iterator for the grapheme clusters in the input
// string. Iterate while Next() is true, and access the grapheme via Text().
func FromString(s string) *StringIterator {
	iter := &StringIterator{
		iterators.NewStringIterator(SplitFunc),
	}
	iter.SetText(s)
	return iter
}
