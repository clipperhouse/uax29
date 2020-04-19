// Package words provides a scanner for Unicode text segmentation word boundaries: https://unicode.org/reports/tr29/#Word_Boundaries
// It does not implement the entire specification, but many of the most important rules.
package words

import (
	"bufio"
	"io"
	"unicode"

	"github.com/clipperhouse/uax29/emoji"
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
		if err != nil {
			sc.err = err
			return false
		}

		switch {
		case sc.wb1(eof):
			// true indicates continue
			sc.accept(current)
			continue
		case sc.wb2(eof):
			// true indicates break
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
			sc.wb4(current),
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
			sc.wb13(current),
			sc.wb13a(current),
			sc.wb13b(current),
			sc.wb15(current),
			sc.wb16(current):
			// true indicates continue
			sc.accept(current)
			continue
		}

	breaking:
		// If we fall through all the above rules, it's a word break
		// wb999 implements https://unicode.org/reports/tr29/#WB999

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

// Text returns the current token, given a successful call to Scan
func (sc *Scanner) Text() string {
	return sc.text
}

// Err returns the current error, given an unsuccessful call to Scan
func (sc *Scanner) Err() error {
	return sc.err
}

// Word boundary rules: https://unicode.org/reports/tr29/#Word_Boundaries
// In most cases, returning true means 'keep going'; check the name of the return var for clarity

var is = unicode.Is

// wb1 implements https://unicode.org/reports/tr29/#WB1
func (sc *Scanner) wb1(eof bool) (continues bool) {
	sot := len(sc.buffer) == 0 // "start of text"
	return sot && !eof
}

// wb2 implements https://unicode.org/reports/tr29/#WB2
func (sc *Scanner) wb2(eof bool) (breaks bool) {
	// A bit silly, but reads consistently in Scan above
	return eof
}

// wb3 implements https://unicode.org/reports/tr29/#WB3
func (sc *Scanner) wb3(current rune) (continues bool) {
	previous := sc.buffer[len(sc.buffer)-1]
	return is(CR, previous) && is(LF, current)
}

// wb3a implements https://unicode.org/reports/tr29/#WB3a
func (sc *Scanner) wb3a(current rune) (breaks bool) {
	previous := sc.buffer[len(sc.buffer)-1]
	return is(CR, previous) || is(LF, previous) || is(Newline, previous)
}

// wb3b implements https://unicode.org/reports/tr29/#WB3b
func (sc *Scanner) wb3b(current rune) (breaks bool) {
	return is(CR, current) || is(LF, current) || is(Newline, current)
}

// wb3c implements https://unicode.org/reports/tr29/#WB3c
func (sc *Scanner) wb3c(current rune) (continues bool) {
	previous := sc.buffer[len(sc.buffer)-1]
	return is(ZWJ, previous) && is(emoji.Extended_Pictographic, current)
}

// wb3d implements https://unicode.org/reports/tr29/#WB3d
func (sc *Scanner) wb3d(current rune) (continues bool) {
	previous := sc.buffer[len(sc.buffer)-1]
	return is(WSegSpace, previous) && is(WSegSpace, current)
}

// wb4 implements https://unicode.org/reports/tr29/#WB4
func (sc *Scanner) wb4(current rune) (continues bool) {
	return is(ExtendFormatZWJ, current)
}

// wb5 implements https://unicode.org/reports/tr29/#WB5
func (sc *Scanner) wb5(current rune) (continues bool) {
	previous := sc.buffer[len(sc.buffer)-1]
	return is(AHLetter, previous) && is(AHLetter, current)
}

// wb6 implements https://unicode.org/reports/tr29/#WB6
func (sc *Scanner) wb6(current, lookahead rune) (continues bool) {
	previous := sc.buffer[len(sc.buffer)-1]
	return is(AHLetter, previous) && (is(MidLetter, current) || is(MidNumLetQ, current)) && is(AHLetter, lookahead)
}

// wb7 implements https://unicode.org/reports/tr29/#WB7
func (sc *Scanner) wb7(current rune) (continues bool) {
	if len(sc.buffer) < 2 {
		return false
	}

	previous := sc.buffer[len(sc.buffer)-1]
	preprevious := sc.buffer[len(sc.buffer)-2]

	return is(AHLetter, preprevious) && (is(MidLetter, previous) || is(MidNumLetQ, previous)) && is(AHLetter, current)
}

// wb7a implements https://unicode.org/reports/tr29/#WB7a
func (sc *Scanner) wb7a(current rune) (continues bool) {
	previous := sc.buffer[len(sc.buffer)-1]
	return is(Hebrew_Letter, previous) && is(Single_Quote, current)
}

// wb7b implements https://unicode.org/reports/tr29/#WB7b
func (sc *Scanner) wb7b(current, lookahead rune) (continues bool) {
	previous := sc.buffer[len(sc.buffer)-1]
	return is(AHLetter, previous) && is(Double_Quote, current) && is(Hebrew_Letter, lookahead)
}

// wb7c implements https://unicode.org/reports/tr29/#WB7c
func (sc *Scanner) wb7c(current rune) (continues bool) {
	if len(sc.buffer) < 2 {
		return false
	}

	previous := sc.buffer[len(sc.buffer)-1]
	preprevious := sc.buffer[len(sc.buffer)-2]

	return is(Hebrew_Letter, preprevious) && is(Double_Quote, previous) && is(Hebrew_Letter, current)
}

// wb8 implements https://unicode.org/reports/tr29/#WB8
func (sc *Scanner) wb8(current rune) (continues bool) {
	previous := sc.buffer[len(sc.buffer)-1]
	return is(Numeric, previous) && is(Numeric, current)
}

// wb9 implements https://unicode.org/reports/tr29/#WB9
func (sc *Scanner) wb9(current rune) (continues bool) {
	previous := sc.buffer[len(sc.buffer)-1]
	return is(AHLetter, previous) && is(Numeric, current)
}

// wb10 implements https://unicode.org/reports/tr29/#WB10
func (sc *Scanner) wb10(current rune) (continues bool) {
	previous := sc.buffer[len(sc.buffer)-1]
	return is(Numeric, previous) && is(AHLetter, current)
}

// wb11 implements https://unicode.org/reports/tr29/#WB11
func (sc *Scanner) wb11(current rune) (continues bool) {
	if len(sc.buffer) < 2 {
		return false
	}

	previous := sc.buffer[len(sc.buffer)-1]
	preprevious := sc.buffer[len(sc.buffer)-2]

	return is(Numeric, preprevious) && (is(MidNum, previous) || is(MidNumLetQ, previous)) && is(Numeric, current)
}

// wb12 implements https://unicode.org/reports/tr29/#WB12
func (sc *Scanner) wb12(current, lookahead rune) (continues bool) {
	previous := sc.buffer[len(sc.buffer)-1]
	return is(Numeric, previous) && (is(MidNum, current) || is(MidNumLetQ, current)) && is(Numeric, lookahead)
}

// wb13 implements https://unicode.org/reports/tr29/#WB13
func (sc *Scanner) wb13(current rune) (continues bool) {
	previous := sc.buffer[len(sc.buffer)-1]
	return is(Katakana, previous) && is(Katakana, current)
}

// wb13a implements https://unicode.org/reports/tr29/#WB13a
func (sc *Scanner) wb13a(current rune) (continues bool) {
	previous := sc.buffer[len(sc.buffer)-1]
	return (is(AHLetter, previous) || is(Numeric, previous) || is(Katakana, previous) || is(ExtendNumLet, previous)) && is(ExtendNumLet, current)
}

// wb13b implements https://unicode.org/reports/tr29/#WB13b
func (sc *Scanner) wb13b(current rune) (continues bool) {
	previous := sc.buffer[len(sc.buffer)-1]
	return is(ExtendNumLet, previous) && (is(AHLetter, current) || is(Numeric, current) || is(Katakana, current))
}

// wb15 implements https://unicode.org/reports/tr29/#WB15
func (sc *Scanner) wb15(current rune) (continues bool) {
	// Buffer comprised entirely of an odd number of RI
	count := 0
	for i := len(sc.buffer) - 1; i >= 0; i-- {
		r := sc.buffer[i]
		if !is(Regional_Indicator, r) {
			return false
		}
		count++
	}
	odd := count > 0 && count%2 == 1
	return odd
}

// wb16 implements https://unicode.org/reports/tr29/#WB16
func (sc *Scanner) wb16(current rune) (continues bool) {
	// Last n runes represent an odd number of RI
	count := 0
	for i := len(sc.buffer) - 1; i >= 0; i-- {
		r := sc.buffer[i]
		if !is(Regional_Indicator, r) {
			break
		}
		count++
	}
	odd := count > 0 && count%2 == 1
	return odd
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
