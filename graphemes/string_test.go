package graphemes_test

import (
	"bytes"
	"reflect"
	"testing"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/v2/graphemes"
	"github.com/clipperhouse/uax29/v2/testdata"
)

func TestStringUnicode(t *testing.T) {
	t.Parallel()

	// From the Unicode test suite; see the gen/ folder.
	var passed, failed int
	for _, test := range unicodeTests {
		test := test

		var all []string
		tokens := graphemes.FromString(string(test.input))
		for tokens.Next() {
			all = append(all, tokens.Value())
		}

		expected := make([]string, len(test.expected))
		for i, v := range test.expected {
			expected[i] = string(v)
		}

		if !reflect.DeepEqual(all, expected) {
			failed++
			t.Errorf(`
	for input %v
	expected  %v
	got       %v
	spec      %s`, test.input, test.expected, all, test.comment)
		} else {
			passed++
		}
	}

	if len(unicodeTests) != passed+failed {
		t.Errorf("Incomplete %d tests: passed %d, failed %d", len(unicodeTests), passed, failed)
	}
}

func TestStringRoundtrip(t *testing.T) {
	t.Parallel()

	const runs = 100

	for i := 0; i < runs; i++ {
		input := string(getRandomBytes())
		tokens := graphemes.FromString(input)

		var output string
		for tokens.Next() {
			output += tokens.Value()
		}

		if output != input {
			t.Fatal("input bytes are not the same as output bytes")
		}
	}
}

func TestStringInvalidUTF8(t *testing.T) {
	t.Parallel()

	// For background, see internal/testdata/UTF-8-test.txt, or:
	// https://www.cl.cam.ac.uk/~mgk25/ucs/examples/UTF-8-test.txt

	// Btw, don't edit UTF-8-test.txt: your editor might turn it into valid UTF-8!

	input, err := testdata.InvalidUTF8()

	if err != nil {
		t.Error(err)
	}

	if utf8.Valid(input) {
		t.Error("input file should not be valid utf8")
	}

	tokens := graphemes.FromString(string(input))

	var output string
	for tokens.Next() {
		output += tokens.Value()
	}

	if output != string(input) {
		t.Fatalf("input bytes are not the same as output bytes")
	}
}

func TestStringUnicode16ForwardCompatibility(t *testing.T) {
	t.Parallel()

	// Test cases for Unicode 16.0 characters that should be segmented as individual graphemes
	// These characters were introduced in Unicode 16.0 and should be handled gracefully
	// even though this package was built for Unicode 15
	testCases := []struct {
		name     string
		input    string
		expected []string
		comment  string
	}{
		{
			name:     "Single Garay Letter",
			input:    "𐍀", // U+10D40 GARAY LETTER KA
			expected: []string{"𐍀"},
			comment:  "Single Garay script character should be one grapheme",
		},
		{
			name:     "Single Gurung Khema Letter",
			input:    "𖌀", // U+16100 GURUNG KHEMA LETTER A
			expected: []string{"𖌀"},
			comment:  "Single Gurung Khema script character should be one grapheme",
		},
		{
			name:     "Single Egyptian Hieroglyph",
			input:    "𓍠", // U+13460 EGYPTIAN HIEROGLYPH A001
			expected: []string{"𓍠"},
			comment:  "Single Egyptian Hieroglyph should be one grapheme",
		},
		{
			name:     "Single Legacy Computing Symbol",
			input:    "Ⲁ", // U+1CC00 LEGACY COMPUTING SYMBOL
			expected: []string{"Ⲁ"},
			comment:  "Single Legacy Computing symbol should be one grapheme",
		},
		{
			name:     "Multiple Unicode 16 Characters",
			input:    "𐍀𖌀𓍠", // Garay + Gurung Khema + Egyptian
			expected: []string{"𐍀", "𖌀", "𓍠"},
			comment:  "Multiple Unicode 16 characters should be separate graphemes",
		},
		{
			name:     "Unicode 16 with ASCII",
			input:    "A𐍀B", // ASCII + Garay + ASCII
			expected: []string{"A", "𐍀", "B"},
			comment:  "Unicode 16 characters should be segmented correctly with ASCII",
		},
		{
			name:     "Unicode 16 with Combining Marks",
			input:    "𐍀́", // Garay letter with combining acute accent
			expected: []string{"𐍀́"},
			comment:  "Unicode 16 character with combining mark should be one grapheme",
		},
		{
			name:     "Mixed Scripts with Unicode 16",
			input:    "Hello𐍀世界", // English + Garay + Chinese
			expected: []string{"H", "e", "l", "l", "o", "𐍀", "世", "界"},
			comment:  "Unicode 16 characters should work with mixed scripts",
		},
		{
			name:     "Unicode 16 with Emoji",
			input:    "𐍀😀𖌀", // Garay + Emoji + Gurung Khema
			expected: []string{"𐍀", "😀", "𖌀"},
			comment:  "Unicode 16 characters should work with emoji",
		},
		{
			name:     "Empty String",
			input:    "",
			expected: nil, // Empty string should produce nil slice
			comment:  "Empty string should produce no graphemes",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Test that the input is valid UTF-8
			if !utf8.ValidString(tc.input) {
				t.Errorf("Input string is not valid UTF-8: %q", tc.input)
				return
			}

			// Test grapheme segmentation
			var actual []string
			tokens := graphemes.FromString(tc.input)
			for tokens.Next() {
				actual = append(actual, tokens.Value())
			}

			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf(`
	Input:    %q
	Expected: %v
	Got:      %v
	Comment:  %s`, tc.input, tc.expected, actual, tc.comment)
			}

			// Test roundtrip - input should equal concatenated output
			reconstructed := ""
			for _, grapheme := range actual {
				reconstructed += grapheme
			}
			if reconstructed != tc.input {
				t.Errorf("Roundtrip failed: input %q != reconstructed %q", tc.input, reconstructed)
			}
		})
	}
}

func BenchmarkString(b *testing.B) {
	file, err := testdata.Sample()
	if err != nil {
		b.Error(err)
	}

	s := string(file)

	b.ResetTimer()
	b.SetBytes(int64(len(file)))

	for i := 0; i < b.N; i++ {
		tokens := graphemes.FromString(s)

		c := 0
		for tokens.Next() {
			_ = tokens.Value()
			c++
		}

		b.ReportMetric(float64(c), "tokens")
	}
}

func BenchmarkStringUnicodeTests(b *testing.B) {
	var buf bytes.Buffer
	for _, test := range unicodeTests {
		buf.Write(test.input)
	}
	file := buf.Bytes()
	s := string(file)

	b.ResetTimer()
	b.SetBytes(int64(len(file)))

	for i := 0; i < b.N; i++ {
		tokens := graphemes.FromString(s)

		c := 0
		for tokens.Next() {
			_ = tokens.Value()
			c++
		}

		b.ReportMetric(float64(c), "tokens")
	}
}
