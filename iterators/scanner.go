package iterators

import (
	"bufio"
	"io"
	"unicode"

	"github.com/clipperhouse/uax29/iterators/filter"
	"github.com/clipperhouse/uax29/iterators/util"
)

type s = bufio.Scanner

type Scanner struct {
	s
	filters []filter.Func
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
			if !f(sc.Bytes()) {
				continue outer
			}
		}

		return scan
	}

	return scan
}

// Contains indicates that the current token (segment) contains one or more runes
// that are in one or more of the ranges.
func (sc *Scanner) Contains(ranges ...*unicode.RangeTable) bool {
	return util.Contains(sc.Bytes(), ranges...)
}

// Entirely indicates that the current token (segment) consists entirely of
// runes that are in one or more of the ranges.
func (sc *Scanner) Entirely(ranges ...*unicode.RangeTable) bool {
	return util.Entirely(sc.Bytes(), ranges...)
}

// Is indicates that the current token (segment) evaluates to true
// for all given filters.
func (sc *Scanner) Is(filters ...filter.Func) bool {
	for _, f := range filters {
		if !f(sc.Bytes()) {
			return false
		}
	}

	return true
}
