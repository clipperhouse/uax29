package uax29

import (
	"bufio"
	"bytes"
	"io"
)

// BreakFunc defines a func that indicates when to break tokens.
// It is called for every rune in the incoming reader. The current rune is buffer[pos].
// Returing true indicates break at the current rune, i.e. it begins a new token;
// false indicates accept the current rune and continue.
type BreakFunc func(buffer Runes, pos Pos) bool

// NewScanner instantiates a new Scanner
func NewScanner(r io.Reader, breaking BreakFunc) *Scanner {
	return &Scanner{
		incoming: bufio.NewReaderSize(r, 64*1024),
		breaking: breaking,
	}
}

// Scanner is a structure for scanning an input Reader. Use NewScanner to instantiate.
// Loop over scanner.Scan while true.
type Scanner struct {
	incoming *bufio.Reader
	breaking BreakFunc

	// a buffer of runes to evaluate
	buffer Runes
	// a cursor for runes in the buffer
	pos Pos

	// outputs
	bytes []byte
	err   error
}

// Scan advances to the next token, returning true if successful. Returns false on error or EOF.
// Use Bytes or Text to retrieve the token value, or Err to retrieve the current error.
func (sc *Scanner) Scan() bool {
	sc.reset()

	for {
		// Fill the buffer with enough runes for lookahead
		for len(sc.buffer) < int(sc.pos)+8 {
			r, eof, err := sc.readRune()
			if err != nil {
				sc.err = err
				return false
			}
			if eof {
				break
			}
			sc.buffer = append(sc.buffer, r)
		}

		// The current rune represents a new token
		if sc.breaking(sc.buffer, sc.pos) {
			break
		}

		// Accept the current rune and continue
		sc.pos++
	}

	return sc.token()
}

// Bytes returns the current token as a byte slice, after a successful call to Scan
func (sc *Scanner) Bytes() []byte {
	return sc.bytes
}

// Text returns the current token, after a successful call to Scan
func (sc *Scanner) Text() string {
	return string(sc.bytes)
}

// Err returns the current error, after an unsuccessful call to Scan
func (sc *Scanner) Err() error {
	return sc.err
}

// reset creates a new bytes.Buffer on the Scanner, and clears previous values
func (sc *Scanner) reset() {
	// Drop the emitted runes (optimization to avoid growing array)
	copy(sc.buffer, sc.buffer[sc.pos:])
	sc.buffer = sc.buffer[:len(sc.buffer)-int(sc.pos)]

	sc.pos = 0

	sc.bytes = nil
	sc.err = nil
}

func (sc *Scanner) token() bool {
	var bb bytes.Buffer
	for _, r := range sc.buffer[:sc.pos] {
		bb.WriteRune(r)
	}
	sc.bytes = bb.Bytes()
	return len(sc.bytes) > 0
}

// readRune gets the next rune, advancing the reader
func (sc *Scanner) readRune() (r rune, eof bool, err error) {
	r, _, err = sc.incoming.ReadRune()

	if err != nil {
		if err == io.EOF {
			return r, true, nil
		}
		return r, false, err
	}

	return r, false, nil
}
