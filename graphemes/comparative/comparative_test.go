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
		{"C1 DCS with C1 ST", "\x90qpayload\x9C"},
		{"C1 SOS with C1 ST", "\x98hello\x9C"},
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

		// DecodeSequence parser parity edge cases
		{"single ESC byte", "\x1b"},
		{"single NUL byte", "\x00"},
		{"ASCII DEL byte", "\x7f"},
		{"DEL between ASCII runes", "a\x7fb"},
		{"double ESC", "\x1b\x1b"},
		{"double ST 7-bit", "\x1b\\\x1b\\"},
		{"double ST 8-bit", "\x9c\x9c"},
		{"single-param OSC", "\x1b]112\x07"},
		{"ESC with intermediate", "\x1b Q"},
		{"DCS containing DEL payload", "\x1bP1;2+xa\x7fb\x1b\\"},
		{"OSC with C1 bytes in payload", "\x1b]11;\x90?\x1b\\"},
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

// TestAnsiBoundaryKnownDivergences documents cases where our grapheme-oriented
// tokenizer intentionally differs from charmbracelet/x/ansi DecodeSequence.
func TestAnsiBoundaryKnownDivergences(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		reason string
	}{
		{
			name:   "unterminated CSI",
			input:  "\x1b[1;2;3",
			reason: "DecodeSequence returns one unterminated CSI token; we split when no final byte is present",
		},
		{
			name:   "unterminated OSC",
			input:  "\x1b]11;ff/00/ff",
			reason: "DecodeSequence returns one unterminated OSC token; we split when OSC has no BEL/ST/CAN/SUB terminator",
		},
		{
			name:   "unterminated OSC followed by CSI",
			input:  "\x1b]11;ff/00/ff\x1b[1;2;3m",
			reason: "DecodeSequence ends OSC at ESC and parses following CSI; we require explicit OSC terminator",
		},
		{
			name:   "unterminated OSC followed by bare ESC",
			input:  "\x1b]11;ff/00/ff\x1b",
			reason: "DecodeSequence emits unterminated OSC then ESC; we split because OSC is invalid without terminator",
		},
		{
			name:   "unterminated DCS",
			input:  "\x1bP1;2+xa",
			reason: "DecodeSequence returns one unterminated DCS token; we split when DCS has no ST/CAN/SUB terminator",
		},
		{
			name:   "invalid DCS immediately terminated",
			input:  "\x1bP\x1b\\ab",
			reason: "DecodeSequence emits ESC P token before ST; we do not treat invalid DCS start as a sequence",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ours := uax29Tokens(tt.input)
			theirs := charmTokens(tt.input)
			if reflect.DeepEqual(ours, theirs) {
				t.Fatalf("expected divergence, but boundaries matched\nreason: %s\ntokens: %q", tt.reason, ours)
			}
			t.Logf("reason: %s", tt.reason)
			t.Logf("ours:   %q", ours)
			t.Logf("theirs: %q", theirs)
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

func ansiSample8Bit() string {
	var b strings.Builder
	lines := []string{
		"drwxr-xr-x  5 user staff  160 Jan  1 12:00 Documents",
		"drwxr-xr-x  3 user staff   96 Feb  2 09:30 Downloads",
		"-rwxr-xr-x  1 user staff 8432 Mar 15 14:22 build.sh",
		"lrwxr-xr-x  1 user staff   11 Apr 20 08:00 config",
		"-rw-r--r--  1 user staff 1024 May  5 16:45 README.md",
	}
	for round := 0; round < 40; round++ {
		for i, line := range lines {
			if i%5 == 0 {
				b.WriteByte(0x9D)
				b.WriteString("0;terminal - round ")
				b.WriteString(string(rune('0' + round%10)))
				b.WriteByte(0x07)
			}
			b.WriteByte(0x9B)
			b.WriteString("1;3")
			b.WriteString(string(rune('0' + (i % 8))))
			b.WriteByte('m')
			b.WriteString(line)
			b.WriteByte(0x9B)
			b.WriteString("0m")
			b.WriteString("\n")
		}
	}
	return b.String()
}

func ansiSampleMixed() string {
	return ansiSample() + ansiSample8Bit()
}

func BenchmarkAnsiIteration(b *testing.B) {
	input7 := ansiSample()
	input8 := ansiSample8Bit()
	inputMixed := ansiSampleMixed()

	b.Run("clipperhouse/uax29/7bit", func(b *testing.B) {
		b.SetBytes(int64(len(input7)))
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			count := 0
			g := graphemes.FromString(input7)
			g.AnsiEscapeSequences = true
			for g.Next() {
				count++
			}
		}
	})

	b.Run("clipperhouse/uax29/8bit", func(b *testing.B) {
		b.SetBytes(int64(len(input8)))
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			count := 0
			g := graphemes.FromString(input8)
			g.AnsiEscapeSequences8Bit = true
			for g.Next() {
				count++
			}
		}
	})

	b.Run("clipperhouse/uax29/both", func(b *testing.B) {
		b.SetBytes(int64(len(inputMixed)))
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			count := 0
			g := graphemes.FromString(inputMixed)
			g.AnsiEscapeSequences = true
			g.AnsiEscapeSequences8Bit = true
			for g.Next() {
				count++
			}
		}
	})

	b.Run("charmbracelet/x/ansi/mixed", func(b *testing.B) {
		b.SetBytes(int64(len(inputMixed)))
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			count := 0
			var state byte
			remaining := inputMixed
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
