// Package sentences provides a scanner for Unicode text segmentation sentence boundaries: https://unicode.org/reports/tr29/#Sentence_Boundaries
package sentences

import (
	"io"
	"unicode"

	"github.com/clipperhouse/uax29"
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
func NewScanner(r io.Reader) *uax29.Scanner {
	return uax29.NewScanner(r, BreakFunc)
}

var is = unicode.Is

// BreakFunc implements sentence boundaries according to https://unicode.org/reports/tr29/#Sentence_Boundaries.
// It is intended for use with uax29.Scanner.
var BreakFunc uax29.BreakFunc = func(buffer uax29.Runes, pos uax29.Pos) bool {
	// Rules: https://unicode.org/reports/tr29/#Sentence_Boundary_Rules

	sot := pos == 0 // "start of text"
	eof := len(buffer) == int(pos)

	// SB1
	if sot && !eof {
		return false
	}

	// SB2
	if eof {
		return true
	}

	current := buffer[pos]
	previous := buffer[pos-1]

	// SB3
	if is(LF, current) && is(CR, previous) {
		return false
	}

	// SB4
	if is(_ParaSep, previous) {
		return true
	}

	// SB5
	if is(_ExtendǀFormat, current) {
		return false
	}

	ignore := _ExtendǀFormat

	// SB6
	if is(Numeric, current) && buffer.SeekPrevious(pos, ignore, ATerm) {
		return false
	}

	// SB7
	if is(Upper, current) {
		previousIndex := buffer.SeekPreviousIndex(pos, ignore, ATerm)
		if previousIndex >= 0 && buffer.SeekPrevious(previousIndex, ignore, _UpperǀLower) {
			return false
		}
	}

	// SB8
	{
		p1 := pos

		// This loop is the 'regex':
		// ( ¬(OLetter | Upper | Lower | ParaSep | SATerm) )*
		for int(p1) < len(buffer) {
			r := buffer[p1]
			if is(_OLetterǀUpperǀLowerǀParaSepǀSATerm, r) {
				break
			}
			p1++
		}

		if buffer.SeekForward(p1-1, ignore, Lower) {
			p2 := pos

			sp := p2
			for {
				sp = buffer.SeekPreviousIndex(sp, ignore, Sp)
				if sp < 0 {
					break
				}
				p2 = sp
			}

			close := p2
			for {
				close = buffer.SeekPreviousIndex(close, ignore, Close)
				if close < 0 {
					break
				}
				p2 = close
			}

			if buffer.SeekPrevious(p2, ignore, ATerm) {
				return false
			}
		}
	}

	// SB8a
	if is(_SContinueǀSATerm, current) {
		pos := pos

		sp := pos
		for {
			sp = buffer.SeekPreviousIndex(sp, ignore, Sp)
			if sp < 0 {
				break
			}
			pos = sp
		}

		close := pos
		for {
			close = buffer.SeekPreviousIndex(close, ignore, Close)
			if close < 0 {
				break
			}
			pos = close
		}

		if buffer.SeekPrevious(pos, ignore, _SATerm) {
			return false
		}
	}

	// SB9
	if is(_CloseǀSpǀParaSep, current) {
		pos := pos

		close := pos
		for {
			close = buffer.SeekPreviousIndex(close, ignore, Close)
			if close < 0 {
				break
			}
			pos = close
		}

		if buffer.SeekPrevious(pos, ignore, _SATerm) {
			return false
		}
	}

	// SB10
	if is(_SpǀParaSep, current) {
		pos := pos

		sp := pos
		for {
			sp = buffer.SeekPreviousIndex(sp, ignore, Sp)
			if sp < 0 {
				break
			}
			pos = sp
		}

		close := pos
		for {
			close = buffer.SeekPreviousIndex(close, ignore, Close)
			if close < 0 {
				break
			}
			pos = close
		}

		if buffer.SeekPrevious(pos, ignore, _SATerm) {
			return false
		}
	}

	// SB11
	{
		pos := pos

		ps := buffer.SeekPreviousIndex(pos, ignore, _SpǀParaSep)
		if ps >= 0 {
			pos = ps
		}

		sp := pos
		for {
			sp = buffer.SeekPreviousIndex(sp, ignore, Sp)
			if sp < 0 {
				break
			}
			pos = sp
		}

		close := pos
		for {
			close = buffer.SeekPreviousIndex(close, ignore, Close)
			if close < 0 {
				break
			}
			pos = close
		}

		if buffer.SeekPrevious(pos, ignore, _SATerm) {
			return true
		}
	}

	// SB998
	if pos > 0 {
		return false
	}

	// If we fall through all the above rules, it's a sentence break
	return true
}
