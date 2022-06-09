package iterators

import (
	"bufio"
	"errors"
	"io"

	"github.com/clipperhouse/uax29/iterators/filter"
	"golang.org/x/text/transform"
)

type s = *bufio.Scanner

type Scanner struct {
	// underlying bufio.Scanner; Bytes, Err and other methods are overridden
	s
	// token overrides (hides) the token of the underlying bufio.Scanner
	token       []byte
	filter      filter.Func
	transformer transform.Transformer
	err         error
}

// NewScanner creates a new Scanner given an io.Reader and bufio.SplitFunc. To use the new scanner,
// iterate while Scan() is true.
func NewScanner(r io.Reader, split bufio.SplitFunc) *Scanner {
	sc := &Scanner{
		s: bufio.NewScanner(r),
	}
	sc.s.Split(split)

	return sc
}

// Bytes returns the current token, which results from calling Scan.
func (sc *Scanner) Bytes() []byte {
	return sc.token
}

// Text returns the current token as a string, which results from calling Scan.
func (sc *Scanner) Text() string {
	return string(sc.token)
}

// Err returns any error that resulted from calling Scan.
func (sc *Scanner) Err() error {
	if sc.err != nil {
		return sc.err
	}

	return sc.s.Err()
}

// Filter applies one or more filters (predicates) to all tokens, only returning those
// where all filters evaluate true. Filters are applied after Transformers.
func (sc *Scanner) Filter(filter filter.Func) {
	sc.filter = filter
}

var ErrorScanCalled = errors.New("cannot call Transform after Scan has been called")

// Transform applies one or more transformers to all tokens, in order. Calling Transform overwrites
// previous transformers, so call it once (it's variadic, you can add multiple). Transformers are
// applied before Filters.
func (sc *Scanner) Transform(transformers ...transform.Transformer) {
	sc.transformer = transform.Chain(transformers...)
}

// Scan advances to the next token. It returns true until end of data, or
// an error. Use Bytes() to retrieve the token, and be sure to check Err().
func (sc *Scanner) Scan() bool {
	if sc.err != nil {
		return false
	}

scan:
	for sc.s.Scan() {
		sc.token = sc.s.Bytes()

		if sc.transformer != nil {
			sc.token, _, sc.err = transform.Bytes(sc.transformer, sc.token)
			if sc.err != nil {
				return false
			}
		}

		if sc.filter != nil && !sc.filter(sc.Bytes()) {
			continue scan
		}

		return true
	}

	return false
}
