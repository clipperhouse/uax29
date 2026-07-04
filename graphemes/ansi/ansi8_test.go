package ansi

import "testing"

// Tests for [EscapeLength8Bit] function.
func TestEscapeLength8Bit(t *testing.T) {
	t.Parallel()

	tests := []ansiCase{
		{name: "empty input", input: "", expectedLen: 0},
		{name: "C1 CSI with empty body", input: "\x9B", expectedLen: 0},
		{name: "C1 CSI then text", input: "\x9B31mhello", expectedLen: 4},
		{name: "C1 CSI multiple params", input: "\x9B1;2;3m", expectedLen: 7},
		{name: "C1 OSC with C1 ST", input: "\x9D0;Title\x9C", expectedLen: 9},
		{name: "C1 OSC with 7-bit ST is not parsed as one sequence", input: "\x9D0;Title\x1b\\", expectedLen: 0},
		{name: "C1 DCS with C1 ST", input: "\x90qpayload\x9C", expectedLen: 10},
		{name: "C1 DCS with 7-bit ST is not parsed as one sequence", input: "\x90qpayload\x1b\\", expectedLen: 0},
		{name: "C1 DCS canceled by CAN", input: "\x90qpayload\x18x", expectedLen: 9},
		{name: "C1 SOS with C1 ST", input: "\x98hello\x9C", expectedLen: 7},
		{name: "C1 PM with 7-bit ST is not parsed as one sequence", input: "\x9Emsg\x1b\\", expectedLen: 0},
		{name: "C1 APC with C1 ST", input: "\x9Fdata\x9C", expectedLen: 6},
		{name: "single C1 Fe control", input: "\x84", expectedLen: 1},
		{name: "C1 OSC unterminated", input: "\x9D0;title", expectedLen: 0},
		{name: "C1 DCS unterminated", input: "\x90data", expectedLen: 0},
		{name: "7-bit ESC sequence is not parsed", input: "\x1b[31mhello", expectedLen: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			returnedLen := EscapeLength8Bit(tt.input)
			if returnedLen != tt.expectedLen {
				t.Fatalf("EscapeLength8Bit returned %d, expected %d", returnedLen, tt.expectedLen)
			}
		})
	}
}

// Tests for [oscLengthC1] function.
func TestOscLengthC1(t *testing.T) {
	t.Parallel()

	tests := []ansiCase{
		{name: "OSC empty input", input: "\x9D", expectedLen: -1},
		{name: "OSC with BEL terminator", input: "\x9D0;Title\x07", expectedLen: 8},
		{name: "OSC with C1 ST terminator", input: "\x9D0;Title\x9C", expectedLen: 8},
		{name: "OSC with cancel", input: "\x9D0;Title\x18", expectedLen: 7},
		{name: "OSC unterminated", input: "\x9D0;Title", expectedLen: -1},
		{name: "OSC empty with cancel", input: "\x9D\x18", expectedLen: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			returnedLen := oscLengthC1(tt.input[1:])
			if returnedLen != tt.expectedLen {
				t.Fatalf("oscLengthC1 returned %d, expected %d", returnedLen, tt.expectedLen)
			}
		})
	}
}

// Tests for [stSequenceLengthC1] function.
func TestStSequenceLengthC1(t *testing.T) {
	t.Parallel()

	tests := []ansiCase{
		{name: "DCS with C1 ST terminator", input: "\x90qpayload\x9C", expectedLen: 9},
		{name: "DCS canceled by CAN", input: "\x90qpayload\x18x", expectedLen: 8},
		{name: "SOS with C1 ST terminator", input: "\x98hello\x9C", expectedLen: 6},
		{name: "PM with C1 ST terminator", input: "\x9Emsg\x9C", expectedLen: 4},
		{name: "APC with C1 ST terminator", input: "\x9Fdata\x9C", expectedLen: 5},
		{name: "unterminated DCS", input: "\x90qpayload", expectedLen: -1},
		{name: "unterminated SOS", input: "\x98hello", expectedLen: -1},
		{name: "unterminated PM", input: "\x9Emsg", expectedLen: -1},
		{name: "unterminated APC", input: "\x9Fdata", expectedLen: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			returnedLen := stSequenceLengthC1(tt.input[1:])
			if returnedLen != tt.expectedLen {
				t.Fatalf("stSequenceLengthC1 returned %d, expected %d", returnedLen, tt.expectedLen)
			}
		})
	}
}
