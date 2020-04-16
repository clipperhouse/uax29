// Package is provide utilities for identifying Unicode categories of runes, relating to Unicode text segmentation:
// https://unicode.org/reports/tr29/
package is

import (
	"unicode"

	"golang.org/x/text/unicode/rangetable"
)

// Alphabetic is defined here: https://unicode.org/reports/tr44/#Alphabetic
func Alphabetic(r rune) bool {
	switch {
	case
		r == '_',
		unicode.IsLetter(r),
		unicode.Is(unicode.Nl, r),
		unicode.Is(unicode.Other_Alphabetic, r):
		return true
	}
	return false
}

// ALetter is defined here: https://unicode.org/reports/tr29/#ALetter
func ALetter(r rune) bool {
	// Logic of the above standard, from the bottom up

	switch {
	case
		HebrewLetter(r),
		unicode.Is(unicode.Hiragana, r),
		unicode.Is(unicode.Katakana, r),
		unicode.Is(unicode.Ideographic, r):
		return false
	}

	if unicode.Is(ALetterTable, r) {
		return true
	}

	return Alphabetic(r)
}

var ALetterTable = &unicode.RangeTable{
	R16: []unicode.Range16{
		{0x02C2, 0x02C5, 1},
		{0x02D2, 0x02D7, 1},
		{0x02DE, 0x02DF, 1},
		{0x02E5, 0x02EB, 1},
		{0x02ED, 0x02ED, 1},
		{0x02EF, 0x02FF, 1},
		{0x055A, 0x055C, 1},
		{0x055E, 0x055E, 1},
		{0x058A, 0x058A, 1},
		{0x05F3, 0x05F3, 1},
		{0xA708, 0xA716, 1},
		{0xA720, 0xA721, 1},
		{0xA789, 0xA78A, 1},
		{0xAB5B, 0xAB5B, 1},
	},
}

// AHLetter is ALetter or HebrewLetter
func AHLetter(r rune) bool {
	return ALetter(r) || HebrewLetter(r)
}

var midLetter = rangetable.New(
	//':',	// TODO Swedish
	'·',
	'·',
	'՟',
	'״',
	'‧',
	'︓',
	'﹕',
	'：',
)

// MidLetter is defined here: https://unicode.org/reports/tr29/#MidLetter
func MidLetter(r rune) bool {
	return unicode.Is(midLetter, r)
}

var midNumLet = rangetable.New(
	'.',
	'’',
	'․',
	'﹒',
	'＇',
	'．',
)

// MidNumLet is defined here: https://unicode.org/reports/tr29/#MidNumLet
func MidNumLet(r rune) bool {
	return unicode.Is(midNumLet, r)
}

// MidNumLetQ is defined here: https://unicode.org/reports/tr29/#MidNumLet
func MidNumLetQ(r rune) bool {
	return MidNumLet(r) || r == '\''
}

var infixNumeric = rangetable.New(
	0x002C,
	0x002E,
	0x003A,
	0x003B,
	0x037E,
	0x0589,
	0x060C,
	0x060D,
	0x07F8,
	0x2044,
	0xFE10,
	0xFE13,
	0xFE14,
)

// InfixNumeric is defined here: https://unicode.org/reports/tr14/
func InfixNumeric(r rune) bool {
	return unicode.Is(infixNumeric, r)
}

var notMidNum = rangetable.New(
	0x003A,
	0xFE13,
	0x002E,
)

var addMidNum = rangetable.New(
	0x066C,
	0xFE50,
	0xFE54,
	0xFF0C,
	0xFF1B,
)

// MidNum is defined here: https://unicode.org/reports/tr29/#MidNum
func MidNum(r rune) bool {
	return !unicode.Is(notMidNum, r) && (InfixNumeric(r) || unicode.Is(addMidNum, r))
}

// Numeric is defined here: https://unicode.org/reports/tr29/#Numeric
func Numeric(r rune) bool {
	switch {
	case r == '٬':
		return false
	case 0xFF10 <= r && r <= 0xFF19:
		return true
	default:
		return unicode.IsNumber(r)
	}
}

// Cr is carriage return (\r, 13)
func Cr(r rune) bool {
	return r == '\r'
}

// Lf is line feed (\n, 10)
func Lf(r rune) bool {
	return r == '\n'
}

var addKatakana = rangetable.New(
	0x3031,
	0x3032,
	0x3033,
	0x3034,
	0x3035,
	0x309B,
	0x309C,
	0x30A0,
	0x30FC,
	0xFF70,
)

// Katakana is defined here: https://unicode.org/reports/tr29/#Katakana
func Katakana(r rune) bool {
	return unicode.Is(unicode.Katakana, r) || unicode.Is(addKatakana, r)
}

// HebrewLetter is defined here: https://unicode.org/reports/tr29/#Hebrew_Letter
func HebrewLetter(r rune) bool {
	return unicode.Is(unicode.Hebrew, r) && unicode.IsLetter(r)
}

var newLine = rangetable.New(
	0x000B,
	0x000C,
	0x0085,
	0x2028,
	0x2029,
)

// Newline is defined here: https://unicode.org/reports/tr29/#WB3a
func Newline(r rune) bool {
	return unicode.Is(newLine, r)
}

// SingleQuote is defined here: https://unicode.org/reports/tr29/#Single_Quote
func SingleQuote(r rune) bool {
	return r == '\''
}

// DoubleQuote is defined here: https://unicode.org/reports/tr29/#Double_Quote
func DoubleQuote(r rune) bool {
	return r == '"'
}

// WSegSpace is defined here: https://unicode.org/reports/tr29/#WSegSpace
func WSegSpace(r rune) bool {
	return unicode.Is(unicode.Zs, r)
}

// ExtendNumLet is defined here: https://unicode.org/reports/tr29/#ExtendNumLetWB
func ExtendNumLet(r rune) bool {
	return unicode.Is(unicode.Pc, r) || r == 0x202F
}
