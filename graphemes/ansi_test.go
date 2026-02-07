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
