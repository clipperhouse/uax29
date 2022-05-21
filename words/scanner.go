// Package words implementes Unicode word boundaries: https://unicode.org/reports/tr29/#Word_Boundaries
package words

import (
	"bufio"
	"io"
)

// NewScanner returns a bufio.Scanner, to tokenize words per https://unicode.org/reports/tr29/#Word_Boundaries.
// Iterate through words by calling Scan() until false. See the bufio.Scanner docs for details.
func NewScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	scanner.Split(SplitFunc)
	return scanner
}
