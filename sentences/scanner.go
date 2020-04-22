// Package sentences provides a scanner for Unicode text segmentation word boundaries: https://unicode.org/reports/tr29/#Word_Boundaries
package sentences

import (
	"bufio"
	"bytes"
	"io"
	"unicode"

	"golang.org/x/text/unicode/rangetable"
)

// NewScanner tokenizes a reader into a stream of tokens according to Unicode Text Segmentation word boundaries https://unicode.org/reports/tr29/#Word_Boundaries
// Iterate through the stream by calling Scan() until false.
//	text := "This is an example. And another!"
//	reader := strings.NewReader(text)
//
//	scanner := sentences.NewScanner(reader)
//	for scanner.Scan() {
//		fmt.Printf("%q\n", scanner.Text())
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

// Scan advances to the next token, returning true if successful. Returns false on error or EOF.
func (sc *Scanner) Scan() bool {
	for {
		// Fill the buffer with enough runes for lookahead
		for len(sc.buffer) < sc.pos+4 {
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

		switch {
		case sc.sb1():
			// true indicates continue
			sc.accept()
			continue
		case sc.sb2():
			// true indicates break
			break
		case
			sc.sb3():
			// true indicates continue
			sc.accept()
			continue
		case
			sc.sb4():
			// true indicates break
			break
		case
			sc.sb5():
			// true indicates continue
			sc.accept()
			continue
		case
			sc.sb6():
			// true indicates continue
			sc.accept()
			continue
		case
			sc.sb7():
			// true indicates continue
			sc.accept()
			continue
		case
			sc.sb8():
			// true indicates continue
			sc.accept()
			continue
		case
			sc.sb8a():
			// true indicates continue
			sc.accept()
			continue
		case
			sc.sb9():
			// true indicates continue
			sc.accept()
			continue
		case
			sc.sb10():
			// true indicates continue
			sc.accept()
			continue
		case
			sc.sb11():
			// true indicates accept & break
			break
		case
			sc.sb998():
			sc.accept()
			continue
			// true indicates continue
		}

		// If we fall through all the above rules, it's a word break, aka WB999
		break
	}

	return sc.emit()
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

// Word boundary rules: https://unicode.org/reports/tr29/#Word_Boundaries
// In most cases, returning true means 'keep going'; check the name of the return var for clarity

var is = unicode.Is

// wb1 implements https://unicode.org/reports/tr29/#SB1
func (sc *Scanner) sb1() (accept bool) {
	sot := sc.pos == 0 // "start of text"
	eof := len(sc.buffer) == sc.pos
	return sot && !eof
}

// wb2 implements https://unicode.org/reports/tr29/#SB2
func (sc *Scanner) sb2() (breaking bool) {
	// eof
	return len(sc.buffer) == sc.pos
}

// wb3 implements https://unicode.org/reports/tr29/#SB3
func (sc *Scanner) sb3() (accept bool) {
	current := sc.buffer[sc.pos]
	if !is(LF, current) {
		return false
	}

	previous := sc.buffer[sc.pos-1]
	return is(CR, previous)
}

var _mergedParaSep = rangetable.Merge(Sep, CR, LF)

// wb3a implements https://unicode.org/reports/tr29/#SB3a
func (sc *Scanner) sb4() (breaking bool) {
	previous := sc.buffer[sc.pos-1]
	return is(_mergedParaSep, previous)
}

var _mergedExtendFormat = rangetable.Merge(Extend, Format)

// wb4 implements https://unicode.org/reports/tr29/#SB4
func (sc *Scanner) sb5() (accept bool) {
	current := sc.buffer[sc.pos]
	return is(_mergedExtendFormat, current)
}

// seekForward looks ahead until it hits a rune satisfying one of the range tables,
// ignoring Extend|Format|ZWJ
// See: https://unicode.org/reports/tr29/#Grapheme_Cluster_and_Format_Rules (driven by WB4)
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
// ignoring Extend|Format|ZWJ, and returns the index of the rune in the buffer
// See: https://unicode.org/reports/tr29/#Grapheme_Cluster_and_Format_Rules (driven by WB4)
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
// ignoring ExtendFormatZWJ, reporting success
// Logic is here: https://unicode.org/reports/tr29/#Grapheme_Cluster_and_Format_Rules (driven by WB4)
func (sc *Scanner) seekPrevious(pos int, rts ...*unicode.RangeTable) bool {
	return sc.seekPreviousIndex(pos, rts...) >= 0
}

// wb5 implements https://unicode.org/reports/tr29/#SB5
func (sc *Scanner) sb6() (accept bool) {
	current := sc.buffer[sc.pos]
	if !is(Numeric, current) {
		return false
	}

	return sc.seekPrevious(sc.pos, ATerm)
}

var _mergedUpperLower = rangetable.Merge(Upper, Lower)

// wb6 implements https://unicode.org/reports/tr29/#SB6
func (sc *Scanner) sb7() (accept bool) {
	current := sc.buffer[sc.pos]
	if !(is(Upper, current)) {
		return false
	}

	previous := sc.seekPreviousIndex(sc.pos, ATerm)
	if !(previous > 0) {
		return false
	}

	return sc.seekPrevious(sc.pos, _mergedUpperLower)
}

var _mergedOLetterUpperLowerParaSepSATerm = rangetable.Merge(OLetter, Upper, Lower, _mergedParaSep, _mergedSATerm)
var _mergedSATerm = rangetable.Merge(STerm, ATerm)

// wb6 implements https://unicode.org/reports/tr29/#SB6
func (sc *Scanner) sb8() (accept bool) {
	if !sc.seekForward(sc.pos-1, Lower) {
		return false
	}

	current := sc.buffer[sc.pos]
	if !is(_mergedOLetterUpperLowerParaSepSATerm, current) {
		return true
	}

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

	return sc.seekPrevious(pos, ATerm)
}

var _mergedSContinueSATerm = rangetable.Merge(SContinue, _mergedSATerm)

// wb6 implements https://unicode.org/reports/tr29/#SB6
func (sc *Scanner) sb8a() (accept bool) {
	current := sc.buffer[sc.pos]
	if !is(_mergedSContinueSATerm, current) {
		return false
	}

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

	return sc.seekPrevious(pos, _mergedSATerm)
}

var _mergedCloseSpParaSep = rangetable.Merge(Close, Sp, _mergedParaSep)

// wb6 implements https://unicode.org/reports/tr29/#SB6
func (sc *Scanner) sb9() (accept bool) {
	current := sc.buffer[sc.pos]
	if !is(_mergedCloseSpParaSep, current) {
		return false
	}

	pos := sc.pos

	close := pos
	for {
		close = sc.seekPreviousIndex(close, Close)
		if close < 0 {
			break
		}
		pos = close
	}

	return sc.seekPrevious(pos, _mergedSATerm)
}

var _mergedSpParaSep = rangetable.Merge(Sp, _mergedParaSep)

// wb6 implements https://unicode.org/reports/tr29/#SB6
func (sc *Scanner) sb10() (accept bool) {
	current := sc.buffer[sc.pos]
	if !is(_mergedSpParaSep, current) {
		return false
	}

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

	return sc.seekPrevious(pos, _mergedSATerm)
}

// wb6 implements https://unicode.org/reports/tr29/#SB6
func (sc *Scanner) sb11() (breaking bool) {
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

	return sc.seekPrevious(pos, _mergedSATerm)
}

// sb998 implements https://unicode.org/reports/tr29/#SB998
func (sc *Scanner) sb998() bool {
	return sc.pos > 0
}

func (sc *Scanner) emit() bool {
	// Get the bytes & reset
	sc.bytes = sc.bb.Bytes()
	sc.bb.Reset()

	// Drop the emitted runes (optimization to avoid growing array)
	copy(sc.buffer, sc.buffer[sc.pos:])
	sc.buffer = sc.buffer[:len(sc.buffer)-sc.pos]

	sc.pos = 0

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
