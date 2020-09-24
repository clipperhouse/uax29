// Package sentences provides a scanner for Unicode text segmentation sentence boundaries: https://unicode.org/reports/tr29/#Sentence_Boundaries
package sentences

import (
	"bufio"
	"io"
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

var trie = newSentencesTrie(0)

// is determines if lookup intersects propert(ies)
func (lookup property) is(properties property) bool {
	return (lookup & properties) != 0
}

var (
	_SATerm  = _STerm | _ATerm
	_ParaSep = _Sep | _CR | _LF
	_Ignore  = _Extend | _Format
)

// SplitFunc is a bufio.SplitFunc implementation of sentence segmentation, for use with bufio.Scanner
func SplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if len(data) == 0 {
		return 0, nil, nil
	}

	// These vars are stateful across loop iterations
	var pos, w int
	var current property

main:
	for {
		sot := pos == 0         // "start of text"
		eot := pos == len(data) // "end of text"

		if eot {
			if !atEOF {
				// Token extends past current data, request more
				return 0, nil, nil
			}

			// https://unicode.org/reports/tr29/#SB2
			break
		}

		/*
			We've switched the evaluation order of SB1↓ and SB2↑. It's ok:
			because we've checked for len(data) at the top of this function,
			sot and eot are mutually exclusive, order doesn't matter.
		*/

		// Rules are usually of the form Cat1 × Cat2; "current" refers to the first property
		// to the right of the ×, from which we look back or forward

		last := current

		current, w = trie.lookup(data[pos:])
		if w == 0 {
			if atEOF {
				// Just return the bytes, we can't do anything with them
				pos = len(data)
				break
			}
			// Rune extends past current data, request more
			return 0, nil, nil
		}

		// https://unicode.org/reports/tr29/#SB1
		if sot {
			pos += w
			continue
		}

		// Optimization: no rule can possibly apply
		if current|last == 0 { // i.e. both are zero
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#SB3
		if current.is(_LF) && last.is(_CR) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#SB4
		if last.is(_ParaSep) {
			break
		}

		// https://unicode.org/reports/tr29/#SB5
		if current.is(_Extend | _Format) {
			pos += w
			continue
		}

		// SB5 applies to subsequent rules; there is an implied "ignoring Extend & Format"
		// https://unicode.org/reports/tr29/#Grapheme_Cluster_and_Format_Rules
		// The previous/subsequent methods are shorthand for "seek a property but skip over Extend & Format on the way"

		// Optimization: determine if SB6 can possibly apply
		considerSB6 := current.is(_Numeric) && last.is(_ATerm|_Ignore)

		// https://unicode.org/reports/tr29/#SB6
		if considerSB6 {
			if previous(_ATerm, data[:pos]) {
				pos += w
				continue
			}
		}

		// Optimization: determine if SB7 can possibly apply
		considerSB7 := current.is(_Upper) && last.is(_ATerm|_Ignore)

		// https://unicode.org/reports/tr29/#SB7
		if considerSB7 {
			pi := previousIndex(_ATerm, data[:pos])
			if pi >= 0 && previous(_Upper|_Lower, data[:pi]) {
				pos += w
				continue
			}
		}

		// Optimization: determine if SB8 can possibly apply
		considerSB8 := last.is(_ATerm | _Close | _Sp | _Ignore)

		// https://unicode.org/reports/tr29/#SB8
		if considerSB8 {
			p := pos

			// ( ¬(OLetter | Upper | Lower | ParaSep | SATerm) )*
			// Zero or more of not-the-above properties
			for p < len(data) {
				lookup, w := trie.lookup(data[p:])
				if w == 0 {
					if atEOF {
						// Just return the bytes, we can't do anything with them
						pos = len(data)
						break main
					}
					// Rune extends past current data, request more
					return 0, nil, nil
				}

				if lookup.is(_OLetter | _Upper | _Lower | _ParaSep | _SATerm) {
					break
				}

				p += w
			}

			if subsequent(_Lower, data[p:]) {
				p2 := pos

				// Zero or more Sp
				sp := pos
				for {
					sp = previousIndex(_Sp, data[:sp])
					if sp < 0 {
						break
					}
					p2 = sp
				}

				// Zero or more Close
				close := p2
				for {
					close = previousIndex(_Close, data[:close])
					if close < 0 {
						break
					}
					p2 = close
				}

				// Having looked back past Sp's, Close's, and intervening Extend|Format,
				// is there an ATerm?
				if previous(_ATerm, data[:p2]) {
					pos += w
					continue
				}
			}
		}

		// Optimization: determine if SB8a can possibly apply
		considerSB8a := current.is(_SContinue|_SATerm) && last.is(_SATerm|_Close|_Sp|_Ignore)

		// https://unicode.org/reports/tr29/#SB8a
		if considerSB8a {
			p := pos

			// Zero or more Sp
			sp := p
			for {
				sp = previousIndex(_Sp, data[:sp])
				if sp < 0 {
					break
				}
				p = sp
			}

			// Zero or more Close
			close := p
			for {
				close = previousIndex(_Close, data[:close])
				if close < 0 {
					break
				}
				p = close
			}

			// Having looked back past Sp, Close, and intervening Extend|Format,
			// is there an SATerm?
			if previous(_SATerm, data[:p]) {
				pos += w
				continue
			}
		}

		// Optimization: determine if SB9 can possibly apply
		considerSB9 := current.is(_Close|_Sp|_ParaSep) && last.is(_SATerm|_Close|_Ignore)

		// https://unicode.org/reports/tr29/#SB9
		if considerSB9 {
			p := pos

			// Zero or more Close's
			close := p
			for {
				close = previousIndex(_Close, data[:close])
				if close < 0 {
					break
				}
				p = close
			}

			// Having looked back past Close's and intervening Extend|Format,
			// is there an SATerm?
			if previous(_SATerm, data[:p]) {
				pos += w
				continue
			}
		}

		// Optimization: determine if SB10 can possibly apply
		considerSB10 := current.is(_Sp|_ParaSep) && last.is(_SATerm|_Close|_Sp|_Ignore)

		// https://unicode.org/reports/tr29/#SB10
		if considerSB10 {
			p := pos

			// Zero or more Sp's
			sp := p
			for {
				sp = previousIndex(_Sp, data[:sp])
				if sp < 0 {
					break
				}
				p = sp
			}

			// Zero or more Close's
			close := p
			for {
				close = previousIndex(_Close, data[:close])
				if close < 0 {
					break
				}
				p = close
			}

			// Having looked back past Sp's, Close's, and intervening Extend|Format,
			// is there an SATerm?
			if previous(_SATerm, data[:p]) {
				pos += w
				continue
			}
		}

		// Optimization: determine if SB11 can possibly apply
		considerSB11 := last.is(_SATerm | _Close | _Sp | _ParaSep | _Ignore)

		// https://unicode.org/reports/tr29/#SB11
		if considerSB11 {
			p := pos

			// Zero or one ParaSep
			ps := previousIndex(_ParaSep, data[:p])
			if ps >= 0 {
				p = ps
			}

			// Zero or more Sp's
			sp := p
			for {
				sp = previousIndex(_Sp, data[:sp])
				if sp < 0 {
					break
				}
				p = sp
			}

			// Zero or more Close's
			close := p
			for {
				close = previousIndex(_Close, data[:close])
				if close < 0 {
					break
				}
				p = close
			}

			// Having looked back past ParaSep, Sp's, Close's, and intervening Extend|Format,
			// is there an SATerm?
			if previous(_SATerm, data[:p]) {
				break
			}
		}

		// https://unicode.org/reports/tr29/#SB998
		pos += w
	}

	// Return token
	return pos, data[:pos], nil
}
