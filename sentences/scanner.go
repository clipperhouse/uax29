// Package sentences provides a scanner for Unicode text segmentation sentence boundaries: https://unicode.org/reports/tr29/#Sentence_Boundaries
package sentences

import (
	"bufio"
	"bytes"
	"io"
	"unicode"
)

// NewScanner tokenizes a reader into a stream of sentence tokens according to Unicode Text Segmentation sentence boundaries https://unicode.org/reports/tr29/#Sentence_Boundaries
// Iterate through the stream by calling Scan() until false.
//	text := "This is an example. And another!"
//	reader := strings.NewReader(text)
//
//	scanner := sentences.NewScanner(reader)
//	for scanner.Scan() {
//		fmt.Printf("%s\n", scanner.Text())
//	}
//	if err := scanner.Err(); err != nil {
//		log.Fatal(err)
//	}
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		incoming: bufio.NewReaderSize(r, 64*1024),
	}
}

// Scanner is the structure for scanning an input Reader. Use NewScanner to instantiate.
type Scanner struct {
	incoming *bufio.Reader

	// a buffer of runes to evaluate
	buffer []rune
	// a cursor for runes in the buffer
	pos int

	bb bytes.Buffer

	// outputs
	bytes []byte
	err   error
}

// reset creates a new bytes.Buffer on the Scanner, and clears previous values
func (sc *Scanner) reset() {
	// Drop the emitted runes (optimization to avoid growing array)
	copy(sc.buffer, sc.buffer[sc.pos:])
	sc.buffer = sc.buffer[:len(sc.buffer)-sc.pos]

	sc.pos = 0

	var bb bytes.Buffer
	sc.bb = bb

	sc.bytes = nil
	sc.err = nil
}

