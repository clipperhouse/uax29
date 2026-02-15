package comparative

import (
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/clipperhouse/uax29/v2/graphemes"
	"github.com/clipperhouse/uax29/v2/testdata"
	"github.com/rivo/uniseg"
)

func BenchmarkGraphemesMixed(b *testing.B) {
	data, err := testdata.Sample()
	if err != nil {
		b.Fatal(err)
	}
	text := string(data)
	n := int64(len(text))

	b.Run("clipperhouse/uax29", func(b *testing.B) {
		b.SetBytes(n)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			count := 0
			g := graphemes.FromString(text)
			for g.Next() {
				count++
			}
		}
	})

	b.Run("rivo/uniseg", func(b *testing.B) {
		b.SetBytes(n)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			count := 0
			g := uniseg.NewGraphemes(text)
			for g.Next() {
				count++
			}
		}
	})
}

func BenchmarkGraphemesASCII(b *testing.B) {
	// Pure ASCII text - should benefit from ASCII hot path
	ascii := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 100)
	n := int64(len(ascii))

	b.Run("clipperhouse/uax29", func(b *testing.B) {
		b.SetBytes(n)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			count := 0
			g := graphemes.FromString(ascii)
			for g.Next() {
				count++
			}
		}
	})

	b.Run("rivo/uniseg", func(b *testing.B) {
		b.SetBytes(n)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			count := 0
			g := uniseg.NewGraphemes(ascii)
			for g.Next() {
				count++
			}
		}
	})
}

// TestAnsiBoundaryAgreement verifies that our ANSI sequence parsing produces
// the same token boundaries as charmbracelet/x/ansi's DecodeSequence.
// Inputs use ASCII text between sequences so grapheme clustering differences
// don't obscure ANSI boundary comparison.
func TestAnsiBoundaryAgreement(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		// 7-bit CSI
		{"SGR reset", "\x1b[0m"},
		{"SGR color then text then reset", "\x1b[31mhello\x1b[0m"},
		{"CSI bold+color", "\x1b[1;32m"},
		{"CSI cursor position", "\x1b[10;20H"},

		// 7-bit OSC
		{"OSC title with BEL", "\x1b]0;My Title\x07"},
		{"OSC title with ST", "\x1b]0;Title\x1b\\"},

		// 7-bit DCS/SOS/PM/APC
		{"DCS with ST", "\x1bPq#0;2;0;0;0\x1b\\"},
		{"SOS with ST", "\x1bXhello\x1b\\"},
		{"PM with ST", "\x1b^msg\x1b\\"},
		{"APC with ST", "\x1b_data\x1b\\"},

		// Two-byte Fe/Fs/Fp
		{"Fe IND", "\x1bD"},
		{"Fs RIS", "\x1bc"},
		{"Fp DECSC", "\x1b7"},

		// C1 8-bit
		{"C1 CSI then text", "\x9B31mhello"},
		{"C1 OSC with C1 ST", "\x9D0;Title\x9C"},
		{"C1 OSC with 7-bit ST", "\x9D0;Title\x1b\\"},
		{"C1 DCS with 7-bit ST", "\x90qpayload\x1b\\"},
		{"C1 DCS with C1 ST", "\x90qpayload\x9C"},
		{"C1 SOS with C1 ST", "\x98hello\x9C"},
		{"C1 PM with 7-bit ST", "\x9Emsg\x1b\\"},
		{"C1 APC with C1 ST", "\x9Fdata\x9C"},

		// CSI variants (from charmbracelet test suite)
		{"CSI private mode", "\x1b[?1049h"},
		{"CSI subparams (colons)", "\x1b[38:2:255:0:255;1m"},
		{"CSI with intermediate", "\x1b[0 q"},
		{"CSI no params", "\x1b[m"},
		{"CSI mouse click", "\x1b[<0;1;1M"},
		{"CSI mouse wheel", "\x1b[<64;2;11m"},
		{"CSI bracketed paste on", "\x1b[?2004h"},
		{"CSI bracketed paste content", "\x1b[200~pasted text\x1b[201~"},

		// SS3 / SS2 (Single Shift)
		{"SS3 7-bit", "\x1bOA"},
		{"SS3 8-bit", "\x8fA"},
		{"SS2 7-bit", "\x1bNA"},
		{"SS2 8-bit", "\x8eA"},

		// nF sequences
		{"nF charset G0", "\x1b(A"},
		{"nF charset G0 then text", "\x1b(Btext"},

		// DCS with params
		{"C1 DCS with params and C1 ST", "\x90?123;456+q\x9c"},

		// APC payload (Kitty graphics protocol)
		{"APC kitty graphics", "\x1b_Gf=24,s=10,v=20,o=z;aGVsbG8gd29ybGQ=\x1b\\"},

		// C1 CSI with multiple params
		{"C1 CSI multiple params", "\x9B1;2;3m"},

		// Mixed 7-bit and C1
		{"mixed 7-bit and C1", "\x1b[1m\x9B31mhello\x1b[0m"},

		// Concatenated sequences
		{"concatenated CSI+OSC", "\x1b[1;2;3m\x1b]2;Terminal\x07"},
		{"OSC then CSI", "\x1b]0;Title\x07\x1b[31mred\x1b[0m"},

		// Text around sequences
		{"text around SGR", "hello, \x1b[1;2;3mworld\x1b[0m!"},

		// Realistic colored output
		{"colored ls", "\x1b[1;34mDocuments\x1b[0m  \x1b[0;32mbuild.sh\x1b[0m"},

		// Plain text (no ANSI)
		{"plain ASCII", "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ours := uax29Tokens(tt.input)
			theirs := charmTokens(tt.input)
			if !reflect.DeepEqual(ours, theirs) {
				t.Errorf("boundary mismatch\nours:   %q\ntheirs: %q", ours, theirs)
			}
		})
	}
}

