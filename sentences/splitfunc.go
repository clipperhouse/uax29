package sentences

var trie = newSentencesTrie(0)

// is determines if lookup intersects propert(ies)
func (lookup property) is(properties property) bool {
	return (lookup & properties) != 0
}

const (
	_SATerm  = _STerm | _ATerm
	_ParaSep = _Sep | _CR | _LF
	_Ignore  = _Extend | _Format
)

// SplitFunc is a bufio.SplitFunc implementation of word segmentation, for use with bufio.Scanner.
func SplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if len(data) == 0 {
		return 0, nil, nil
	}

	// These vars are stateful across loop iterations
	var pos int
	var lastExIgnore property     // "last excluding ignored categories"
	var lastLastExIgnore property // "last one before that"
	var lastExIgnoreSp property
	var lastExIgnoreClose property
	var lastExIgnoreSpClose property

	// Rules are usually of the form Cat1 × Cat2; "current" refers to the first property
	// to the right of the ×, from which we look back or forward

	current, w := trie.lookup(data[pos:])
	if w == 0 {
		if !atEOF {
			// Rune extends past current data, request more
			return 0, nil, nil
		}
		pos = len(data)
		return pos, data[:pos], nil
	}

	// https://unicode.org/reports/tr29/#SB1
	// Start of text always advances
	pos += w

main:
	for {
		eot := pos == len(data) // "end of text"

		if eot {
			if !atEOF {
				// Token extends past current data, request more
				return 0, nil, nil
			}

			// https://unicode.org/reports/tr29/#SB2
			break
		}

		// Remember previous properties to avoid lookups/lookbacks
		last := current

		if !last.is(_Ignore) {
			lastLastExIgnore = lastExIgnore
			lastExIgnore = last
		}

		if !lastExIgnore.is(_Sp) {
			lastExIgnoreSp = lastExIgnore
		}

		if !lastExIgnore.is(_Close) {
			lastExIgnoreClose = lastExIgnore
		}

		if !lastExIgnoreSp.is(_Close) {
			lastExIgnoreSpClose = lastExIgnoreSp
		}

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
		// https://unicode.org/reports/tr29/#Sentence_Boundary_Rules
		// The previous/subsequent methods are shorthand for "seek a property but skip over Extend & Format on the way"

		// https://unicode.org/reports/tr29/#SB6
		if current.is(_Numeric) && lastExIgnore.is(_ATerm) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#SB7
		if current.is(_Upper) && lastExIgnore.is(_ATerm) && lastLastExIgnore.is(_Upper|_Lower) {
			pos += w
			continue
		}

		// Optimization: determine if SB8 can possibly apply
		maybeSB8 := lastExIgnoreSpClose.is(_ATerm)

		// https://unicode.org/reports/tr29/#SB8
		if maybeSB8 {
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

			found, more := subsequent(_Lower, data[p:], atEOF)

			if more {
				// Rune or token extends past current data, request more
				return 0, nil, nil
			}

			if found {
				pos += w
				continue
			}
		}

		// https://unicode.org/reports/tr29/#SB8a
		if current.is(_SContinue|_SATerm) && lastExIgnoreSpClose.is(_SATerm) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#SB9
		if current.is(_Close|_Sp|_ParaSep) && lastExIgnoreClose.is(_SATerm) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#SB10
		if current.is(_Sp|_ParaSep) && lastExIgnoreSpClose.is(_SATerm) {
			pos += w
			continue
		}

		// Optimization: determine if SB11 can possibly apply
		maybeSB11 := lastExIgnore.is(_SATerm | _Close | _Sp | _ParaSep)

		// https://unicode.org/reports/tr29/#SB11
		if maybeSB11 {
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
