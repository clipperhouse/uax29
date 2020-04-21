// Package words provides a scanner for Unicode text segmentation word boundaries: https://unicode.org/reports/tr29/#Word_Boundaries
package words

import (
	"bufio"
	"bytes"
	"io"
	"unicode"

	"github.com/clipperhouse/uax29/emoji"
)

// NewScanner tokenizes a reader into a stream of tokens according to Unicode Text Segmentation word boundaries https://unicode.org/reports/tr29/#Word_Boundaries
// Iterate through the stream by calling Scan() until false.
//	text := "This is an example."
//	reader := strings.NewReader(text)
//
//	scanner := words.NewScanner(reader)
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
		case sc.wb1():
			// true indicates continue
			sc.accept()
			continue
		case sc.wb2():
			// true indicates break
			goto breaking
		case
			sc.wb3():
			// true indicates continue
			sc.accept()
			continue
		case
			sc.wb3a(),
			sc.wb3b():
			// true indicates break
			goto breaking
		case
			sc.wb3c(),
			sc.wb3d(),
			sc.wb4(),
			sc.wb5(),
			sc.wb6(),
			sc.wb7(),
			sc.wb7a(),
			sc.wb7b(),
			sc.wb7c(),
			sc.wb8(),
			sc.wb9(),
			sc.wb10(),
			sc.wb11(),
			sc.wb12(),
			sc.wb13(),
			sc.wb13a(),
			sc.wb13b(),
			sc.wb15(),
			sc.wb16():
			// true indicates continue
			sc.accept()
			continue
		}

	breaking:
		// If we fall through all the above rules, it's a word break, aka WB999

		return sc.wb999()
	}
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

// wb1 implements https://unicode.org/reports/tr29/#WB1
func (sc *Scanner) wb1() (accept bool) {
	sot := sc.pos == 0 // "start of text"
	eof := len(sc.buffer) == sc.pos
	return sot && !eof
}

// wb2 implements https://unicode.org/reports/tr29/#WB2
func (sc *Scanner) wb2() (breaking bool) {
	// eof
	return len(sc.buffer) == sc.pos
}

// wb3 implements https://unicode.org/reports/tr29/#WB3
func (sc *Scanner) wb3() (accept bool) {
	current := sc.buffer[sc.pos]
	if !is(LF, current) {
		return false
	}

	previous := sc.buffer[sc.pos-1]
	return is(CR, previous)
}

// wb3a implements https://unicode.org/reports/tr29/#WB3a
func (sc *Scanner) wb3a() (breaking bool) {
	previous := sc.buffer[sc.pos-1]
	return is(CR, previous) || is(LF, previous) || is(Newline, previous)
}

// wb3b implements https://unicode.org/reports/tr29/#WB3b
func (sc *Scanner) wb3b() (breaking bool) {
	current := sc.buffer[sc.pos]
	return is(CR, current) || is(LF, current) || is(Newline, current)
}

// wb3c implements https://unicode.org/reports/tr29/#WB3c
func (sc *Scanner) wb3c() (accept bool) {
	current := sc.buffer[sc.pos]
	if !is(emoji.Extended_Pictographic, current) {
		return false
	}

	previous := sc.buffer[sc.pos-1]
	return is(ZWJ, previous)
}

// wb3d implements https://unicode.org/reports/tr29/#WB3d
func (sc *Scanner) wb3d() (accept bool) {
	current := sc.buffer[sc.pos]
	if !is(WSegSpace, current) {
		return false
	}

	previous := sc.buffer[sc.pos-1]
	return is(WSegSpace, previous)
}

// wb4 implements https://unicode.org/reports/tr29/#WB4
func (sc *Scanner) wb4() (accept bool) {
	current := sc.buffer[sc.pos]
	return is(_mergedExtendFormatZWJ, current)
}

