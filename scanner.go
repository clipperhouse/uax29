// Package uax29 provides a scanner for Unicode text segmentation word boundaries: https://unicode.org/reports/tr29/#Word_Boundaries
// It does not implement the eintire specification, but many of the most important rules.
package uax29

import (
	"bufio"
	"io"
	"unicode"

	"github.com/clipperhouse/uax29/wordbreak"
)

// NewScanner tokenizes a reader into a stream of tokens. Iterate through the stream by calling Scan() or Next().
//
// Its uses several specs from Unicode Text Segmentation word boundaries https://unicode.org/reports/tr29/#Word_Boundaries. It's not a full implementation, but a decent approximation for many mainstream cases.
//
// It returns all tokens (including white space), so text can be reconstructed with fidelity.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		incoming: bufio.NewReaderSize(r, 64*1024),
	}
}

// Scanner is the structure for scanning an input Reader. Use NewScanner to instantiate.
type Scanner struct {
	incoming *bufio.Reader
	buffer   []rune

	text string
	err  error
}

// Scan advances to the next token, returning true if successful. Returns false on error or EOF.
func (sc *Scanner) Scan() bool {
	for {
		current, eof, err := sc.readRune()
		switch {
		case err != nil:
			sc.err = err
			return false
		case eof:
			// This is WB2
			// https://unicode.org/reports/tr29/#WB2
			sc.text = sc.token()
			sc.err = nil
			return sc.text != ""
		}

		// Some funcs below require lookahead; better to do I/O here than there
		// (we don't care about eof for lookahead, irrelevant)
		lookahead, _, err := sc.peekRune()
		if err != nil {
			sc.err = err
			return false
		}

		switch {
		case
			sc.wb3(current):
			// true indicates continue
			sc.accept(current)
			continue
		case
			sc.wb3a(current),
			sc.wb3b(current):
			// true indicates break
			goto breaking
		case
			sc.wb3d(current),
			sc.wb5(current),
			sc.wb6(current, lookahead),
			sc.wb7(current),
			sc.wb7a(current),
			sc.wb7b(current, lookahead),
			sc.wb7c(current),
			sc.wb8(current),
			sc.wb9(current),
			sc.wb10(current),
			sc.wb11(current),
			sc.wb12(current, lookahead),
			sc.wb13(current):
			// true indicates continue
			sc.accept(current)
			continue
		}

	breaking:
		// If we fall through all the above rules, it's a word break
		// https://unicode.org/reports/tr29/#WB999

		if len(sc.buffer) > 0 {
			sc.text = sc.token()
			sc.err = nil
			sc.accept(current)
			return true
		}

		sc.accept(current)
		continue
	}
}

func (sc *Scanner) Text() string {
	return sc.text
}

func (sc *Scanner) Err() error {
	return sc.err
}

// Word boundary rules: https://unicode.org/reports/tr29/#Word_Boundaries
// Typically they take the form of Category1 × Category2; × means don't break between runes of these categories.
// The funcs below test the 'left' side first, when len(buffer) == 0, i.e. beginning of token.
// Then, they test the 'right' side, if something is already in the buffer.
// In most cases, returning true means 'keep going'.

var is = unicode.Is

// https://unicode.org/reports/tr29/#WB3
func (sc *Scanner) wb3(current rune) (continues bool) {
	if len(sc.buffer) == 0 {
		return is(wordbreak.CR, current)
	}

	previous := sc.buffer[len(sc.buffer)-1]
	return is(wordbreak.CR, previous) && is(wordbreak.LF, current)
}

// https://unicode.org/reports/tr29/#WB3a
func (sc *Scanner) wb3a(current rune) (breaks bool) {
	if len(sc.buffer) == 0 {
		return is(wordbreak.CR, current) || is(wordbreak.LF, current) || is(wordbreak.Newline, current)
	}

	previous := sc.buffer[len(sc.buffer)-1]
	return is(wordbreak.CR, previous) || is(wordbreak.LF, previous) || is(wordbreak.Newline, previous)
}

