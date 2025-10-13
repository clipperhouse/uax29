package words

import (
	"github.com/clipperhouse/stringish"
	"github.com/clipperhouse/uax29/v2/internal/iterators"
)

type Iterator[T stringish.Interface] struct {
	*iterators.Iterator[T]
}

var (
	splitFuncString = splitFunc[string]
	splitFuncBytes  = splitFunc[[]byte]
)

// FromString returns an iterator for the words in the input string.
// Iterate while Next() is true, and access the word via Value().
func FromString(s string) Iterator[string] {
	return Iterator[string]{
		iterators.New(splitFuncString, s),
	}
}

// FromBytes returns an iterator for the words in the input bytes.
// Iterate while Next() is true, and access the word via Value().
func FromBytes(b []byte) Iterator[[]byte] {
	return Iterator[[]byte]{
		iterators.New(splitFuncBytes, b),
	}
}
