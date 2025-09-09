package words

// FromString returns an iterator for the words in the input string.
// Iterate while Next() is true, and access the word via Text().
func FromString(s string) *Iterator[string] {
	return NewIterator(s)
}
