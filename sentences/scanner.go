// Package sentences implements Unicode sentence boundaries: https://unicode.org/reports/tr29/#Sentence_Boundaries
package sentences

import (
	"bufio"
	"io"
)

// NewScanner returns a bufio.Scanner, to tokenize sentences per https://unicode.org/reports/tr29/#Sentence_Boundaries.
// Iterate through sentences by calling Scan() until false. See the bufio.Scanner docs for details.
func NewScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	scanner.Split(SplitFunc)
	return scanner
}
