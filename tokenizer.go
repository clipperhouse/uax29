package uax29

import (
	"bufio"
	"io"

	"github.com/clipperhouse/uax29/is"
)

// NewTokenizer tokenizes a reader into a stream of tokens. Iterate through the stream by calling Scan() or Next().
//
// Its uses several specs from Unicode Text Segmentation https://unicode.org/reports/tr29/. It's not a full implementation, but a decent approximation for many mainstream cases.
//
// Tokenize returns all tokens (including white space), so text can be reconstructed with fidelity.
func NewTokenizer(r io.Reader) *Tokenizer {
	return &Tokenizer{
		incoming: bufio.NewReaderSize(r, 64*1024),
	}
}

type Tokenizer struct {
	incoming *bufio.Reader
	buffer   []rune

	text string
	err  error
}

// Scan returns the Scan token. Call until it returns nil.
func (t *Tokenizer) Scan() bool {
	for {
		current, eof, err := t.readRune()
		switch {
		case err != nil:
			t.err = err
			return false
		case eof:
			// This is WB2
			// https://unicode.org/reports/tr29/#WB2

			t.text = t.token()
			return t.text != ""
		}

		// Some funcs below require lookahead; better to do I/O here than there
		// (we don't care about eof for lookahead, irrelevant)
		lookahead, _, err := t.peekRune()
		if err != nil {
			t.err = err
			return false
		}

		switch {
		case
			t.wb3(current):
			// true indicates continue
			t.accept(current)
			continue
		case
			t.wb3a(current),
			t.wb3b(current):
			// true indicates break
			goto breaking
		case
			t.wb3d(current),
			t.wb5(current),
			t.wb6(current, lookahead),
			t.wb7(current),
			t.wb7a(current),
			t.wb7b(current, lookahead),
			t.wb7c(current),
			t.wb8(current),
			t.wb9(current),
			t.wb10(current),
			t.wb11(current),
			t.wb12(current, lookahead),
			t.wb13(current):
			// true indicates continue
			t.accept(current)
			continue
		}

	breaking:
		// If we fall through all the above rules, it's a word break
		// https://unicode.org/reports/tr29/#WB999

		if len(t.buffer) > 0 {
			t.text = t.token()
			t.accept(current)
			return true
		}

		t.accept(current)
		continue
	}
}

func (t *Tokenizer) Text() string {
	return t.text
}

func (t *Tokenizer) Err() error {
	return t.err
}

// Word boundary rules: https://unicode.org/reports/tr29/#Word_Boundaries
// Typically they take the form of Category1 × Category2; × means don't break between runes of these categories.
// The funcs below test the 'left' side first, when len(buffer) == 0, i.e. beginning of token.
// Then, they test the 'right' side, if something is already in the buffer.
// In most cases, returning true means 'keep going'.

// https://unicode.org/reports/tr29/#WB3
func (t *Tokenizer) wb3(current rune) (continues bool) {
	if len(t.buffer) == 0 {
		return is.Cr(current)
	}

	previous := t.buffer[len(t.buffer)-1]
	return is.Cr(previous) && is.Lf(current)
}

// https://unicode.org/reports/tr29/#WB3a
func (t *Tokenizer) wb3a(current rune) (breaks bool) {
	if len(t.buffer) == 0 {
		return is.Cr(current) || is.Lf(current) || is.Newline(current)
	}

	previous := t.buffer[len(t.buffer)-1]
	return is.Cr(previous) || is.Lf(previous) || is.Newline(previous)
}

// https://unicode.org/reports/tr29/#WB3b
func (t *Tokenizer) wb3b(current rune) (breaks bool) {
	return is.Cr(current) || is.Lf(current) || is.Newline(current)
}

// https://unicode.org/reports/tr29/#WB5
func (t *Tokenizer) wb3d(current rune) (continues bool) {
	if len(t.buffer) == 0 {
		return is.WSegSpace(current)
	}

	previous := t.buffer[len(t.buffer)-1]
	return is.WSegSpace(previous) && is.WSegSpace(current)
}

// https://unicode.org/reports/tr29/#WB5
func (t *Tokenizer) wb5(current rune) (continues bool) {
	if len(t.buffer) == 0 {
		return is.AHLetter(current)
	}

	previous := t.buffer[len(t.buffer)-1]
	return is.AHLetter(previous) && is.AHLetter(current)
}

// https://unicode.org/reports/tr29/#WB6
func (t *Tokenizer) wb6(current, lookahead rune) (continues bool) {
	if len(t.buffer) == 0 {
		return is.AHLetter(current)
	}

	previous := t.buffer[len(t.buffer)-1]
	return is.AHLetter(previous) && (is.MidLetter(current) || is.MidNumLetQ(current)) && is.AHLetter(lookahead)
}

// https://unicode.org/reports/tr29/#WB7
func (t *Tokenizer) wb7(current rune) (continues bool) {
	if len(t.buffer) < 2 {
		return false
	}

	previous := t.buffer[len(t.buffer)-1]
	preprevious := t.buffer[len(t.buffer)-2]

	return is.AHLetter(preprevious) && (is.MidLetter(previous) || is.MidNumLetQ(previous)) && is.AHLetter(current)
}

