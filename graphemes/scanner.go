// Package graphemes provides a scanner for Unicode text segmentation grapheme cluster boundaries: https://unicode.org/reports/tr29/#Grapheme_Cluster_Boundaries
package graphemes

import (
	"bufio"
	"bytes"
	"io"
	"unicode"

	"github.com/clipperhouse/uax29/emoji"
)

// NewScanner tokenizes a reader into a stream of grapheme clusters according to Unicode Text Segmentation boundaries https://unicode.org/reports/tr29/#Grapheme_Cluster_Boundaries
// Iterate through the stream by calling Scan() until false.
//	text := "Good dog! 👍🏼🐶"
//	reader := strings.NewReader(text)
//
//	scanner := graphemes.NewScanner(reader)
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

// Scan advances to the next grapheme cluster, returning true if successful. Returns false on error or EOF.
func (sc *Scanner) Scan() bool {
	for {
		// Fill the buffer with enough runes for lookahead
		for len(sc.buffer) < sc.pos+6 {
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
		case sc.gb1():
			// true indicates continue
			sc.accept()
			continue
		case sc.gb2():
			// true indicates break
			break
		case
			sc.gb3():
			// true indicates continue
			sc.accept()
			continue
		case
			sc.gb4(),
			sc.gb5():
			// true indicates break
			break
		case
			sc.gb6(),
			sc.gb7(),
			sc.gb8(),
			sc.gb9(),
			sc.gb9a(),
			sc.gb9b(),
			sc.gb11(),
			sc.gb12(),
			sc.gb13():
			// true indicates continue
			sc.accept()
			continue
		}

		// If we fall through all the above rules, it's a break
		break
	}

	return sc.gb999()
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

// Grapheme cluster rules: https://unicode.org/reports/tr29/#Grapheme_Cluster_Boundaries
// In most cases, returning true means 'keep going'; check the name of the return var for clarity

var is = unicode.Is

// gb1 implements https://unicode.org/reports/tr29/#GB1
func (sc *Scanner) gb1() (accept bool) {
	sot := sc.pos == 0 // "start of text"
	eof := len(sc.buffer) == sc.pos
	return sot && !eof
}

// gb2 implements https://unicode.org/reports/tr29/#GB2
func (sc *Scanner) gb2() (breaking bool) {
	// eof
	return len(sc.buffer) == sc.pos
}

// gb3 implements https://unicode.org/reports/tr29/#GB3
func (sc *Scanner) gb3() (accept bool) {
	current := sc.buffer[sc.pos]
	if !is(LF, current) {
		return false
	}

	previous := sc.buffer[sc.pos-1]
	return is(CR, previous)
}

// gb4 implements https://unicode.org/reports/tr29/#GB4
func (sc *Scanner) gb4() (breaking bool) {
	previous := sc.buffer[sc.pos-1]
	return is(_mergedControlCRLF, previous)
}

// gb5 implements https://unicode.org/reports/tr29/#GB5
func (sc *Scanner) gb5() (breaking bool) {
	current := sc.buffer[sc.pos]
	return is(_mergedControlCRLF, current)
}

// gb6 implements https://unicode.org/reports/tr29/#GB6
func (sc *Scanner) gb6() (accept bool) {
	if sc.pos < 1 {
		return false
	}

	current := sc.buffer[sc.pos]
	if !is(_mergedLVLVLVT, current) {
		return false
	}

	previous := sc.buffer[sc.pos-1]
	return is(L, previous)
}

// gb7 implements https://unicode.org/reports/tr29/#GB7
func (sc *Scanner) gb7() (accept bool) {
	if sc.pos < 1 {
		return false
	}

	current := sc.buffer[sc.pos]
	if !is(_mergedVT, current) {
		return false
	}

	previous := sc.buffer[sc.pos-1]
	return is(_mergedLVV, previous)
}

// gb8 implements https://unicode.org/reports/tr29/#GB8
func (sc *Scanner) gb8() (accept bool) {
	if sc.pos < 1 {
		return false
	}

	current := sc.buffer[sc.pos]
	if !is(T, current) {
		return false
	}

	previous := sc.buffer[sc.pos-1]
	return is(_mergedLVTT, previous)
}

// gb9 implements https://unicode.org/reports/tr29/#GB9
func (sc *Scanner) gb9() (accept bool) {
	current := sc.buffer[sc.pos]
	return is(_mergedExtendZWJ, current)
}

// gb9 implements https://unicode.org/reports/tr29/#GB9a
func (sc *Scanner) gb9a() (accept bool) {
	current := sc.buffer[sc.pos]
	return is(SpacingMark, current)
}

// gb9 implements https://unicode.org/reports/tr29/#GB9b
func (sc *Scanner) gb9b() (accept bool) {
	if sc.pos < 1 {
		return false
	}

	previous := sc.buffer[sc.pos-1]
	return is(Prepend, previous)
}

// seekPreviousIndex works backward until it hits a rune satisfying one of the range tables,
// ignoring Extend, and returns the index of the rune in the buffer
// See: https://unicode.org/reports/tr29/#Grapheme_Cluster_and_Format_Rules (driven by SB5)
func (sc *Scanner) seekPreviousIndex(pos int, rts ...*unicode.RangeTable) int {
	// Start at the end of the buffer and move backwards
	for i := pos - 1; i >= 0; i-- {
		r := sc.buffer[i]

		// Ignore Extend
		if is(Extend, r) {
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

// gb11 implements https://unicode.org/reports/tr29/#GB11
func (sc *Scanner) gb11() (accept bool) {
	if sc.pos < 2 {
		return false
	}

	current := sc.buffer[sc.pos]
	if !is(emoji.Extended_Pictographic, current) {
		return false
	}

	previous := sc.buffer[sc.pos-1]
	if !is(ZWJ, previous) {
		return false
	}

	return sc.seekPrevious(sc.pos-1, emoji.Extended_Pictographic)
}

// gb12 implements https://unicode.org/reports/tr29/#GB12
func (sc *Scanner) gb12() (accept bool) {
	current := sc.buffer[sc.pos]
	if !is(Regional_Indicator, current) {
		return false
	}

	// Buffer comprised entirely of an odd number of RI
	count := 0
	for i := sc.pos - 1; i >= 0; i-- {
		r := sc.buffer[i]
		if !is(Regional_Indicator, r) {
			return false
		}
		count++
	}

	// If we fall through, we've seen the whole buffer,
	// so it's all Regional_Indicator
	odd := count > 0 && count%2 == 1
	return odd
}

// gb13 implements https://unicode.org/reports/tr29/#GB13
func (sc *Scanner) gb13() (accept bool) {
	current := sc.buffer[sc.pos]
	if !is(Regional_Indicator, current) {
		return false
	}

	// Last n runes represent an odd number of RI
	count := 0
	for i := sc.pos - 1; i >= 0; i-- {
		r := sc.buffer[i]
		if !is(Regional_Indicator, r) {
			odd := count > 0 && count%2 == 1
			return odd
		}
		count++
	}

	return false
}

// gb999 implements https://unicode.org/reports/tr29/#GB999
// i.e. break
func (sc *Scanner) gb999() bool {
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
