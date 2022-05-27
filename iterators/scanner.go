package iterators

import (
	"bufio"
	"io"

	"github.com/clipperhouse/uax29/iterators/filter"
	"github.com/clipperhouse/uax29/iterators/transform"
)

type s = bufio.Scanner

type Scanner struct {
	s
	predicates []filter.Func
	transforms []transform.Func
}

// NewScanner creates a new Scanner given an io.Reader and bufio.SplitFunc. To use the new scanner,
// iterate while Scan() is true. See also the bufio.Scanner docs.
func NewScanner(r io.Reader, split bufio.SplitFunc) *Scanner {
	sc := &Scanner{
		s: *bufio.NewScanner(r),
	}
	sc.s.Split(split)
	return sc
}

// Filter applies one or more filters (predicates) to all tokens (segments), only returning those
// where all predicates evaluate true.
func (sc *Scanner) Filter(predicates ...filter.Func) {
	sc.predicates = predicates
}

// Transform applies one or more transforms to all tokens (segments). Calling Transform will overwrite
// previous transforms, so call it once (it's variadic, you can add multiple).
func (sc *Scanner) Transform(transforms ...transform.Func) {
	sc.transforms = transforms
}

// Bytes returns the current token (segment).
func (sc *Scanner) Bytes() []byte {
	b := sc.s.Bytes()
	for _, t := range sc.transforms {
		b = t(b)
	}
	return b
}

// Scan advances to the next token (segment). It returns true until end of data, or
// an error. Use Bytes() to retrieve the token.
func (sc *Scanner) Scan() bool {
	scan := true

outer:
	for scan {
		scan = sc.s.Scan()
		if !scan {
			break
		}

		for _, f := range sc.predicates {
			if !f(sc.Bytes()) {
				continue outer
			}
		}

		return scan
	}

	return scan
}