// https://unicode.org/reports/tr29/#WB3b
func (t *Scanner) wb3b(current rune) (breaks bool) {
	return is(wordbreak.CR, current) || is(wordbreak.LF, current) || is(wordbreak.Newline, current)
}

// https://unicode.org/reports/tr29/#WB5
func (sc *Scanner) wb3d(current rune) (continues bool) {
	if len(sc.buffer) == 0 {
		return is(wordbreak.WSegSpace, current)
	}

	previous := sc.buffer[len(sc.buffer)-1]
	return is(wordbreak.WSegSpace, previous) && is(wordbreak.WSegSpace, current)
}

// https://unicode.org/reports/tr29/#WB5
func (sc *Scanner) wb5(current rune) (continues bool) {
	if len(sc.buffer) == 0 {
		return is(wordbreak.AHLetter, current)
	}

	previous := sc.buffer[len(sc.buffer)-1]
	return is(wordbreak.AHLetter, previous) && is(wordbreak.AHLetter, current)
}

// https://unicode.org/reports/tr29/#WB6
func (sc *Scanner) wb6(current, lookahead rune) (continues bool) {
	if len(sc.buffer) == 0 {
		return is(wordbreak.AHLetter, current)
	}

	previous := sc.buffer[len(sc.buffer)-1]
	return is(wordbreak.AHLetter, previous) && (is(wordbreak.MidLetter, current) || is(wordbreak.MidNumLetQ, current)) && is(wordbreak.AHLetter, lookahead)
}

// https://unicode.org/reports/tr29/#WB7
func (sc *Scanner) wb7(current rune) (continues bool) {
	if len(sc.buffer) < 2 {
		return false
	}

	previous := sc.buffer[len(sc.buffer)-1]
	preprevious := sc.buffer[len(sc.buffer)-2]

	return is(wordbreak.AHLetter, preprevious) && (is(wordbreak.MidLetter, previous) || is(wordbreak.MidNumLetQ, previous)) && is(wordbreak.AHLetter, current)
}

// https://unicode.org/reports/tr29/#WB7a
func (sc *Scanner) wb7a(current rune) (continues bool) {
	if len(sc.buffer) == 0 {
		return is(wordbreak.Hebrew_Letter, current)
	}

	previous := sc.buffer[len(sc.buffer)-1]
	return is(wordbreak.Hebrew_Letter, previous) && is(wordbreak.Single_Quote, current)
}

// https://unicode.org/reports/tr29/#WB7b
func (sc *Scanner) wb7b(current, lookahead rune) (continues bool) {
	if len(sc.buffer) == 0 {
		return is(wordbreak.Hebrew_Letter, current)
	}

	previous := sc.buffer[len(sc.buffer)-1]
	return is(wordbreak.AHLetter, previous) && is(wordbreak.Double_Quote, current) && is(wordbreak.Hebrew_Letter, lookahead)
}

// https://unicode.org/reports/tr29/#WB7c
func (sc *Scanner) wb7c(current rune) (continues bool) {
	if len(sc.buffer) < 2 {
		return false
	}

	previous := sc.buffer[len(sc.buffer)-1]
	preprevious := sc.buffer[len(sc.buffer)-2]

	return is(wordbreak.Hebrew_Letter, preprevious) && is(wordbreak.Double_Quote, previous) && is(wordbreak.Hebrew_Letter, current)
}

// https://unicode.org/reports/tr29/#WB8
func (sc *Scanner) wb8(current rune) (continues bool) {
	// If it's a new token and Numeric
	if len(sc.buffer) == 0 {
		return is(wordbreak.Numeric, current)
	}

	previous := sc.buffer[len(sc.buffer)-1]
	return is(wordbreak.Numeric, previous) && is(wordbreak.Numeric, current)
}

