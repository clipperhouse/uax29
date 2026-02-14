package graphemes

// ansiEscapeLength returns the byte length of a valid ANSI escape/control
// sequence at the start of data, or 0 if none.
//
// Input is UTF-8. This recognizes both:
//   - 7-bit representations (ESC + final/intermediate bytes), and
//   - UTF-8 encodings of 8-bit C1 controls (U+0080..U+009F => 0xC2 0x80..0x9F).
//
// Recognized forms (ECMA-48 / ISO 6429):
//   - CSI: ESC [ then parameter bytes (0x30–0x3F), intermediate (0x20–0x2F), final (0x40–0x7E)
//   - OSC: ESC ] then payload until ST (ESC \) or BEL (0x07)
//   - DCS, SOS, PM, APC: ESC P / X / ^ / _ then payload until ST (ESC \)
//   - Two-byte: ESC + Fe/Fs (0x40–0x7E excluding above), or Fp (0x30–0x3F), or nF (0x20–0x2F then final)
func ansiEscapeLength[T ~string | ~[]byte](data T) int {
	n := len(data)
	if n < 2 {
		return 0
	}

	switch data[0] {
	case esc:
		b1 := data[1]
		switch b1 {
		case '[': // CSI
			body := csiLength(data[2:])
			if body == 0 {
				return 0
			}
			return 2 + body
		case ']': // OSC – allows BEL or ST as terminator
			body := oscLength(data[2:])
			if body < 0 {
				return 0
			}
			return 2 + body
		case 'P', 'X', '^', '_': // DCS, SOS, PM, APC – require ST only
			body := stSequenceLength(data[2:])
			if body < 0 {
				return 0
			}
			return 2 + body
		}
		if b1 >= 0x40 && b1 <= 0x7E {
			// Fe/Fs two-byte; [ ] P X ^ _ handled above
			return 2
		}
		if b1 >= 0x30 && b1 <= 0x3F {
			// Fp (private) two-byte
			return 2
		}
		if b1 >= 0x20 && b1 <= 0x2F {
			// nF: intermediates then one final (0x30–0x7E)
			i := 2
			for i < n && data[i] >= 0x20 && data[i] <= 0x2F {
				i++
			}
			if i < n && data[i] >= 0x30 && data[i] <= 0x7E {
				return i + 1
			}
			return 0
		}

	case c1UTF8Lead:
		b1 := data[1]
		if b1 < 0x80 || b1 > 0x9F {
			return 0
		}

		switch b1 {
		case 0x9B: // CSI
			body := csiLength(data[2:])
			if body == 0 {
				return 0
			}
			return 2 + body
		case 0x9D: // OSC – allows BEL or ST as terminator
			body := oscLength(data[2:])
			if body < 0 {
				return 0
			}
			return 2 + body
		case 0x90, 0x98, 0x9E, 0x9F: // DCS, SOS, PM, APC – require ST only
			body := stSequenceLength(data[2:])
			if body < 0 {
				return 0
			}
			return 2 + body
		default:
			// Any other C1 control (UTF-8 encoded) is one control sequence token.
			return 2
		}
	}

	return 0
}

// csiLength returns the length of the CSI body (param/intermediate/final bytes).
// data is the slice after "ESC [".
// Per ECMA-48, the CSI body has the form:
//
//	parameters (0x30–0x3F)*, intermediates (0x20–0x2F)*, final (0x40–0x7E)
//
// Once an intermediate byte is seen, subsequent parameter bytes are invalid.
func csiLength[T ~string | ~[]byte](data T) int {
	seenIntermediate := false
	for i := 0; i < len(data); i++ {
		b := data[i]
		if b >= 0x30 && b <= 0x3F {
			if seenIntermediate {
				return 0
			}
			continue
		}
		if b >= 0x20 && b <= 0x2F {
			seenIntermediate = true
			continue
		}
		if b >= 0x40 && b <= 0x7E {
			return i + 1
		}
		return 0
	}
	return 0
}

// oscLength returns the length of the OSC body.
// data is the slice after "ESC ]" (or C1 OSC).
//
// Returns:
//   - n >= 0: consumed body length (includes BEL/ST terminator when present)
//   - -1: not terminated in the provided data
//
// OSC accepts BEL (0x07) or ST as terminator by widespread convention.
// Per ECMA-48, CAN (0x18) and SUB (0x1A) cancel the control string; in that
// case they are not part of the OSC sequence length.
func oscLength[T ~string | ~[]byte](data T) int {
	for i := 0; i < len(data); i++ {
		b := data[i]
		if b == bel {
			return i + 1
		}
		if b == can || b == sub {
			return i
		}
		if b == esc && i+1 < len(data) && data[i+1] == '\\' {
			return i + 2
		}
		if b == c1UTF8Lead && i+1 < len(data) && data[i+1] == 0x9C {
			return i + 2
		}
	}
	return -1
}

// stSequenceLength returns the length of a control-string body.
// data is the slice after "ESC x" (or C1 DCS/SOS/PM/APC).
//
// Returns:
//   - n >= 0: consumed body length (includes ST terminator when present)
//   - -1: not terminated in the provided data
//
// Used for DCS, SOS, PM, and APC, which per ECMA-48 terminate with ST.
// CAN (0x18) and SUB (0x1A) cancel the control string; in that case they are
// not part of the sequence length.
func stSequenceLength[T ~string | ~[]byte](data T) int {
	for i := 0; i < len(data); i++ {
		if data[i] == can || data[i] == sub {
			return i
		}
		if data[i] == esc && i+1 < len(data) && data[i+1] == '\\' {
			return i + 2
		}
		if data[i] == c1UTF8Lead && i+1 < len(data) && data[i+1] == 0x9C {
			return i + 2
		}
	}
	return -1
}
