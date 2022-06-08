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
	s
	filters []filter.Func

	r     io.Reader // gotta keep references to these for Transform, sigh
	split bufio.SplitFunc

	// generally, we defer to the Err of the underlying bufio.Scanner
	// sometimes we can have an err above that, see Transform and Err methods below
	err        error
	scanCalled bool
}

// NewScanner creates a new Scanner given an io.Reader and bufio.SplitFunc. To use the new scanner,
// iterate while Scan() is true. See also the bufio.Scanner docs.
func NewScanner(r io.Reader, split bufio.SplitFunc) *Scanner {
	sc := &Scanner{
		s: bufio.NewScanner(r),
	}
	sc.s.Split(split)

	// Keep references which may be needed for transformers
	sc.r = r
	sc.split = split

	return sc
}

func (sc *Scanner) Err() error {
	if sc.err != nil {
		return sc.err
	}

	return sc.s.Err()
}

// Filter applies one or more filters (predicates) to all tokens, only returning those
// where all filters evaluate true. Filters are applied after Transformers.
func (sc *Scanner) Filter(filters ...filter.Func) {
	sc.filters = filters
}

var ErrorScanCalled = errors.New("cannot call Transform after Scan has been called")

// Transform applies one or more transformers to all tokens (segments). Calling Transform will overwrite
// previous transformers, so call it once (it's variadic, you can add multiple). Transformers are
// applied before Filters.
//
// This method must be called (applied) before calling Scan. Calling Transform after Scan
// will result in an error.
func (sc *Scanner) Transform(transformers ...transform.Transformer) {
	if sc.scanCalled {
		// this will be checked on a future invocation of Scan
		sc.err = ErrorScanCalled
		return
	}

	// For transformers to work correctly, apply them to the upstream Reader
	t := transform.Chain(transformers...)
	r := transform.NewReader(sc.r, t)

	// Gotta swap out the underlying bufio.Scanner. A little risky.
	// See Scanner.scanCalled and Scanner.err for how we prevent misuse.
	sc.s = bufio.NewScanner(r)
	sc.s.Split(sc.split)
}

// Scan advances to the next token (segment). It returns true until end of data, or
// an error. Use Bytes() to retrieve the token.
func (sc *Scanner) Scan() bool {
	if sc.err != nil {
		// probably came from Transform above
		return false
	}

	sc.scanCalled = true

scan:
	for sc.s.Scan() {
		for _, f := range sc.filters {
			if !f(sc.Bytes()) {
				continue scan
			}
		}

		return true
	}

	return false
}
