// Package words implements Unicode word boundaries: https://unicode.org/reports/tr29/#Word_Boundaries
package words

import (
	"io"

	"github.com/clipperhouse/uax29/iterators"
)

// NewScanner Scanner, to tokenize words per https://unicode.org/reports/tr29/#Word_Boundaries.
// Iterate through words by calling Scan() until false, then check Err(). See also the bufio.Scanner docs.
func NewScanner(r io.Reader) *iterators.Scanner {
	sc := iterators.NewScanner(r, SplitFunc)
	return sc
}
