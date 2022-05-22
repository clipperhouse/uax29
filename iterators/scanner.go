package iterators

import (
	"bufio"
	"io"

	"github.com/clipperhouse/uax29/iterators/filter"
)

type s = bufio.Scanner

type Scanner struct {
	s
	filters []filter.Func
}

// NewScanner creates a new Scanner given a SplitFunc. To use the new scanner,
// call SetText() and then iterate while Next() is true.
func NewScanner(r io.Reader, split bufio.SplitFunc) *Scanner {
	sc := &Scanner{
		s: *bufio.NewScanner(r),
	}
	sc.s.Split(split)
	return sc
}

// Filters applies one or more filters to all tokens (segments), only returning those
// where all filters evaluate true.
func (sc *Scanner) Filters(f ...filter.Func) {
	sc.filters = f
}

func (sc *Scanner) Scan() bool {
	scan := true

outer:
	for scan {
		scan = sc.s.Scan()

		if !scan {
			break
		}

		for _, f := range sc.filters {
			if !f(sc.s.Bytes()) {
				continue outer
			}
		}

		return scan
	}

	return scan
}
