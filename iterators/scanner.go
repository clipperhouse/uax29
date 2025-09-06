package iterators

import (
	"bufio"
	"io"
)

type s = *bufio.Scanner

type Scanner struct {
	// underlying bufio.Scanner; Bytes, Err and other methods are overridden
	s
	// token overrides (hides) the token of the underlying bufio.Scanner
	token []byte
	err   error
}

// NewScanner creates a new Scanner given an io.Reader and bufio.SplitFunc. To use the new scanner,
// iterate while Scan() is true.
func NewScanner(r io.Reader, split bufio.SplitFunc) *Scanner {
	sc := &Scanner{
		s: bufio.NewScanner(r),
	}
	sc.s.Split(split)

	return sc
}

// Bytes returns the current token, which results from calling Scan.
func (sc *Scanner) Bytes() []byte {
	return sc.token
}

// Text returns the current token as a string, which results from calling Scan.
func (sc *Scanner) Text() string {
	return string(sc.token)
}

// Err returns any error that resulted from calling Scan.
func (sc *Scanner) Err() error {
	if sc.err != nil {
		return sc.err
	}

	return sc.s.Err()
}

// Scan advances to the next token. It returns true until end of data, or
// an error. Use Bytes() to retrieve the token, and be sure to check Err().
func (sc *Scanner) Scan() bool {
	if sc.err != nil {
		return false
	}

	for sc.s.Scan() {
		sc.token = sc.s.Bytes()
		return true
	}

	return false
}