// https://unicode.org/reports/tr29/#WB9
func (sc *Scanner) wb9(current rune) (continues bool) {
	if len(sc.buffer) == 0 {
		return is(wordbreak.AHLetter, current)
	}

	previous := sc.buffer[len(sc.buffer)-1]
	return is(wordbreak.AHLetter, previous) && is(wordbreak.Numeric, current)
}

// https://unicode.org/reports/tr29/#WB9
func (sc *Scanner) wb10(current rune) (continues bool) {
	if len(sc.buffer) == 0 {
		return is(wordbreak.Numeric, current)
	}

	previous := sc.buffer[len(sc.buffer)-1]
	return is(wordbreak.Numeric, previous) && is(wordbreak.AHLetter, current)
}

// https://unicode.org/reports/tr29/#WB11
func (sc *Scanner) wb11(current rune) (continues bool) {
	if len(sc.buffer) < 2 {
		return false
	}

	previous := sc.buffer[len(sc.buffer)-1]
	preprevious := sc.buffer[len(sc.buffer)-2]

	return is(wordbreak.Numeric, preprevious) && (is(wordbreak.MidNum, previous) || is(wordbreak.MidNumLetQ, previous)) && is(wordbreak.Numeric, current)
}

// https://unicode.org/reports/tr29/#WB12
func (sc *Scanner) wb12(current, lookahead rune) (continues bool) {
	if len(sc.buffer) == 0 {
		return is(wordbreak.Numeric, current)
	}

	previous := sc.buffer[len(sc.buffer)-1]
	return is(wordbreak.Numeric, previous) && (is(wordbreak.MidNum, current) || is(wordbreak.MidNumLetQ, current)) && is(wordbreak.Numeric, lookahead)
}

// https://unicode.org/reports/tr29/#WB13
func (sc *Scanner) wb13(current rune) (continues bool) {
	// If it's a new token and Katakana
	if len(sc.buffer) == 0 {
		return is(wordbreak.Katakana, current)
	}

	previous := sc.buffer[len(sc.buffer)-1]
	return is(wordbreak.Katakana, previous) && is(wordbreak.Katakana, current)
}

// https://unicode.org/reports/tr29/#WB13a
func (sc *Scanner) wb13a(current rune) (continues bool) {
	if len(sc.buffer) == 0 {
		return is(wordbreak.AHLetter, current) || is(wordbreak.Numeric, current) || is(wordbreak.Katakana, current) || is(wordbreak.ExtendNumLet, current)
	}

	previous := sc.buffer[len(sc.buffer)-1]
	return (is(wordbreak.AHLetter, previous) || is(wordbreak.Numeric, previous) || is(wordbreak.Katakana, previous) || is(wordbreak.ExtendNumLet, previous)) && is(wordbreak.ExtendNumLet, current)
}

// https://unicode.org/reports/tr29/#WB13b
func (sc *Scanner) wb13b(current rune) (continues bool) {
	if len(sc.buffer) == 0 {
		return is(wordbreak.ExtendNumLet, current)
	}

	previous := sc.buffer[len(sc.buffer)-1]
	return is(wordbreak.ExtendNumLet, previous) && (is(wordbreak.AHLetter, current) || is(wordbreak.Numeric, current) || is(wordbreak.Katakana, current))
}

func (sc *Scanner) token() string {
	if len(sc.buffer) == 0 {
		return ""
	}

	s := string(sc.buffer)
	sc.buffer = sc.buffer[:0]

	return s
}

func (sc *Scanner) accept(r rune) {
	sc.buffer = append(sc.buffer, r)
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

func (sc *Scanner) unreadRune() error {
	err := sc.incoming.UnreadRune()

	if err != nil {
		return err
	}

	return nil
}

// peekRune peeks the next rune, without advancing the reader
func (sc *Scanner) peekRune() (r rune, eof bool, err error) {
	r, _, err = sc.incoming.ReadRune()

	if err != nil {
		if err == io.EOF {
			return r, true, nil
		}
		return r, false, err
	}

	// Unread ASAP!
	err = sc.incoming.UnreadRune()
	if err != nil {
		return r, false, err
	}

	return r, false, nil
}
