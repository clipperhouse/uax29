// Package sentences provides a scanner for Unicode text segmentation sentence boundaries: https://unicode.org/reports/tr29/#Sentence_Boundaries
package sentences

import (
	"bufio"
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
func BreakFunc(buffer uax29.Runes, pos uax29.Pos) bool {

	// Rules: https://unicode.org/reports/tr29/#Sentence_Boundary_Rules
	// true = breaking, false = accept and continue

	sot := pos == 0 // "start of text"
	eof := len(buffer) == int(pos)

	// https://unicode.org/reports/tr29/#SB1
	if sot && !eof {
		return uax29.Accept
	}

	// https://unicode.org/reports/tr29/#SB2
	if eof {
		return uax29.Break
	}

	// Rules are usually of the form Cat1 × Cat2; "current" refers to the first category
	// to the right of the ×, from which we look back or forward

	current := buffer[pos]
	previous := buffer[pos-1]

	// https://unicode.org/reports/tr29/#SB3
	if is(LF, current) && is(CR, previous) {
		return uax29.Accept
	}

	// https://unicode.org/reports/tr29/#SB4
	if is(_ParaSep, previous) {
		return uax29.Break
	}

	// https://unicode.org/reports/tr29/#SB5
	if is(_ExtendǀFormat, current) {
		return uax29.Accept
	}

	// SB5 applies to subsequent rules; there is an implied "ignoring Extend & Format"
	// https://unicode.org/reports/tr29/#Grapheme_Cluster_and_Format_Rules
	// The Seek* methods are shorthand for "seek a category but skip over Extend & Format on the way"

	ignore := _ExtendǀFormat

	// https://unicode.org/reports/tr29/#SB6
	if is(Numeric, current) && buffer.SeekPrevious(pos, ignore, ATerm) {
		return uax29.Accept
	}

	// https://unicode.org/reports/tr29/#SB7
	if is(Upper, current) {
		previousIndex := buffer.SeekPreviousIndex(pos, ignore, ATerm)
		if previousIndex >= 0 && buffer.SeekPrevious(previousIndex, ignore, _UpperǀLower) {
			return uax29.Accept
		}
	}

	// https://unicode.org/reports/tr29/#SB8
	{
		p := pos

		// ( ¬(OLetter | Upper | Lower | ParaSep | SATerm) )*
		// Zero or more of not-the-above categories
		for int(p) < len(buffer) {
			r := buffer[p]
			if !is(_OLetterǀUpperǀLowerǀParaSepǀSATerm, r) {
				p++
				continue
			}
			break
		}

		if buffer.SeekForward(p-1, ignore, Lower) {
			p2 := pos

			// Zero or more Sp
			sp := pos
			for {
				sp = buffer.SeekPreviousIndex(sp, ignore, Sp)
				if sp < 0 {
					break
				}
				p2 = sp
			}

			// Zero or more Close
			close := p2
			for {
				close = buffer.SeekPreviousIndex(close, ignore, Close)
				if close < 0 {
					break
				}
				p2 = close
			}

			// Having looked back past Sp's, Close's, and intervening Extend|Format,
			// is there an ATerm?
			if buffer.SeekPrevious(p2, ignore, ATerm) {
				return uax29.Accept
			}
		}
	}

	// https://unicode.org/reports/tr29/#SB8a
	if is(_SContinueǀSATerm, current) {
		p := pos

		// Zero or more Sp
		sp := p
		for {
			sp = buffer.SeekPreviousIndex(sp, ignore, Sp)
			if sp < 0 {
				break
			}
			p = sp
		}

		// Zero or more Close
		close := p
		for {
			close = buffer.SeekPreviousIndex(close, ignore, Close)
			if close < 0 {
				break
			}
			p = close
		}

		// Having looked back past Sp, Close, and intervening Extend|Format,
		// is there an SATerm?
		if buffer.SeekPrevious(p, ignore, _SATerm) {
			return uax29.Accept
		}
	}

	// https://unicode.org/reports/tr29/#SB9
	if is(_CloseǀSpǀParaSep, current) {
		p := pos

		// Zero or more Close's
		close := p
		for {
			close = buffer.SeekPreviousIndex(close, ignore, Close)
			if close < 0 {
				break
			}
			p = close
		}

		// Having looked back past Close's and intervening Extend|Format,
		// is there an SATerm?
		if buffer.SeekPrevious(p, ignore, _SATerm) {
			return uax29.Accept
		}
	}

	// https://unicode.org/reports/tr29/#SB10
	if is(_SpǀParaSep, current) {
		p := pos

		// Zero or more Sp's
		sp := p
		for {
			sp = buffer.SeekPreviousIndex(sp, ignore, Sp)
			if sp < 0 {
				break
			}
			p = sp
		}

		// Zero or more Close's
		close := p
		for {
			close = buffer.SeekPreviousIndex(close, ignore, Close)
			if close < 0 {
				break
			}
			p = close
		}

		// Having looked back past Sp's, Close's, and intervening Extend|Format,
		// is there an SATerm?
		if buffer.SeekPrevious(p, ignore, _SATerm) {
			return uax29.Accept
		}
	}

	// https://unicode.org/reports/tr29/#SB11
	{
		p := pos

		// Zero or one Sp|ParaSep
		ps := buffer.SeekPreviousIndex(p, ignore, _SpǀParaSep)
		if ps >= 0 {
			p = ps
		}

		// Zero or more Sp's
		sp := p
		for {
			sp = buffer.SeekPreviousIndex(sp, ignore, Sp)
			if sp < 0 {
				break
			}
			p = sp
		}

		// Zero or more Close's
		close := p
		for {
			close = buffer.SeekPreviousIndex(close, ignore, Close)
			if close < 0 {
				break
			}
			p = close
		}

		// Having looked back past Sp|ParaSep, Sp's, Close's, and intervening Extend|Format,
		// is there an SATerm?
		if buffer.SeekPrevious(p, ignore, _SATerm) {
			return uax29.Break
		}
	}

	// https://unicode.org/reports/tr29/#SB998
	if pos > 0 {
		return uax29.Accept
	}

	// If we fall through all the above rules, it's a sentence break
	return uax29.Break
}

// SplitFunc is a bufio.SplitFunc implementation of sentence segmentation, for use with bufio.Scanner
var SplitFunc bufio.SplitFunc

func init() {
	SplitFunc = uax29.NewSplitFunc(BreakFunc)
}
