package iterators

import (
	"bufio"
	"io"

	"github.com/clipperhouse/uax29/iterators/filter"
	"golang.org/x/text/transform"
)

type s = bufio.Scanner

type Scanner struct {
	s
	predicates []filter.Func

	r     io.Reader // gotta keep references to these for transformers, sigh
	split bufio.SplitFunc
}

// NewScanner creates a new Scanner given an io.Reader and bufio.SplitFunc. To use the new scanner,
// iterate while Scan() is true. See also the bufio.Scanner docs.
func NewScanner(r io.Reader, split bufio.SplitFunc) *Scanner {
	sc := &Scanner{
		s: *bufio.NewScanner(r),
	}
	sc.s.Split(split)

	// Keep references which may be needed for transformers
	sc.r = r
	sc.split = split

	return sc
}

// Filter applies one or more filters (predicates) to all tokens (segments), only returning those
// where all predicates evaluate true. Filters are applied after Transformers.
func (sc *Scanner) Filter(predicates ...filter.Func) {
	sc.predicates = predicates
}

// Transform applies one or more transformers to all tokens (segments). Calling Transform will overwrite
// previous transformers, so call it once (it's variadic, you can add multiple). Transformers are
// applied before Filters.
func (sc *Scanner) Transform(transformers ...transform.Transformer) {
	t := transform.Chain(transformers...)

	// gotta swap out the reader for transformers to work; it's ugly
	// TODO: ensure that Scan has not been called prior, or think of something cleaner
	r := transform.NewReader(sc.r, t)
	sc.s = *bufio.NewScanner(r)
	sc.s.Split(sc.split)
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
