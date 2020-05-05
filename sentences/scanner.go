// Package sentences provides a scanner for Unicode text segmentation sentence boundaries: https://unicode.org/reports/tr29/#Sentence_Boundaries
package sentences

import (
	"bufio"
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/seek"
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
func NewScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	scanner.Split(SplitFunc)
	return scanner
}

var is = unicode.Is

// SplitFunc is a bufio.SplitFunc implementation of sentence segmentation, for use with bufio.Scanner
func SplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	pos := 0

	for {
		if pos == len(data) && !atEOF {
			// Request more data
			return 0, nil, nil
		}

		sot := pos == 0 // "start of text"
		eof := len(data) == pos

		// https://unicode.org/reports/tr29/#SB1
		if sot && !eof {
			_, w := utf8.DecodeRune(data[pos:])
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#SB2
		if eof {
			break
		}

		// Rules are usually of the form Cat1 × Cat2; "current" refers to the first category
		// to the right of the ×, from which we look back or forward

		current, w := utf8.DecodeRune(data[pos:])
		if current == utf8.RuneError {
			return 0, nil, fmt.Errorf("error decoding rune")
		}

		previous, _ := utf8.DecodeLastRune(data[:pos])

		// https://unicode.org/reports/tr29/#SB3
		if is(LF, current) && is(CR, previous) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#SB4
		if is(_ParaSep, previous) {
			break
		}

		// https://unicode.org/reports/tr29/#SB5
		if is(_ExtendǀFormat, current) {
			pos += w
			continue
		}

		// SB5 applies to subsequent rules; there is an implied "ignoring Extend & Format"
		// https://unicode.org/reports/tr29/#Grapheme_Cluster_and_Format_Rules
		// The Seek* methods are shorthand for "seek a category but skip over Extend & Format on the way"

		ignore := _ExtendǀFormat

		// https://unicode.org/reports/tr29/#SB6
		if is(Numeric, current) && seek.Previous(data[:pos], ignore, ATerm) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#SB7
		if is(Upper, current) {
			previousIndex := seek.PreviousIndex(data[:pos], ignore, ATerm)
			if previousIndex >= 0 && seek.Previous(data[:previousIndex], ignore, _UpperǀLower) {
				pos += w
				continue
			}
		}

		// https://unicode.org/reports/tr29/#SB8
		{
			p := pos

			// ( ¬(OLetter | Upper | Lower | ParaSep | SATerm) )*
			// Zero or more of not-the-above categories
			for p < len(data) {
				r, w := utf8.DecodeRune(data[p:])
				if !is(_OLetterǀUpperǀLowerǀParaSepǀSATerm, r) {
					p += w
					continue
				}
				break
			}

			if seek.Forward(data[p:], ignore, Lower) {
				p2 := pos

				// Zero or more Sp
				sp := pos
				for {
					sp = seek.PreviousIndex(data[:sp], ignore, Sp)
					if sp < 0 {
						break
					}
					p2 = sp
				}

				// Zero or more Close
				close := p2
				for {
					close = seek.PreviousIndex(data[:close], ignore, Close)
					if close < 0 {
						break
					}
					p2 = close
				}

				// Having looked back past Sp's, Close's, and intervening Extend|Format,
				// is there an ATerm?
				if seek.Previous(data[:p2], ignore, ATerm) {
					pos += w
					continue
				}
			}
		}

		// https://unicode.org/reports/tr29/#SB8a
		if is(_SContinueǀSATerm, current) {
			p := pos

			// Zero or more Sp
			sp := p
			for {
				sp = seek.PreviousIndex(data[:sp], ignore, Sp)
				if sp < 0 {
					break
				}
				p = sp
			}

			// Zero or more Close
			close := p
			for {
				close = seek.PreviousIndex(data[:close], ignore, Close)
				if close < 0 {
					break
				}
				p = close
			}

			// Having looked back past Sp, Close, and intervening Extend|Format,
			// is there an SATerm?
			if seek.Previous(data[:p], ignore, _SATerm) {
				pos += w
				continue
			}
		}

		// https://unicode.org/reports/tr29/#SB9
		if is(_CloseǀSpǀParaSep, current) {
			p := pos

			// Zero or more Close's
			close := p
			for {
				close = seek.PreviousIndex(data[:close], ignore, Close)
				if close < 0 {
					break
				}
				p = close
			}

			// Having looked back past Close's and intervening Extend|Format,
			// is there an SATerm?
			if seek.Previous(data[:p], ignore, _SATerm) {
				pos += w
				continue
			}
		}

		// https://unicode.org/reports/tr29/#SB10
		if is(_SpǀParaSep, current) {
			p := pos

			// Zero or more Sp's
			sp := p
			for {
				sp = seek.PreviousIndex(data[:sp], ignore, Sp)
				if sp < 0 {
					break
				}
				p = sp
			}

			// Zero or more Close's
			close := p
			for {
				close = seek.PreviousIndex(data[:close], ignore, Close)
				if close < 0 {
					break
				}
				p = close
			}

			// Having looked back past Sp's, Close's, and intervening Extend|Format,
			// is there an SATerm?
			if seek.Previous(data[:p], ignore, _SATerm) {
				pos += w
				continue
			}
		}

		// https://unicode.org/reports/tr29/#SB11
		{
			p := pos

			// Zero or one Sp|ParaSep
			ps := seek.PreviousIndex(data[:p], ignore, _SpǀParaSep)
			if ps >= 0 {
				p = ps
			}

			// Zero or more Sp's
			sp := p
			for {
				sp = seek.PreviousIndex(data[:sp], ignore, Sp)
				if sp < 0 {
					break
				}
				p = sp
			}

			// Zero or more Close's
			close := p
			for {
				close = seek.PreviousIndex(data[:close], ignore, Close)
				if close < 0 {
					break
				}
				p = close
			}

			// Having looked back past Sp|ParaSep, Sp's, Close's, and intervening Extend|Format,
			// is there an SATerm?
			if seek.Previous(data[:p], ignore, _SATerm) {
				break
			}
		}

		// https://unicode.org/reports/tr29/#SB998
		if pos > 0 {
			pos += w
			continue
		}

		// If we fall through all the above rules, it's a sentence break
		break
	}

	// Return token
	return pos, data[:pos], nil
}
