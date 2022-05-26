// Package sentences implements Unicode sentence boundaries: https://unicode.org/reports/tr29/#Sentence_Boundaries
package sentences

import (
	"io"

	"github.com/clipperhouse/uax29/iterators"
)

// NewScanner returns a bufio.Scanner, to tokenize sentences per https://unicode.org/reports/tr29/#Sentence_Boundaries.
// Iterate through sentences by calling Scan() until false. See also the bufio.Scanner docs.
func NewScanner(r io.Reader) *iterators.Scanner {
	sc := iterators.NewScanner(r, SplitFunc)
	return sc
}