// Scan advances to the next token, returning true if successful. Returns false on error or EOF.
func (sc *Scanner) Scan() bool {
	sc.reset()

	for {
		// Fill the buffer with enough runes for lookahead
		for len(sc.buffer) < sc.pos+8 {
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

		// SB1
		sot := sc.pos == 0 // "start of text"
		eof := len(sc.buffer) == sc.pos
		if sot && !eof {
			sc.accept()
			continue
		}

		// SB2
		if eof {
			break
		}

		current := sc.buffer[sc.pos]
		previous := sc.buffer[sc.pos-1]

		// SB3
		if is(LF, current) && is(CR, previous) {
			sc.accept()
			continue
		}

		// SB4
		if is(_mergedParaSep, previous) {
			break
		}

		// SB5
		if is(_mergedExtendFormat, current) {
			sc.accept()
			continue
		}

		// SB6
		if is(Numeric, current) && sc.seekPrevious(sc.pos, ATerm) {
			sc.accept()
			continue
		}

		// SB7
		if is(Upper, current) {
			previousIndex := sc.seekPreviousIndex(sc.pos, ATerm)
			if previousIndex >= 0 && sc.seekPrevious(previousIndex, _mergedUpperLower) {
				sc.accept()
				continue
			}
		}

		// SB8
		{
			// This loop is the 'regex':
			// ( ¬(OLetter | Upper | Lower | ParaSep | SATerm) )*
			pos := sc.pos
			for pos < len(sc.buffer) {
				current := sc.buffer[pos]
				if is(_mergedOLetterUpperLowerParaSepSATerm, current) {
					break
				}
				pos++
			}

			if sc.seekForward(pos-1, Lower) {
				pos := sc.pos

				sp := pos
				for {
					sp = sc.seekPreviousIndex(sp, Sp)
					if sp < 0 {
						break
					}
					pos = sp
				}

				close := pos
				for {
					close = sc.seekPreviousIndex(close, Close)
					if close < 0 {
						break
					}
					pos = close
				}

				if sc.seekPrevious(pos, ATerm) {
					sc.accept()
					continue
				}
			}
		}

		// SB8a
		if is(_mergedSContinueSATerm, current) {
			pos := sc.pos

			sp := pos
			for {
				sp = sc.seekPreviousIndex(sp, Sp)
				if sp < 0 {
					break
				}
				pos = sp
			}

			close := pos
			for {
				close = sc.seekPreviousIndex(close, Close)
				if close < 0 {
					break
				}
				pos = close
			}

			if sc.seekPrevious(pos, _mergedSATerm) {
				sc.accept()
				continue
			}
		}

		// SB9
		if is(_mergedCloseSpParaSep, current) {
			pos := sc.pos

			close := pos
			for {
				close = sc.seekPreviousIndex(close, Close)
				if close < 0 {
					break
				}
				pos = close
			}

			if sc.seekPrevious(pos, _mergedSATerm) {
				sc.accept()
				continue
			}
		}

		// SB10
		if is(_mergedSpParaSep, current) {
			pos := sc.pos

			sp := pos
			for {
				sp = sc.seekPreviousIndex(sp, Sp)
				if sp < 0 {
					break
				}
				pos = sp
			}

			close := pos
			for {
				close = sc.seekPreviousIndex(close, Close)
				if close < 0 {
					break
				}
				pos = close
			}

			if sc.seekPrevious(pos, _mergedSATerm) {
				sc.accept()
				continue
			}
		}

		// SB11
		{
			pos := sc.pos

			ps := sc.seekPreviousIndex(pos, _mergedSpParaSep)
			if ps >= 0 {
				pos = ps
			}

			sp := pos
			for {
				sp = sc.seekPreviousIndex(sp, Sp)
				if sp < 0 {
					break
				}
				pos = sp
			}

			close := pos
			for {
				close = sc.seekPreviousIndex(close, Close)
				if close < 0 {
					break
				}
				pos = close
			}

			if sc.seekPrevious(pos, _mergedSATerm) {
				break
			}
		}

		// SB998
		if sc.pos > 0 {
			sc.accept()
			continue
		}

		// If we fall through all the above rules, it's a sentence break
		break
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

// Sentence boundary rules: https://unicode.org/reports/tr29/#Sentence_Boundaries
// In most cases, returning true means 'keep going'; check the name of the return var for clarity

var is = unicode.Is

// seekForward looks ahead until it hits a rune satisfying one of the range tables,
// ignoring Extend|Format
// See: https://unicode.org/reports/tr29/#Grapheme_Cluster_and_Format_Rules (driven by SB5)
func (sc *Scanner) seekForward(pos int, rts ...*unicode.RangeTable) bool {
	for i := pos + 1; i < len(sc.buffer); i++ {
		r := sc.buffer[i]

		// Ignore Extend|Format
		if is(_mergedExtendFormat, r) {
			continue
		}

		// See if any of the range tables apply
		for _, rt := range rts {
			if is(rt, r) {
				return true
			}
		}

		// If we get this far, it's not there
		break
	}

	return false
}

// seekPreviousIndex works backward until it hits a rune satisfying one of the range tables,
// ignoring Extend|Format, and returns the index of the rune in the buffer
// See: https://unicode.org/reports/tr29/#Grapheme_Cluster_and_Format_Rules (driven by SB5)
func (sc *Scanner) seekPreviousIndex(pos int, rts ...*unicode.RangeTable) int {
	// Start at the end of the buffer and move backwards
	for i := pos - 1; i >= 0; i-- {
		r := sc.buffer[i]

		// Ignore Extend|Format
		if is(_mergedExtendFormat, r) {
			continue
		}

		// See if any of the range tables apply
		for _, rt := range rts {
			if is(rt, r) {
				return i
			}
		}

		// If we get this far, it's not there
		break
	}

	return -1
}

// seekPreviousIndex works backward ahead until it hits a rune satisfying one of the range tables,
// ignoring Extend|Format, reporting success
// Logic is here: https://unicode.org/reports/tr29/#Grapheme_Cluster_and_Format_Rules (driven by SB5)
func (sc *Scanner) seekPrevious(pos int, rts ...*unicode.RangeTable) bool {
	return sc.seekPreviousIndex(pos, rts...) >= 0
}

func (sc *Scanner) token() bool {
	sc.bytes = sc.bb.Bytes()
	return len(sc.bytes) > 0
}

// accept forwards the buffer cursor (pos) by 1
func (sc *Scanner) accept() {
	sc.bb.WriteRune(sc.buffer[sc.pos])
	sc.pos++
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
