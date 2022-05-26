package iterators

import (
	"bufio"
	"io"

	"github.com/clipperhouse/uax29/iterators/filter"
)

type s = bufio.Scanner

type Scanner struct {
	s
	predicates []filter.Predicate
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
func (sc *Scanner) Filter(predicates ...filter.Predicate) {
	sc.predicates = predicates
}

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
