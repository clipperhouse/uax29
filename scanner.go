// Package uax29 provides a scanner for Unicode text segmentation word boundaries: https://unicode.org/reports/tr29/#Word_Boundaries
// It does not implement the eintire specification, but many of the most important rules.
package uax29

import (
	"bufio"
	"io"

	"github.com/clipperhouse/uax29/is"
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

// https://unicode.org/reports/tr29/#WB3
func (sc *Scanner) wb3(current rune) (continues bool) {
	if len(sc.buffer) == 0 {
		return is.Cr(current)
	}

	previous := sc.buffer[len(sc.buffer)-1]
	return is.Cr(previous) && is.Lf(current)
}

// https://unicode.org/reports/tr29/#WB3a
func (sc *Scanner) wb3a(current rune) (breaks bool) {
	if len(sc.buffer) == 0 {
		return is.Cr(current) || is.Lf(current) || is.Newline(current)
	}

	previous := sc.buffer[len(sc.buffer)-1]
	return is.Cr(previous) || is.Lf(previous) || is.Newline(previous)
}

// https://unicode.org/reports/tr29/#WB3b
func (t *Scanner) wb3b(current rune) (breaks bool) {
	return is.Cr(current) || is.Lf(current) || is.Newline(current)
}

// https://unicode.org/reports/tr29/#WB5
func (sc *Scanner) wb3d(current rune) (continues bool) {
	if len(sc.buffer) == 0 {
		return is.WSegSpace(current)
	}

	previous := sc.buffer[len(sc.buffer)-1]
	return is.WSegSpace(previous) && is.WSegSpace(current)
}

// https://unicode.org/reports/tr29/#WB5
func (sc *Scanner) wb5(current rune) (continues bool) {
	if len(sc.buffer) == 0 {
		return is.AHLetter(current)
	}

	previous := sc.buffer[len(sc.buffer)-1]
	return is.AHLetter(previous) && is.AHLetter(current)
}

// https://unicode.org/reports/tr29/#WB6
func (sc *Scanner) wb6(current, lookahead rune) (continues bool) {
	if len(sc.buffer) == 0 {
		return is.AHLetter(current)
	}

	previous := sc.buffer[len(sc.buffer)-1]
	return is.AHLetter(previous) && (is.MidLetter(current) || is.MidNumLetQ(current)) && is.AHLetter(lookahead)
}

// https://unicode.org/reports/tr29/#WB7
func (sc *Scanner) wb7(current rune) (continues bool) {
	if len(sc.buffer) < 2 {
		return false
	}

	previous := sc.buffer[len(sc.buffer)-1]
	preprevious := sc.buffer[len(sc.buffer)-2]

	return is.AHLetter(preprevious) && (is.MidLetter(previous) || is.MidNumLetQ(previous)) && is.AHLetter(current)
}

// https://unicode.org/reports/tr29/#WB7a
func (sc *Scanner) wb7a(current rune) (continues bool) {
	if len(sc.buffer) == 0 {
		return is.HebrewLetter(current)
	}

	previous := sc.buffer[len(sc.buffer)-1]
	return is.HebrewLetter(previous) && is.SingleQuote(current)
}

// https://unicode.org/reports/tr29/#WB7b
func (sc *Scanner) wb7b(current, lookahead rune) (continues bool) {
	if len(sc.buffer) == 0 {
		return is.HebrewLetter(current)
	}

	previous := sc.buffer[len(sc.buffer)-1]
	return is.AHLetter(previous) && is.DoubleQuote(current) && is.HebrewLetter(lookahead)
}

// https://unicode.org/reports/tr29/#WB7c
func (sc *Scanner) wb7c(current rune) (continues bool) {
	if len(sc.buffer) < 2 {
		return false
	}

	previous := sc.buffer[len(sc.buffer)-1]
	preprevious := sc.buffer[len(sc.buffer)-2]

	return is.HebrewLetter(preprevious) && is.DoubleQuote(previous) && is.HebrewLetter(current)
}

// https://unicode.org/reports/tr29/#WB8
func (sc *Scanner) wb8(current rune) (continues bool) {
	// If it's a new token and Numeric
	if len(sc.buffer) == 0 {
		return is.Numeric(current)
	}

	previous := sc.buffer[len(sc.buffer)-1]
	return is.Numeric(previous) && is.Numeric(current)
}

// https://unicode.org/reports/tr29/#WB9
func (sc *Scanner) wb9(current rune) (continues bool) {
	if len(sc.buffer) == 0 {
		return is.AHLetter(current)
	}

	previous := sc.buffer[len(sc.buffer)-1]
	return is.AHLetter(previous) && is.Numeric(current)
}

// https://unicode.org/reports/tr29/#WB9
func (sc *Scanner) wb10(current rune) (continues bool) {
	if len(sc.buffer) == 0 {
		return is.Numeric(current)
	}

	previous := sc.buffer[len(sc.buffer)-1]
	return is.Numeric(previous) && is.AHLetter(current)
}

// https://unicode.org/reports/tr29/#WB11
func (sc *Scanner) wb11(current rune) (continues bool) {
	if len(sc.buffer) < 2 {
		return false
	}

	previous := sc.buffer[len(sc.buffer)-1]
	preprevious := sc.buffer[len(sc.buffer)-2]

	return is.Numeric(preprevious) && (is.MidNum(previous) || is.MidNumLetQ(previous)) && is.Numeric(current)
}

// https://unicode.org/reports/tr29/#WB12
func (sc *Scanner) wb12(current, lookahead rune) (continues bool) {
	if len(sc.buffer) == 0 {
		return is.Numeric(current)
	}

	previous := sc.buffer[len(sc.buffer)-1]
	return is.Numeric(previous) && (is.MidNum(current) || is.MidNumLetQ(current)) && is.Numeric(lookahead)
}

// https://unicode.org/reports/tr29/#WB13
func (sc *Scanner) wb13(current rune) (continues bool) {
	// If it's a new token and Katakana
	if len(sc.buffer) == 0 {
		return is.Katakana(current)
	}

	previous := sc.buffer[len(sc.buffer)-1]
	return is.Katakana(previous) && is.Katakana(current)
}

// https://unicode.org/reports/tr29/#WB13a
func (sc *Scanner) wb13a(current rune) (continues bool) {
	if len(sc.buffer) == 0 {
		return is.AHLetter(current) || is.Numeric(current) || is.Katakana(current) || is.ExtendNumLet(current)
	}

	previous := sc.buffer[len(sc.buffer)-1]
	return (is.AHLetter(previous) || is.Numeric(previous) || is.Katakana(previous) || is.ExtendNumLet(previous)) && is.ExtendNumLet(current)
}

// https://unicode.org/reports/tr29/#WB13b
func (sc *Scanner) wb13b(current rune) (continues bool) {
	if len(sc.buffer) == 0 {
		return is.ExtendNumLet(current)
	}

	previous := sc.buffer[len(sc.buffer)-1]
	return is.ExtendNumLet(previous) && (is.AHLetter(current) || is.Numeric(current) || is.Katakana(current))
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
