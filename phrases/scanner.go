// Package phrases implements Unicode phrase boundaries: https://unicode.org/reports/tr29/#phrase_Boundaries
package phrases

import (
	"io"

	"github.com/clipperhouse/uax29/iterators"
)

// NewScanner returns a Scanner, to tokenize phrases per https://unicode.org/reports/tr29/#phrase_Boundaries.
// Iterate through phrases by calling Scan() until false, then check Err(). See also the bufio.Scanner docs.
func NewScanner(r io.Reader) *iterators.Scanner {
	sc := iterators.NewScanner(r, SplitFunc)
	return sc
}
