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

		// Rules: https://unicode.org/reports/tr29/#Word_Boundary_Rules

		sot := sc.pos == 0 // "start of text"
		eof := len(sc.buffer) == sc.pos

		// WB1
		if sot && !eof {
			sc.accept()
			continue
		}

		// WB2
		if eof {
			break
		}

		current := sc.buffer[sc.pos]
		previous := sc.buffer[sc.pos-1]

		// WB3
		if is(LF, current) && is(CR, previous) {
			sc.accept()
			continue
		}

		// WB3a
		if is(_CRǀLFǀNewline, previous) {
			break
		}

		// WB3b
		if is(_CRǀLFǀNewline, current) {
			break
		}

		// WB3c
		if is(emoji.Extended_Pictographic, current) && is(ZWJ, previous) {
			sc.accept()
			continue
		}

		// WB3d
		if is(WSegSpace, current) && is(WSegSpace, previous) {
			sc.accept()
			continue
		}

		// WB4
		if is(_ExtendǀFormatǀZWJ, current) {
			sc.accept()
			continue
		}

		// WB5
		if is(AHLetter, current) && sc.seekPrevious(sc.pos, AHLetter) {
			sc.accept()
			continue
		}

		// WB6
		if is(_MidLetterǀMidNumLetQ, current) && sc.seekForward(AHLetter) && sc.seekPrevious(sc.pos, AHLetter) {
			sc.accept()
			continue
		}

		// WB7
		if is(AHLetter, current) {
			previousIndex := sc.seekPreviousIndex(sc.pos, _MidLetterǀMidNumLetQ)
			if previousIndex >= 0 && sc.seekPrevious(previousIndex, AHLetter) {
				sc.accept()
				continue
			}
		}

		// WB7a
		if is(Single_Quote, current) && sc.seekPrevious(sc.pos, Hebrew_Letter) {
			sc.accept()
			continue
		}

		// WB7b
		if is(Double_Quote, current) && sc.seekForward(Hebrew_Letter) && sc.seekPrevious(sc.pos, Hebrew_Letter) {
			sc.accept()
			continue
		}

		// WB7c
		if is(Hebrew_Letter, current) {
			previousIndex := sc.seekPreviousIndex(sc.pos, Double_Quote)
			if previousIndex >= 0 && sc.seekPrevious(previousIndex, Hebrew_Letter) {
				sc.accept()
				continue
			}
		}

		// WB8
		if is(Numeric, current) && sc.seekPrevious(sc.pos, Numeric) {
			sc.accept()
			continue
		}

		// WB9
		if is(Numeric, current) && sc.seekPrevious(sc.pos, AHLetter) {
			sc.accept()
			continue
		}

		// WB10
		if is(AHLetter, current) && sc.seekPrevious(sc.pos, Numeric) {
			sc.accept()
			continue
		}

		// WB11
		if is(Numeric, current) {
			previousIndex := sc.seekPreviousIndex(sc.pos, _MidNumǀMidNumLetQ)
			if previousIndex >= 0 && sc.seekPrevious(previousIndex, Numeric) {
				sc.accept()
				continue
			}
		}

		// WB12
		if is(_MidNumǀMidNumLetQ, current) && sc.seekForward(Numeric) && sc.seekPrevious(sc.pos, Numeric) {
			sc.accept()
			continue
		}

		// WB13
		if is(Katakana, current) && sc.seekPrevious(sc.pos, Katakana) {
			sc.accept()
			continue
		}

		// WB13a
		if is(ExtendNumLet, current) && sc.seekPrevious(sc.pos, _AHLetterǀNumericǀKatakanaǀExtendNumLet) {
			sc.accept()
			continue
		}

		// WB13b
		if is(_AHLetterǀNumericǀKatakana, current) && sc.seekPrevious(sc.pos, ExtendNumLet) {
			sc.accept()
			continue
		}

		// WB15
		if is(Regional_Indicator, current) {
			ok := true

			// Buffer comprised entirely of an odd number of RI, ignoring Extend|Format|ZWJ
			count := 0
			for i := sc.pos - 1; i >= 0; i-- {
				r := sc.buffer[i]
				if is(_ExtendǀFormatǀZWJ, r) {
					continue
				}
				if !is(Regional_Indicator, r) {
					ok = false
					break
				}
				count++
			}

			// If we fall through, we've seen the whole buffer,
			// so it's all Regional_Indicator | Extend|Format|ZWJ
			if ok {
				odd := count > 0 && count%2 == 1
				if odd {
					sc.accept()
					continue
				}
			}
		}

		// WB16
		if is(Regional_Indicator, current) {
			ok := false
			// Last n runes represent an odd number of RI, ignoring Extend|Format|ZWJ
			count := 0
			for i := sc.pos - 1; i >= 0; i-- {
				r := sc.buffer[i]
				if is(_ExtendǀFormatǀZWJ, r) {
					continue
				}
				if !is(Regional_Indicator, r) {
					odd := count > 0 && count%2 == 1
					ok = odd
					break
				}
				count++
			}

			if ok {
				sc.accept()
				continue
			}
		}

		// WB999
		// If we fall through all the above rules, it's a word break
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

// Word boundary rules: https://unicode.org/reports/tr29/#Word_Boundaries
// In most cases, returning true means 'keep going'; check the name of the return var for clarity

var is = unicode.Is

// seekForward looks ahead until it hits a rune satisfying one of the range tables,
// ignoring Extend|Format|ZWJ
// See: https://unicode.org/reports/tr29/#Grapheme_Cluster_and_Format_Rules (driven by WB4)
func (sc *Scanner) seekForward(rts ...*unicode.RangeTable) bool {
	for i := sc.pos + 1; i < len(sc.buffer); i++ {
		r := sc.buffer[i]

		// Ignore Extend|Format|ZWJ
		if is(_ExtendǀFormatǀZWJ, r) {
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

		// Ignore Extend|Format|ZWJ
		if is(_ExtendǀFormatǀZWJ, r) {
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

// token returns the accrued token in the buffer
func (sc *Scanner) token() bool {
	var bb bytes.Buffer
	for _, r := range sc.buffer[:sc.pos] {
		bb.WriteRune(r)
	}
	sc.bytes = bb.Bytes()
	return len(sc.bytes) > 0
}

// accept forwards the buffer cursor (pos) by 1
func (sc *Scanner) accept() {
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
