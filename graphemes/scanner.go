// Package graphemes implements Unicode grapheme cluster boundaries: https://unicode.org/reports/tr29/#Grapheme_Cluster_Boundaries
package graphemes

import (
	"io"

	"github.com/clipperhouse/uax29/iterators"
)

// NewScanner returns a bufio.Scanner, to tokenize graphemes per https://unicode.org/reports/tr29/#Grapheme_Cluster_Boundaries.
// Iterate through graphemes by calling Scan() until false. See the bufio.Scanner docs for details.
func NewScanner(r io.Reader) *iterators.Scanner {
	scanner := iterators.NewScanner(r, SplitFunc)
	return scanner
}
