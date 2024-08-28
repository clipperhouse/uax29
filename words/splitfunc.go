package words

var trie = newWordsTrie(0)

// is determines if lookup intersects propert(ies)
func (lookup property) is(properties property) bool {
	return (lookup & properties) != 0
}

const (
	_AHLetter   = _ALetter | _HebrewLetter
	_MidNumLetQ = _MidNumLet | _SingleQuote
	_Ignore     = _Extend | _Format | _ZWJ
)

// SplitFunc is a bufio.SplitFunc implementation of word segmentation, for use with bufio.Scanner.
func SplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if len(data) == 0 {
		return 0, nil, nil
	}

	// These vars are stateful across loop iterations
	var pos, w int
	var current property

	var lastExIgnore property     // "last excluding ignored categories"
	var lastLastExIgnore property // "the last one before that"
	var regionalIndicatorCount int

	// https://unicode.org/reports/tr29/#WB1
	{
		// Start of text always advances
		current, w = trie.lookup(data[pos:])
		if w == 0 {
			if !atEOF {
				// Rune extends past current data, request more
				return 0, nil, nil
			}
			pos = len(data)
			return pos, data[:pos], nil
		}

		pos += w
	}

	for {
		eot := pos == len(data) // "end of text"

		if eot {
			if !atEOF {
				// Token extends past current data, request more
				return 0, nil, nil
			}

			// https://unicode.org/reports/tr29/#WB2
			break
		}

		// Rules are usually of the form Cat1 × Cat2; "current" refers to the first property
		// to the right of the ×, from which we look back or forward

		last := current
		if !last.is(_Ignore) {
			lastLastExIgnore = lastExIgnore
			lastExIgnore = last
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
			break
		}

		// https://unicode.org/reports/tr29/#WB3
		if current.is(_LF) && last.is(_CR) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB3a
		// https://unicode.org/reports/tr29/#WB3b
		if (last | current).is(_Newline | _CR | _LF) {
			break
		}

		// https://unicode.org/reports/tr29/#WB3c
		if current.is(_ExtendedPictographic) && last.is(_ZWJ) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB3d
		if (current & last).is(_WSegSpace) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB4
		if current.is(_Extend | _Format | _ZWJ) {
			pos += w
			continue
		}

		// WB4 applies to subsequent rules; there is an implied "ignoring Extend & Format & ZWJ"
		// https://unicode.org/reports/tr29/#Grapheme_Cluster_and_Format_Rules
		// The previous/subsequent methods are shorthand for "seek a property but skip over Extend|Format|ZWJ on the way"

		// https://unicode.org/reports/tr29/#WB5
		if current.is(_AHLetter) && lastExIgnore.is(_AHLetter) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB6
		if current.is(_MidLetter|_MidNumLetQ) && lastExIgnore.is(_AHLetter) {
			found, more := subsequent(_AHLetter, data[pos+w:], atEOF)

			if more {
				// Token extends past current data, request more
				return 0, nil, nil
			}

			if found {
				pos += w
				continue
			}
		}

		// https://unicode.org/reports/tr29/#WB7
		if current.is(_AHLetter) && lastExIgnore.is(_MidLetter|_MidNumLetQ|_Ignore) && lastLastExIgnore.is(_AHLetter) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB7a
		if current.is(_SingleQuote) && lastExIgnore.is(_HebrewLetter) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB7b
		if current.is(_DoubleQuote) && lastExIgnore.is(_HebrewLetter|_Ignore) {
			found, more := subsequent(_HebrewLetter, data[pos+w:], atEOF)

			if more {
				// Token extends past current data, request more
				return 0, nil, nil
			}

			if found {
				pos += w
				continue
			}
		}

		// https://unicode.org/reports/tr29/#WB7c
		if current.is(_HebrewLetter) && lastExIgnore.is(_DoubleQuote|_Ignore) && lastLastExIgnore.is(_HebrewLetter) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB8
		// https://unicode.org/reports/tr29/#WB9
		// https://unicode.org/reports/tr29/#WB10
		if current.is(_Numeric|_AHLetter) && lastExIgnore.is(_Numeric|_AHLetter) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB11
		if current.is(_Numeric) && lastExIgnore.is(_MidNum|_MidNumLetQ) && lastLastExIgnore.is(_Numeric) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB12
		if current.is(_MidNum|_MidNumLetQ) && lastExIgnore.is(_Numeric) {
			found, more := subsequent(_Numeric, data[pos+w:], atEOF)

			if more {
				// Token extends past current data, request more
				return 0, nil, nil
			}

			if found {
				pos += w
				continue
			}
		}

		// https://unicode.org/reports/tr29/#WB13
		if current.is(_Katakana) && lastExIgnore.is(_Katakana) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB13a
		if current.is(_ExtendNumLet) && lastExIgnore.is(_AHLetter|_Numeric|_Katakana|_ExtendNumLet) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB13b
		if current.is(_AHLetter|_Numeric|_Katakana) && lastExIgnore.is(_ExtendNumLet) {
			pos += w
			continue
		}

		maybeWB1516 := current.is(_RegionalIndicator) && lastExIgnore.is(_RegionalIndicator)

		// https://unicode.org/reports/tr29/#WB15 and
		// https://unicode.org/reports/tr29/#WB16
		if maybeWB1516 {
			regionalIndicatorCount++

			odd := regionalIndicatorCount%2 == 1
			if odd {
				pos += w
				continue
			}
		}

		// https://unicode.org/reports/tr29/#WB999
		// If we fall through all the above rules, it's a word break
		break
	}

	return pos, data[:pos], nil
}
