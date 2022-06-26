// Package words implements Unicode word boundaries: https://unicode.org/reports/tr29/#Word_Boundaries
package words

import (
	"bufio"
	"io"

	"github.com/clipperhouse/uax29/iterators"
)

// NewScanner returns a Scanner, to tokenize words per https://unicode.org/reports/tr29/#Word_Boundaries.
// Iterate while Scan() is true, access the segmented word via Bytes(), and check Err().
func NewScanner(r io.Reader) *iterators.Scanner {
	sc := iterators.NewScanner(r, SplitFunc)
	return sc
}

// NewScannerWeb returns a Scanner, which is an iterator over the source text.
// It joins tokens on 'web' characters such as '@' (for email addresses and handles),
// and '#' (for hashtags), and many characters for URLs (such as '/', '?', etc).
// The basic scanner would treat these as scanner tokens; this web segmenter will
// join them into a single token.
//
// It is fairly naive, in that it makes no attempt to validate (for example) email
// addresses or URLs. It simply treats the above-mentioned characters as alphanumeric.
//
// Iterate while Scan() is true, access the segmented word via Bytes(), and check Err().
func NewScannerWeb(r io.Reader) *iterators.Scanner {
	opts := options{Web: true}
	var split bufio.SplitFunc = func(data []byte, atEOF bool) (int, []byte, error) {
		return splitFuncOpts(data, atEOF, opts)
	}

	sc := iterators.NewScanner(r, split)
	return sc
}
