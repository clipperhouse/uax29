package words

// FromBytes returns an iterator for the words in the input bytes.
// Iterate while Next() is true, and access the word via Bytes().
func FromBytes(b []byte) *Iterator[[]byte] {
	return NewIterator(b)
}