// https://unicode.org/reports/tr29/#WB7a
func (t *Tokenizer) wb7a(current rune) (continues bool) {
	if len(t.buffer) == 0 {
		return is.HebrewLetter(current)
	}

	previous := t.buffer[len(t.buffer)-1]
	return is.HebrewLetter(previous) && is.SingleQuote(current)
}

// https://unicode.org/reports/tr29/#WB7b
func (t *Tokenizer) wb7b(current, lookahead rune) (continues bool) {
	if len(t.buffer) == 0 {
		return is.HebrewLetter(current)
	}

	previous := t.buffer[len(t.buffer)-1]
	return is.AHLetter(previous) && is.DoubleQuote(current) && is.HebrewLetter(lookahead)
}

// https://unicode.org/reports/tr29/#WB7c
func (t *Tokenizer) wb7c(current rune) (continues bool) {
	if len(t.buffer) < 2 {
		return false
	}

	previous := t.buffer[len(t.buffer)-1]
	preprevious := t.buffer[len(t.buffer)-2]

	return is.HebrewLetter(preprevious) && is.DoubleQuote(previous) && is.HebrewLetter(current)
}

// https://unicode.org/reports/tr29/#WB8
func (t *Tokenizer) wb8(current rune) (continues bool) {
	// If it's a new token and Numeric
	if len(t.buffer) == 0 {
		return is.Numeric(current)
	}

	previous := t.buffer[len(t.buffer)-1]
	return is.Numeric(previous) && is.Numeric(current)
}

// https://unicode.org/reports/tr29/#WB9
func (t *Tokenizer) wb9(current rune) (continues bool) {
	if len(t.buffer) == 0 {
		return is.AHLetter(current)
	}

	previous := t.buffer[len(t.buffer)-1]
	return is.AHLetter(previous) && is.Numeric(current)
}

// https://unicode.org/reports/tr29/#WB9
func (t *Tokenizer) wb10(current rune) (continues bool) {
	if len(t.buffer) == 0 {
		return is.Numeric(current)
	}

	previous := t.buffer[len(t.buffer)-1]
	return is.Numeric(previous) && is.AHLetter(current)
}

// https://unicode.org/reports/tr29/#WB11
func (t *Tokenizer) wb11(current rune) (continues bool) {
	if len(t.buffer) < 2 {
		return false
	}

	previous := t.buffer[len(t.buffer)-1]
	preprevious := t.buffer[len(t.buffer)-2]

	return is.Numeric(preprevious) && (is.MidNum(previous) || is.MidNumLetQ(previous)) && is.Numeric(current)
}

// https://unicode.org/reports/tr29/#WB12
func (t *Tokenizer) wb12(current, lookahead rune) (continues bool) {
	if len(t.buffer) == 0 {
		return is.Numeric(current)
	}

	previous := t.buffer[len(t.buffer)-1]
	return is.Numeric(previous) && (is.MidNum(current) || is.MidNumLetQ(current)) && is.Numeric(lookahead)
}

// https://unicode.org/reports/tr29/#WB13
func (t *Tokenizer) wb13(current rune) (continues bool) {
	// If it's a new token and Katakana
	if len(t.buffer) == 0 {
		return is.Katakana(current)
	}

	previous := t.buffer[len(t.buffer)-1]
	return is.Katakana(previous) && is.Katakana(current)
}

// https://unicode.org/reports/tr29/#WB13a
func (t *Tokenizer) wb13a(current rune) (continues bool) {
	if len(t.buffer) == 0 {
		return is.AHLetter(current) || is.Numeric(current) || is.Katakana(current) || is.ExtendNumLet(current)
	}

	previous := t.buffer[len(t.buffer)-1]
	return (is.AHLetter(previous) || is.Numeric(previous) || is.Katakana(previous) || is.ExtendNumLet(previous)) && is.ExtendNumLet(current)
}

// https://unicode.org/reports/tr29/#WB13b
func (t *Tokenizer) wb13b(current rune) (continues bool) {
	if len(t.buffer) == 0 {
		return is.ExtendNumLet(current)
	}

	previous := t.buffer[len(t.buffer)-1]
	return is.ExtendNumLet(previous) && (is.AHLetter(current) || is.Numeric(current) || is.Katakana(current))
}

func (t *Tokenizer) token() string {
	if len(t.buffer) == 0 {
		return ""
	}

	s := string(t.buffer)
	t.buffer = t.buffer[:0]

	return s
}

func (t *Tokenizer) accept(r rune) {
	t.buffer = append(t.buffer, r)
}

// readRune gets the next rune, advancing the reader
func (t *Tokenizer) readRune() (r rune, eof bool, err error) {
	r, _, err = t.incoming.ReadRune()

	if err != nil {
		if err == io.EOF {
			return r, true, nil
		}
		return r, false, err
	}

	return r, false, nil
}

func (t *Tokenizer) unreadRune() error {
	err := t.incoming.UnreadRune()

	if err != nil {
		return err
	}

	return nil
}

// peekRune peeks the next rune, without advancing the reader
func (t *Tokenizer) peekRune() (r rune, eof bool, err error) {
	r, _, err = t.incoming.ReadRune()

	if err != nil {
		if err == io.EOF {
			return r, true, nil
		}
		return r, false, err
	}

	// Unread ASAP!
	err = t.incoming.UnreadRune()
	if err != nil {
		return r, false, err
	}

	return r, false, nil
}
