package words

import "github.com/clipperhouse/uax29/internal/iterators"

// StringIterator is an iterator for words. Iterate while Next() is
// true, and access the word via Text().
type StringIterator struct {
	*iterators.StringIterator
}

// FromString returns an iterator for the words in the input string.
// Iterate while Next() is true, and access the word via Text().
func FromString(s string) *StringIterator {
	iter := &StringIterator{
		iterators.NewStringIterator(SplitFunc),
	}
	iter.SetText(s)
	return iter
}
