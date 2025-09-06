// Package words implements Unicode word boundaries: https://unicode.org/reports/tr29/#Word_Boundaries
package words

import (
	"io"

	"github.com/clipperhouse/uax29/internal/iterators"
)

type Scanner struct {
	*iterators.Scanner
}

// NewScanner returns a Scanner, to tokenize words per https://unicode.org/reports/tr29/#Word_Boundaries.
// Iterate through words by calling Scan() until false, then check Err(). See also the bufio.Scanner docs.
func NewScanner(r io.Reader) *Scanner {
	sc := &Scanner{
		iterators.NewScanner(r, SplitFunc),
	}
	return sc
}

// Joiners sets runes that should be treated like word characters, where
// otherwsie words sill be split. See the [Joiners] type.
func (sc *Scanner) Joiners(j *Joiners) {
	sc.Split(j.splitFunc)
}
