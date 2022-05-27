package transform

import (
	"bytes"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type Func func([]byte) []byte

// Lower transforms text to lowercase
var Lower Func = bytes.ToLower

// Upper transforms text to uppercase
var Upper Func = bytes.ToUpper

// NFC normalizes Unicode text to the NFC form https://unicode.org/reports/tr15/#Norm_Forms
var NFC Func = norm.NFC.Bytes

// NFD normalizes Unicode text to the NFD form https://unicode.org/reports/tr15/#Norm_Forms
var NFD Func = norm.NFD.Bytes

// NFKC normalizes Unicode text to the NFKC form https://unicode.org/reports/tr15/#Norm_Forms
var NFKC Func = norm.NFKC.Bytes

// NFKD normalizes Unicode text to the NFKD form https://unicode.org/reports/tr15/#Norm_Forms
var NFKD Func = norm.NFKD.Bytes

// RemoveDiacritics 'flattens' characters with diacritics, such as accents. For example,
// café → cafe, façade → facade.
var RemoveDiacritics Func = func(b []byte) []byte {
	// https://stackoverflow.com/q/24588295
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.Bytes(t, b) // eliding the error, is this risky? a quick read seems like it's ok.
	return result
}
