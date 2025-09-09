package words

// Generic iterator implementation

// Iterator is a generic iterator for words that are either []byte or string.
// Iterate while Next() is true, and access the word via Value().
type Iterator[T stringish] struct {
	split func(T, bool) (int, T, error)
	data  T
	pos   int
	start int
	token T
}

// NewIterator creates a new Iterator for the given data and SplitFunc.
func NewIterator[T stringish](data T) *Iterator[T] {
	// Create a joiners instance for this type
	joiners := &Joiners[T]{}

	iter := &Iterator[T]{
		split: joiners.splitFunc,
		data:  data,
	}
	return iter
}

// SetText sets the text for the iterator to operate on, and resets all state.
func (iter *Iterator[T]) SetText(data T) {
	iter.data = data
	iter.pos = 0
	iter.start = 0
	var empty T
	iter.token = empty
}

// Split sets the SplitFunc for the Iterator.
func (iter *Iterator[T]) Split(split func(T, bool) (int, T, error)) {
	iter.split = split
}

// Next advances the iterator to the next token. It returns false when there
// are no remaining tokens or an error occurred.
func (iter *Iterator[T]) Next() bool {
	if iter.pos == len(iter.data) {
		return false
	}
	if iter.pos > len(iter.data) {
		panic("SplitFunc advanced beyond the end of the data")
	}

	iter.start = iter.pos

	advance, token, err := iter.split(iter.data[iter.pos:], true)
	if err != nil {
		panic(err)
	}
	if advance <= 0 {
		panic("SplitFunc returned a zero or negative advance")
	}

	iter.pos += advance
	if iter.pos > len(iter.data) {
		panic("SplitFunc advanced beyond the end of the data")
	}

	iter.token = token

	return true
}

// Value returns the current token.
func (iter *Iterator[T]) Value() T {
	return iter.token
}

// Start returns the byte position of the current token in the original data.
func (iter *Iterator[T]) Start() int {
	return iter.start
}

// End returns the byte position after the current token in the original data.
func (iter *Iterator[T]) End() int {
	return iter.pos
}

// Reset resets the iterator to the beginning of the data.
func (iter *Iterator[T]) Reset() {
	iter.pos = 0
	iter.start = 0
	var empty T
	iter.token = empty
}
