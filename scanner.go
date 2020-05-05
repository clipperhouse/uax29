package uax29

import (
	"bufio"
	"fmt"
	"io"
	"sync"
	"unicode/utf8"
)

// BreakFunc defines a func that indicates when to break tokens.
// It is called for every rune in the incoming reader. The current rune is buffer[pos].
// Returing true indicates break at the current rune, i.e. it begins a new token;
// false indicates accept the current rune and continue.
type BreakFunc func(buffer Runes, pos Pos) bool

// NewScanner instantiates a new Scanner
func NewScanner(r io.Reader, breakFunc BreakFunc) *Scanner {
	return &Scanner{
		incoming:  bufio.NewReaderSize(r, 64*1024),
		breakFunc: breakFunc,
	}
}

// Break is the result of a BreakFunc, signaling to break at (before) the current rune
const Break = true

// Accept is the result of a BreakFunc, signaling to accept the current rune and continue
const Accept = false

const lookahead = 8

// Scanner is a structure for scanning an input Reader. Use NewScanner to instantiate.
// Loop over scanner.Scan while true.
type Scanner struct {
	incoming  *bufio.Reader
	breakFunc BreakFunc

	// a buffer of runes to evaluate
	buffer Runes

	// underlying store for outgoing tokens; an optimization to avoid allocation; see token()
	results []byte

	// outputs
	bytes []byte
	err   error
}

// Scan advances to the next token, returning true if successful. Returns false on error or EOF.
// Use Bytes or Text to retrieve the token value, or Err to retrieve the current error.
func (sc *Scanner) Scan() bool {
	// Reset state
	sc.bytes = nil
	sc.err = nil

	// A cursor for runes in the buffer
	var pos Pos

	for {
		// Fill the buffer with enough runes for lookahead
		for len(sc.buffer) < int(pos)+lookahead {
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

		if sc.breakFunc(sc.buffer, pos) == Break {
			// The current rune represents a new token
			break
		}

		// Otherwise, accept the current rune and continue
		pos++
	}

	// Create the token
	sc.bytes = sc.token(pos)

	// Drop the emitted runes (with optimization to avoid growing array)
	copy(sc.buffer, sc.buffer[pos:])
	sc.buffer = sc.buffer[:len(sc.buffer)-int(pos)]

	return len(sc.bytes) > 0
}

const arena = 1024 // 1k underlying store by default

// token returns the current token runes as bytes
func (sc *Scanner) token(pos Pos) []byte {
	// This might be an arena allocator?
	// Use a shared underlying store for tokens, at a fixed size.
	// Allocate a new one when we run out of space.
	// "Immutable" in that we don't resize or mutate any bytes after writing.

	// Count the bytes
	n := 0
	for _, r := range sc.buffer[:pos] {
		n += utf8.RuneLen(r)
	}

	// Ensure underlying store is big enough; don't resize
	if len(sc.results) < n {
		// At least as big as the current token
		size := arena
		if size < n {
			size = n
		}
		sc.results = make([]byte, size, size)
	}

	// Write runes into the store
	i := 0
	for _, r := range sc.buffer[:pos] {
		w := utf8.EncodeRune(sc.results[i:], r)
		i += w
	}

	// Grab the bytes
	token := sc.results[:n]
	// Reset the store
	sc.results = sc.results[n:]

	return token
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

var runesPool = sync.Pool{
	New: func() interface{} {
		var runes Runes
		return runes
	},
}

// NewSplitFunc creates a new bufio.SplitFunc, based on a uax29.BreakFunc
func NewSplitFunc(breakFunc BreakFunc) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		buffer := runesPool.Get().(Runes)[:0]
		var pos Pos

		i := 0

		for {
			// Fill the buffer with enough runes for lookahead
			for i < len(data) && len(buffer) < lookahead {
				r, w := utf8.DecodeRune(data[i:])
				if r == utf8.RuneError {
					return 0, nil, fmt.Errorf("error decoding rune")
				}
				i += w
				buffer = append(buffer, r)
			}

			if len(buffer) < lookahead {
				if !atEOF {
					// Need to request more data
					return 0, nil, nil
				}
			}

			if breakFunc(buffer, pos) == Break {
				// The current rune represents a new token
				break
			}

			// Otherwise, accept the current rune and continue
			pos++
		}

		// Count the bytes
		n := 0
		for _, r := range buffer[:pos] {
			n += utf8.RuneLen(r)
		}

		runesPool.Put(buffer)

		return n, data[:n], nil
	}
}