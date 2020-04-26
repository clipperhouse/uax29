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

		sot := sc.pos == 0 // "start of text"
		eof := len(sc.buffer) == sc.pos

		// GB1
		if sot && !eof {
			sc.pos++
			continue
		}

		// GB2
		if eof {
			break
		}

		current := sc.buffer[sc.pos]
		previous := sc.buffer[sc.pos-1]

		// GB3
		if is(LF, current) && is(CR, previous) {
			sc.pos++
			continue
		}

		// GB4
		if is(_ControlǀCRǀLF, previous) {
			break
		}

		// GB5
		if is(_ControlǀCRǀLF, current) {
			break
		}

		// GB6
		if is(_LǀVǀLVǀLVT, current) && is(L, previous) {
			sc.pos++
			continue
		}

		// GB7
		if is(_VǀT, current) && is(_LVǀV, previous) {
			sc.pos++
			continue
		}

		// GB8
		if is(T, current) && is(_LVTǀT, previous) {
			sc.pos++
			continue
		}

		// GB9
		if is(_ExtendǀZWJ, current) {
			sc.pos++
			continue
		}

		// GB9a
		if is(SpacingMark, current) {
			sc.pos++
			continue
		}

		// GB9b
		if is(Prepend, previous) {
			sc.pos++
			continue
		}

		// GB11
		if is(emoji.Extended_Pictographic, current) && is(ZWJ, previous) && sc.seekPrevious(sc.pos-1, emoji.Extended_Pictographic) {
			sc.pos++
			continue
		}

		// GB12
		if is(Regional_Indicator, current) {
			// Buffer comprised entirely of an odd number of RI
			ok := true
			count := 0
			for i := sc.pos - 1; i >= 0; i-- {
				r := sc.buffer[i]
				if !is(Regional_Indicator, r) {
					ok = false
				}
				count++
			}

			if ok {
				// If we fall through, we've seen the whole buffer,
				// so it's all Regional_Indicator
				odd := count > 0 && count%2 == 1
				if odd {
					sc.pos++
					continue
				}
			}
		}

		// GB13
		if is(Regional_Indicator, current) {
			// Last n runes represent an odd number of RI
			ok := false
			count := 0
			for i := sc.pos - 1; i >= 0; i-- {
				r := sc.buffer[i]
				if !is(Regional_Indicator, r) {
					odd := count > 0 && count%2 == 1
					ok = odd
					break
				}
				count++
			}

			if ok {
				sc.pos++
				continue
			}
		}

		// WB999
		// If we fall through all the above rules, it's a break
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

// Grapheme cluster rules: https://unicode.org/reports/tr29/#Grapheme_Cluster_Boundaries
// In most cases, returning true means 'keep going'; check the name of the return var for clarity

var is = unicode.Is

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

// gb999 implements https://unicode.org/reports/tr29/#GB999
// i.e. break
func (sc *Scanner) token() bool {
	var bb bytes.Buffer
	for _, r := range sc.buffer[:sc.pos] {
		bb.WriteRune(r)
	}
	sc.bytes = bb.Bytes()
	return len(sc.bytes) > 0
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