// seekForward looks ahead until it hits a rune satisfying one of the range tables,
// ignoring ExtendFormatZWJ
// See: https://unicode.org/reports/tr29/#Grapheme_Cluster_and_Format_Rules (driven by WB4)
func (sc *Scanner) seekForward(rts ...*unicode.RangeTable) bool {
	for i := sc.pos + 1; i < len(sc.buffer); i++ {
		r := sc.buffer[i]

		// Ignore ExtendFormatZWJ
		if is(_mergedExtendFormatZWJ, r) {
			continue
		}

		// See if any of the range tables apply
		for _, rt := range rts {
			if is(rt, r) {
				return true
			}
		}

		// If we get this far, it's false
		break
	}

	return false
}

// seekPreviousIndex works backward ahead until it hits a rune satisfying one of the range tables,
// ignoring ExtendFormatZWJ, returning the index of the rune in the buffer
// See: https://unicode.org/reports/tr29/#Grapheme_Cluster_and_Format_Rules (driven by WB4)
func (sc *Scanner) seekPreviousIndex(pos int, rts ...*unicode.RangeTable) int {
	// Start at the end of the buffer and move backwards
	for i := pos - 1; i >= 0; i-- {
		r := sc.buffer[i]

		// Ignore ExtendFormatZWJ
		if is(_mergedExtendFormatZWJ, r) {
			continue
		}

		// See if any of the range tables apply
		for _, rt := range rts {
			if is(rt, r) {
				return i
			}
		}

		// If we get this far, it's false
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

// wb5 implements https://unicode.org/reports/tr29/#WB5
func (sc *Scanner) wb5() (accept bool) {
	current := sc.buffer[sc.pos]
	if !is(AHLetter, current) {
		return false
	}

	return sc.seekPrevious(sc.pos, AHLetter)
}

// wb6 implements https://unicode.org/reports/tr29/#WB6
func (sc *Scanner) wb6() (accept bool) {
	current := sc.buffer[sc.pos]
	if !(is(MidLetter, current) || is(MidNumLetQ, current)) {
		return false
	}

	if !sc.seekForward(AHLetter) {
		return false
	}

	return sc.seekPrevious(sc.pos, AHLetter)
}

// wb7 implements https://unicode.org/reports/tr29/#WB7
func (sc *Scanner) wb7() (accept bool) {
	current := sc.buffer[sc.pos]
	if !(is(AHLetter, current) || is(_mergedExtendFormatZWJ, current)) {
		return false
	}

	previous := sc.seekPreviousIndex(sc.pos, MidLetter, MidNumLetQ)
	if previous < 0 {
		return false
	}

	return sc.seekPrevious(previous, AHLetter)
}

// wb7a implements https://unicode.org/reports/tr29/#WB7a
func (sc *Scanner) wb7a() (accept bool) {
	current := sc.buffer[sc.pos]
	if !is(Single_Quote, current) {
		return false
	}

	return sc.seekPrevious(sc.pos, Hebrew_Letter)
}

// wb7b implements https://unicode.org/reports/tr29/#WB7b
func (sc *Scanner) wb7b() (accept bool) {
	current := sc.buffer[sc.pos]
	if !is(Double_Quote, current) {
		return false
	}

	if !sc.seekForward(Hebrew_Letter) {
		return false
	}

	return sc.seekPrevious(sc.pos, Hebrew_Letter)
}

// wb7c implements https://unicode.org/reports/tr29/#WB7c
func (sc *Scanner) wb7c() (accept bool) {
	current := sc.buffer[sc.pos]

	if !is(Hebrew_Letter, current) {
		return false
	}

	previous := sc.seekPreviousIndex(sc.pos, Double_Quote)
	if previous < 0 {
		return false
	}

	return sc.seekPrevious(previous, Hebrew_Letter)
}

// wb8 implements https://unicode.org/reports/tr29/#WB8
func (sc *Scanner) wb8() (accept bool) {
	current := sc.buffer[sc.pos]
	if !is(Numeric, current) {
		return false
	}

	return sc.seekPrevious(sc.pos, Numeric)
}

// wb9 implements https://unicode.org/reports/tr29/#WB9
func (sc *Scanner) wb9() (accept bool) {
	current := sc.buffer[sc.pos]
	if !is(Numeric, current) {
		return false
	}

	return sc.seekPrevious(sc.pos, AHLetter)
}

// wb10 implements https://unicode.org/reports/tr29/#WB10
func (sc *Scanner) wb10() (accept bool) {
	current := sc.buffer[sc.pos]
	if !is(AHLetter, current) {
		return false
	}

	return sc.seekPrevious(sc.pos, Numeric)
}

// wb11 implements https://unicode.org/reports/tr29/#WB11
func (sc *Scanner) wb11() (accept bool) {
	current := sc.buffer[sc.pos]
	if !is(Numeric, current) {
		return false
	}

	previous := sc.seekPreviousIndex(sc.pos, MidNum, MidNumLetQ)
	if previous < 0 {
		return false
	}

	return sc.seekPrevious(previous, Numeric)
}

// wb12 implements https://unicode.org/reports/tr29/#WB12
func (sc *Scanner) wb12() (accept bool) {
	current := sc.buffer[sc.pos]
	if !(is(MidNum, current) || is(MidNumLetQ, current)) {
		return false
	}

	if !sc.seekForward(Numeric) {
		return false
	}

	return sc.seekPrevious(sc.pos, Numeric)
}

// wb13 implements https://unicode.org/reports/tr29/#WB13
func (sc *Scanner) wb13() (accept bool) {
	current := sc.buffer[sc.pos]
	if !is(Katakana, current) {
		return false
	}

	return sc.seekPrevious(sc.pos, Katakana)
}

// wb13a implements https://unicode.org/reports/tr29/#WB13a
func (sc *Scanner) wb13a() (accept bool) {
	current := sc.buffer[sc.pos]
	if !is(ExtendNumLet, current) {
		return false
	}

	return sc.seekPrevious(sc.pos, _mergedAHLetterNumericKatakanaExtendNumLet)
}

// wb13b implements https://unicode.org/reports/tr29/#WB13b
func (sc *Scanner) wb13b() (accept bool) {
	current := sc.buffer[sc.pos]
	if !(is(_mergedAHLetterNumericKatakana, current)) {
		return false
	}

	return sc.seekPrevious(sc.pos, ExtendNumLet)
}

// wb15 implements https://unicode.org/reports/tr29/#WB15
func (sc *Scanner) wb15() (accept bool) {
	current := sc.buffer[sc.pos]
	if !is(Regional_Indicator, current) {
		return false
	}

	// Buffer comprised entirely of an odd number of RI, ignoring ExtendFormatZWJ
	count := 0
	for i := sc.pos - 1; i >= 0; i-- {
		r := sc.buffer[i]
		if is(_mergedExtendFormatZWJ, r) {
			continue
		}
		if !is(Regional_Indicator, r) {
			return false
		}
		count++
	}

	// If we fall through, we've seen the whole buffer,
	// so it's all Regional_Indicator | ExtendFormatZWJ
	odd := count > 0 && count%2 == 1
	return odd
}

// wb16 implements https://unicode.org/reports/tr29/#WB16
func (sc *Scanner) wb16() (accept bool) {
	current := sc.buffer[sc.pos]
	if !is(Regional_Indicator, current) {
		return false
	}

	// Last n runes represent an odd number of RI, ignoring ExtendFormatZWJ
	count := 0
	for i := sc.pos - 1; i >= 0; i-- {
		r := sc.buffer[i]
		if is(_mergedExtendFormatZWJ, r) {
			continue
		}
		if !is(Regional_Indicator, r) {
			odd := count > 0 && count%2 == 1
			return odd
		}
		count++
	}

	return false
}

// wb999 implements https://unicode.org/reports/tr29/#WB999
// i.e. word break
func (sc *Scanner) wb999() bool {
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
