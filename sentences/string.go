package sentences

import "github.com/clipperhouse/uax29/internal/iterators"

// StringIterator is an iterator for sentences. Iterate while Next() is
// true, and access the sentence via Text().
type StringIterator struct {
	*iterators.StringIterator
}

// FromString returns an iterator for the sentences in the input string.
// source text. Iterate while Next() is true, and access the sentence via
// Text().
func FromString(s string) *StringIterator {
	iter := &StringIterator{
		iterators.NewStringIterator(SplitFunc),
	}
	iter.SetText(s)
	return iter
}
