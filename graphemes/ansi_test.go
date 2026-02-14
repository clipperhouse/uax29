package graphemes_test

import (
	"strings"
	"testing"

	"github.com/clipperhouse/uax29/v2/graphemes"
	"github.com/clipperhouse/uax29/v2/testdata"
)

func TestAnsiEscapeSequencesAsGraphemes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "SGR reset",
			input:    "\x1b[0m",
			expected: []string{"\x1b[0m"},
		},
		{
			name:     "SGR red then text",
			input:    "\x1b[31mhello",
			expected: []string{"\x1b[31m", "h", "e", "l", "l", "o"},
		},
		{
			name:     "CSI with params and final",
			input:    "\x1b[1;32m",
			expected: []string{"\x1b[1;32m"},
		},
		{
			name:     "OSC window title then BEL",
			input:    "\x1b]0;My Title\x07",
			expected: []string{"\x1b]0;My Title\x07"},
		},
		{
			name:     "OSC window title then ST",
			input:    "\x1b]0;Title\x1b\\",
			expected: []string{"\x1b]0;Title\x1b\\"},
		},
		{
			name:     "DCS with ST terminator",
			input:    "\x1bPq#0;2;0;0;0\x1b\\",
			expected: []string{"\x1bPq#0;2;0;0;0\x1b\\"},
		},
		{
			name:     "DCS with BEL in payload is not a single sequence",
			input:    "\x1bPq\x07rest",
			expected: []string{"\x1b", "P", "q", "\x07", "r", "e", "s", "t"},
		},
		{
			name:     "DCS canceled by CAN",
			input:    "\x1bPqdata\x18z",
			expected: []string{"\x1bPqdata", "\x18", "z"},
		},
		{
			name:     "DCS canceled immediately by CAN",
			input:    "\x1bP\x18z",
			expected: []string{"\x1bP", "\x18", "z"},
		},
		{
			name:     "DCS canceled by SUB",
			input:    "\x1bPqdata\x1az",
			expected: []string{"\x1bPqdata", "\x1a", "z"},
		},
		{
			name:     "SOS with ST terminator",
			input:    "\x1bXhello\x1b\\",
			expected: []string{"\x1bXhello\x1b\\"},
		},
		{
			name:     "SOS with BEL in payload is not a single sequence",
			input:    "\x1bXhi\x07",
			expected: []string{"\x1b", "X", "h", "i", "\x07"},
		},
		{
			name:     "PM with ST terminator",
			input:    "\x1b^msg\x1b\\",
			expected: []string{"\x1b^msg\x1b\\"},
		},
		{
			name:     "PM with BEL in payload is not a single sequence",
			input:    "\x1b^m\x07",
			expected: []string{"\x1b", "^", "m", "\x07"},
		},
		{
			name:     "APC with ST terminator",
			input:    "\x1b_data\x1b\\",
			expected: []string{"\x1b_data\x1b\\"},
		},
		{
			name:     "APC with BEL in payload is not a single sequence",
			input:    "\x1b_d\x07",
			expected: []string{"\x1b", "_", "d", "\x07"},
		},
		{
			name:     "OSC empty payload with BEL",
			input:    "\x1b]\x07",
			expected: []string{"\x1b]\x07"},
		},
		{
			name:     "DCS unterminated",
			input:    "\x1bPdata",
			expected: []string{"\x1b", "P", "d", "a", "t", "a"},
		},
		{
			name:     "OSC unterminated",
			input:    "\x1b]0;title",
			expected: []string{"\x1b", "]", "0", ";", "t", "i", "t", "l", "e"},
		},
		{
			name:     "OSC canceled by CAN",
			input:    "\x1b]0;title\x18x",
			expected: []string{"\x1b]0;title", "\x18", "x"},
		},
		{
			name:     "OSC canceled by SUB",
			input:    "\x1b]0;title\x1ax",
			expected: []string{"\x1b]0;title", "\x1a", "x"},
		},
		{
			name:     "two-byte Fe",
			input:    "\x1bD", // IND
			expected: []string{"\x1bD"},
		},
		{
			name:     "two-byte Fp DECSC",
			input:    "\x1b7",
			expected: []string{"\x1b7"},
		},
		{
			name:     "nF ESC SP F",
			input:    "\x1b F",
			expected: []string{"\x1b F"},
		},
		{
			name:     "mixed: CSI then letter",
			input:    "\x1b[mx",
			expected: []string{"\x1b[m", "x"},
		},
		{
			name:     "UTF-8 C1 CSI then text",
			input:    "\xC2\x9B31mhello",
			expected: []string{"\xC2\x9B31m", "h", "e", "l", "l", "o"},
		},
		{
			name:     "UTF-8 C1 OSC with UTF-8 C1 ST terminator",
			input:    "\xC2\x9D0;Title\xC2\x9C",
			expected: []string{"\xC2\x9D0;Title\xC2\x9C"},
		},
		{
			name:     "UTF-8 C1 OSC with 7-bit ST terminator",
			input:    "\xC2\x9D0;Title\x1b\\",
			expected: []string{"\xC2\x9D0;Title\x1b\\"},
		},
		{
			name:     "7-bit OSC with UTF-8 C1 ST terminator",
			input:    "\x1b]0;Title\xC2\x9C",
			expected: []string{"\x1b]0;Title\xC2\x9C"},
		},
		{
			name:     "UTF-8 C1 DCS with UTF-8 C1 ST terminator",
			input:    "\xC2\x90qpayload\xC2\x9C",
			expected: []string{"\xC2\x90qpayload\xC2\x9C"},
		},
		{
			name:     "UTF-8 C1 DCS canceled by CAN",
			input:    "\xC2\x90qpayload\x18x",
			expected: []string{"\xC2\x90qpayload", "\x18", "x"},
		},
		{
			name:     "UTF-8 C1 DCS with 7-bit ST terminator",
			input:    "\xC2\x90qpayload\x1b\\",
			expected: []string{"\xC2\x90qpayload\x1b\\"},
		},
		{
			name:     "7-bit DCS with UTF-8 C1 ST terminator",
			input:    "\x1bPqpayload\xC2\x9C",
			expected: []string{"\x1bPqpayload\xC2\x9C"},
		},
		{
			name:     "UTF-8 C1 Fe IND control",
			input:    "\xC2\x84",
			expected: []string{"\xC2\x84"},
		},
		{
			name:     "UTF-8 C1 lead byte for non-C1 codepoint is not ANSI",
			input:    "\u00A9",
			expected: []string{"\u00A9"},
		},
		{
			name:     "malformed CSI: param after intermediate",
			input:    "\x1b[ 1mok",
			expected: []string{"\x1b", "[", " ", "1", "m", "o", "k"},
		},
		{
			name:     "no ANSI at start",
			input:    "plain",
			expected: []string{"p", "l", "a", "i", "n"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			iter := graphemes.FromString(tt.input)
			iter.AnsiEscapeSequences = true
			var got []string
			for iter.Next() {
				got = append(got, iter.Value())
			}
			if len(got) != len(tt.expected) {
				t.Errorf("len(got)=%d len(expected)=%d\ngot %q\nexpected %q", len(got), len(tt.expected), got, tt.expected)
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("at %d: got %q expected %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestAnsiEscapeSequencesBytes(t *testing.T) {
	t.Parallel()
	input := []byte("\x1b[31mhi\x1b[0m")
	iter := graphemes.FromBytes(input)
	iter.AnsiEscapeSequences = true
	var got [][]byte
	for iter.Next() {
		got = append(got, iter.Value())
	}
	expected := [][]byte{
		[]byte("\x1b[31m"),
		[]byte("h"), []byte("i"),
		[]byte("\x1b[0m"),
	}
	if len(got) != len(expected) {
		t.Fatalf("len(got)=%d len(expected)=%d", len(got), len(expected))
	}
	for i := range got {
		if string(got[i]) != string(expected[i]) {
			t.Errorf("at %d: got %q expected %q", i, got[i], expected[i])
		}
	}
}

// ansiSample builds a string that mixes ANSI escape sequences with regular text,
// simulating realistic terminal output (colored words, resets, bold, etc.).
func ansiSample() string {
	var b strings.Builder

	// Simulate `ls --color=auto`-style output: colored filenames with resets
	colors := []string{
		"\x1b[1;34m", // bold blue (directories)
		"\x1b[0;32m", // green (executables)
		"\x1b[0;36m", // cyan (symlinks)
		"\x1b[1;31m", // bold red (errors)
		"\x1b[33m",   // yellow (warnings)
	}
	reset := "\x1b[0m"

	lines := []string{
		"drwxr-xr-x  5 user staff  160 Jan  1 12:00 Documents",
		"drwxr-xr-x  3 user staff   96 Feb  2 09:30 Downloads",
		"-rwxr-xr-x  1 user staff 8432 Mar 15 14:22 build.sh",
		"lrwxr-xr-x  1 user staff   11 Apr 20 08:00 config -> /etc/config",
		"-rw-r--r--  1 user staff 1024 May  5 16:45 README.md",
		"total 42",
		"drwxr-xr-x  2 user staff   64 Jun 10 11:11 src",
		"-rw-r--r--  1 user staff  512 Jul  7 07:07 main.go",
		"error: file not found: missing.txt",
		"warning: deprecated function used in line 42",
	}

	// Repeat to get a decent-sized sample
	for round := 0; round < 20; round++ {
		for i, line := range lines {
			color := colors[i%len(colors)]
			// OSC title update every 5 lines
			if i%5 == 0 {
				b.WriteString("\x1b]0;terminal - round ")
				b.WriteString(string(rune('0' + round%10)))
				b.WriteString("\x07")
			}
			b.WriteString(color)
			b.WriteString(line)
			b.WriteString(reset)
			b.WriteString("\n")
		}
	}
	return b.String()
}

// BenchmarkAnsiOption benchmarks the iterator on text that contains ANSI escapes,
// and on plain text, with the AnsiEscapeSequences option on and off.
func BenchmarkAnsiOption(b *testing.B) {
	ansi := ansiSample()
	plain, err := testdata.Sample()
	if err != nil {
		b.Fatal(err)
	}
	plainStr := string(plain)

	b.Run("AnsiText/OptionOn", func(b *testing.B) {
		b.SetBytes(int64(len(ansi)))
		for i := 0; i < b.N; i++ {
			iter := graphemes.FromString(ansi)
			iter.AnsiEscapeSequences = true
			c := 0
			for iter.Next() {
				_ = iter.Value()
				c++
			}
			b.ReportMetric(float64(c), "tokens")
		}
	})

	b.Run("AnsiText/OptionOff", func(b *testing.B) {
		b.SetBytes(int64(len(ansi)))
		for i := 0; i < b.N; i++ {
			iter := graphemes.FromString(ansi)
			c := 0
			for iter.Next() {
				_ = iter.Value()
				c++
			}
			b.ReportMetric(float64(c), "tokens")
		}
	})

	b.Run("PlainText/OptionOn", func(b *testing.B) {
		b.SetBytes(int64(len(plainStr)))
		for i := 0; i < b.N; i++ {
			iter := graphemes.FromString(plainStr)
			iter.AnsiEscapeSequences = true
			c := 0
			for iter.Next() {
				_ = iter.Value()
				c++
			}
			b.ReportMetric(float64(c), "tokens")
		}
	})

	b.Run("PlainText/OptionOff", func(b *testing.B) {
		b.SetBytes(int64(len(plainStr)))
		for i := 0; i < b.N; i++ {
			iter := graphemes.FromString(plainStr)
			c := 0
			for iter.Next() {
				_ = iter.Value()
				c++
			}
			b.ReportMetric(float64(c), "tokens")
		}
	})
}
