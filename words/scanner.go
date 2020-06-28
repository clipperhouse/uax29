// Package words provides a scanner for Unicode text segmentation word boundaries: https://unicode.org/reports/tr29/#Word_Boundaries
package words

import (
	"bufio"
	"io"
	"unicode/utf8"
)

// NewScanner tokenizes a reader into a stream of tokens according to Unicode Text Segmentation word boundaries https://unicode.org/reports/tr29/#Word_Boundaries.
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
func NewScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	scanner.Split(SplitFunc)
	return scanner
}

var trie = newWordsTrie(0)

// is determines if lookup intersects propert(ies)
func (lookup property) is(properties property) bool {
	return (lookup & properties) != 0
}

var (
	_AHLetter   = _ALetter | _HebrewLetter
	_MidNumLetQ = _MidNumLet | _SingleQuote
	_Ignore     = _Extend | _Format | _ZWJ
)

// SplitFunc is a bufio.SplitFunc implementation of word segmentation, for use with bufio.Scanner
func SplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if len(data) == 0 {
		return 0, nil, nil
	}

	// These vars are stateful across loop iterations
	var pos, w int
	var current property

	for {
		sot := pos == 0         // "start of text"
		eot := pos == len(data) // "end of text"

		if eot && !atEOF {
			// Token extends past current data, request more
			return 0, nil, nil
		}

		// https://unicode.org/reports/tr29/#WB1
		if sot {
			current, w = trie.lookup(data[pos:])
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB2
		if eot {
			break
		}

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

		next := pos + w

		// https://unicode.org/reports/tr29/#WB3
		if current.is(_LF) && last.is(_CR) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB3a
		if last.is(_CR | _LF | _Newline) {
			break
		}

		// https://unicode.org/reports/tr29/#WB3b
		if current.is(_CR | _LF | _Newline) {
			break
		}

		// https://unicode.org/reports/tr29/#WB3c
		if current.is(_ExtendedPictographic) && last.is(_ZWJ) {
			pos += w
			continue
		}

		// https://unicode.org/reports/tr29/#WB3d
		if current.is(_WSegSpace) && last.is(_WSegSpace) {
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

		// Optimization: determine if WB5 can possibly apply
		considerWB5 := current.is(_AHLetter) && last.is(_AHLetter|_Ignore)

		// https://unicode.org/reports/tr29/#WB5
		if considerWB5 {
			if previous(_AHLetter, data[:pos]) {
				pos += w

				// Optimization: there's a likelihood of a run of AHLetter
				for pos < len(data) {
					lookup, w2 := trie.lookup(data[pos:])

					if !lookup.is(_AHLetter) {
						break
					}

					// Update stateful vars
					current = lookup
					w = w2

					pos += w
				}

				continue
			}
		}

		// Optimization: determine if WB6 can possibly apply
		considerWB6 := current.is(_MidLetter|_MidNumLetQ) && last.is(_AHLetter|_Ignore)

		// https://unicode.org/reports/tr29/#WB6
		if considerWB6 {
			if subsequent(_AHLetter, data[next:]) && previous(_AHLetter, data[:pos]) {
				pos += w
				continue
			}
		}

		// Optimization: determine if WB7 can possibly apply
		considerWB7 := current.is(_AHLetter) && last.is(_MidLetter|_MidNumLetQ|_Ignore)

		// https://unicode.org/reports/tr29/#WB7
		if considerWB7 {
			i := previousIndex(_MidLetter|_MidNumLetQ, data[:pos])
			if i > 0 && previous(_AHLetter, data[:i]) {
				pos += w
				continue
			}
		}

		// Optimization: determine if WB7a can possibly apply
		considerWB7a := current.is(_SingleQuote) && last.is(_HebrewLetter|_Ignore)

		// https://unicode.org/reports/tr29/#WB7a
		if considerWB7a {
			if previous(_HebrewLetter, data[:pos]) {
				pos += w
				continue
			}
		}

		// Optimization: determine if WB7b can possibly apply
		considerWB7b := current.is(_DoubleQuote) && last.is(_HebrewLetter|_Ignore)

		// https://unicode.org/reports/tr29/#WB7b
		if considerWB7b {
			if subsequent(_HebrewLetter, data[next:]) && previous(_HebrewLetter, data[:pos]) {
				pos += w
				continue
			}
		}

		// Optimization: determine if WB7c can possibly apply
		considerWB7c := current.is(_HebrewLetter) && last.is(_DoubleQuote|_Ignore)

		// https://unicode.org/reports/tr29/#WB7c
		if considerWB7c {
			i := previousIndex(_DoubleQuote, data[:pos])
			if i > 0 && previous(_HebrewLetter, data[:i]) {
				pos += w
				continue
			}
		}

		// Optimization: determine if WB8 can possibly apply
		considerWB8 := current.is(_Numeric) && last.is(_Numeric|_Ignore)

		// https://unicode.org/reports/tr29/#WB8
		if considerWB8 {
			if previous(_Numeric, data[:pos]) {
				pos += w

				// Optimization: there's a likelihood of a run of Numeric
				for pos < len(data) {
					lookup, w2 := trie.lookup(data[pos:])

					if !lookup.is(_Numeric) {
						break
					}

					// Update stateful vars
					current = lookup
					w = w2

					pos += w
				}

				continue
			}
		}

		// Optimization: determine if WB9 can possibly apply
		considerWB9 := current.is(_Numeric) && last.is(_AHLetter|_Ignore)

		// https://unicode.org/reports/tr29/#WB9
		if considerWB9 {
			if previous(_AHLetter, data[:pos]) {
				pos += w
				continue
			}
		}

		// Optimization: determine if WB10 can possibly apply
		considerWB10 := current.is(_AHLetter) && last.is(_Numeric|_Ignore)

		// https://unicode.org/reports/tr29/#WB10
		if considerWB10 {
			if previous(_Numeric, data[:pos]) {
				pos += w
				continue
			}
		}

		// Optimization: determine if WB11 can possibly apply
		considerWB11 := current.is(_Numeric) && last.is(_MidNum|_MidNumLetQ|_Ignore)

		// https://unicode.org/reports/tr29/#WB11
		if considerWB11 {
			i := previousIndex(_MidNum|_MidNumLetQ, data[:pos])
			if i > 0 && previous(_Numeric, data[:i]) {
				pos += w
				continue
			}
		}

		// Optimization: determine if WB12 can possibly apply
		considerWB12 := current.is(_MidNum|_MidNumLet|_SingleQuote) && last.is(_Numeric|_Ignore)

		// https://unicode.org/reports/tr29/#WB12
		if considerWB12 {
			if subsequent(_Numeric, data[next:]) && previous(_Numeric, data[:pos]) {
				pos += w
				continue
			}
		}

		// Optimization: determine if WB13 can possibly apply
		considerWB13 := current.is(_Katakana) && last.is(_Katakana|_Ignore)

		// https://unicode.org/reports/tr29/#WB13
		if considerWB13 {
			if previous(_Katakana, data[:pos]) {
				pos += w

				// Optimization: there's a likelihood of a run of Katakana
				for pos < len(data) {
					lookup, w2 := trie.lookup(data[pos:])
					if !lookup.is(_Katakana) {
						break
					}

					// Update stateful vars
					current = lookup
					w = w2

					pos += w
				}

				continue
			}
		}

		// Optimization: determine if WB13a can possibly apply
		considerWB13a := current.is(_ExtendNumLet) && last.is(_AHLetter|_Numeric|_Katakana|_ExtendNumLet|_Ignore)

		// https://unicode.org/reports/tr29/#WB13a
		if considerWB13a {
			if previous(_AHLetter|_Numeric|_Katakana|_ExtendNumLet, data[:pos]) {
				pos += w
				continue
			}
		}

		// Optimization: determine if WB13b can possibly apply
		considerWB13b := current.is(_AHLetter|_Numeric|_Katakana) && last.is(_ExtendNumLet|_Ignore)

		// https://unicode.org/reports/tr29/#WB13b
		if considerWB13b {
			if previous(_ExtendNumLet, data[:pos]) {
				pos += w
				continue
			}
		}

		// Optimization: determine if WB15 or WB16 can possibly apply
		considerWB1516 := current.is(_RegionalIndicator) && last.is(_RegionalIndicator|_Ignore)

		// https://unicode.org/reports/tr29/#WB15 and
		// https://unicode.org/reports/tr29/#WB16
		if considerWB1516 {
			// WB15: Odd number of RI before hitting start of text
			// WB16: Odd number of RI before hitting [^RI], aka "not RI"

			i := pos
			count := 0

			for i > 0 {
				_, w := utf8.DecodeLastRune(data[:i])
				i -= w

				lookup, _ := trie.lookup(data[i:])

				if lookup.is(_Ignore) {
					continue
				}

				if !lookup.is(_RegionalIndicator) {
					// It's WB16
					break
				}

				count++
			}

			// If i == 0, we fell through and hit sot (start of text), so WB15 applies
			// If i > 0, we hit a non-RI, so WB16 applies

			// Note: I *suspect* WB16 is unreachable with the current logic.
			// A non-RI will have caused a word break on a previous pass,
			// by falling through to the break below (WB999). Therefore,
			// the non-RI will not be present in the current pass, so we will
			// always hit start-of-text, i.e., WB15.

			// The tests pass, however. This means there is an untested logical
			// flaw above, or WB16 is a redundant rule, or I'm simply mistaken.

			oddRI := count%2 == 1
			if oddRI {
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