// ansiSample builds a realistic ANSI-heavy string simulating colored terminal output.
func ansiSample() string {
	var b strings.Builder
	colors := []string{
		"\x1b[1;34m", // bold blue
		"\x1b[0;32m", // green
		"\x1b[0;36m", // cyan
		"\x1b[1;31m", // bold red
		"\x1b[33m",   // yellow
	}
	reset := "\x1b[0m"
	lines := []string{
		"drwxr-xr-x  5 user staff  160 Jan  1 12:00 Documents",
		"drwxr-xr-x  3 user staff   96 Feb  2 09:30 Downloads",
		"-rwxr-xr-x  1 user staff 8432 Mar 15 14:22 build.sh",
		"lrwxr-xr-x  1 user staff   11 Apr 20 08:00 config",
		"-rw-r--r--  1 user staff 1024 May  5 16:45 README.md",
	}
	for round := 0; round < 40; round++ {
		for i, line := range lines {
			color := colors[i%len(colors)]
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

func BenchmarkAnsiIteration(b *testing.B) {
	input := ansiSample()
	n := int64(len(input))

	b.Run("clipperhouse/uax29", func(b *testing.B) {
		b.SetBytes(n)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			count := 0
			g := graphemes.FromString(input)
			g.AnsiEscapeSequences = true
			for g.Next() {
				count++
			}
		}
	})

	b.Run("charmbracelet/x/ansi", func(b *testing.B) {
		b.SetBytes(n)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			count := 0
			var state byte
			remaining := input
			for len(remaining) > 0 {
				_, _, advance, newState := ansi.DecodeSequence(remaining, state, nil)
				state = newState
				remaining = remaining[advance:]
				count++
			}
		}
	})
}

// uax29Tokens segments the input using our graphemes iterator with ANSI support.
func uax29Tokens(input string) []string {
	iter := graphemes.FromString(input)
	iter.AnsiEscapeSequences = true
	iter.AnsiEscapeSequences8Bit = true
	var tokens []string
	for iter.Next() {
		tokens = append(tokens, iter.Value())
	}
	return tokens
}

// charmTokens segments the input using charmbracelet/x/ansi's DecodeSequence.
func charmTokens(input string) []string {
	var state byte
	remaining := input
	var tokens []string
	for len(remaining) > 0 {
		seq, _, n, newState := ansi.DecodeSequence(remaining, state, nil)
		tokens = append(tokens, seq)
		state = newState
		remaining = remaining[n:]
	}
	return tokens
}
