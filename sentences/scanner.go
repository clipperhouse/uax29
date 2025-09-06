// Package sentences implements Unicode sentence boundaries: https://unicode.org/reports/tr29/#Sentence_Boundaries
package sentences

import (
	"bufio"
	"io"
)

type Scanner struct {
	*bufio.Scanner
}

// NewScanner returns a Scanner, to split sentences per
// https://unicode.org/reports/tr29/#Sentence_Boundaries.
//
// It embeds a [bufio.Scanner], so you can use its methods.
//
// Iterate through sentences by calling Scan() until false, then check Err().
func NewScanner(r io.Reader) *Scanner {
	sc := bufio.NewScanner(r)
	sc.Split(SplitFunc)
	return &Scanner{
		Scanner: sc,
	}
}
