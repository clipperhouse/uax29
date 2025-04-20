package words

import (
	"testing"
)

func TestSubsequent(t *testing.T) {
	tests := []struct {
		name       string
		properties property
		data       []byte
		atEOF      bool
		want       int
	}{
		{
			name:       "find immediate",
			properties: _Numeric,
			data:       []byte("1a"),
			atEOF:      true,
			want:       1, // width of '1'
		},
		{
			name:       "find after ignore",
			properties: _Numeric,
			data:       []byte("\u200d1a"), // ZWJ + '1' + 'a'
			atEOF:      true,
			want:       3 + 1, // width of ZWJ + width of '1'
		},
		{
			name:       "not found",
			properties: _Numeric,
			data:       []byte("abc"),
			atEOF:      true,
			want:       notfound,
		},
		{
			name:       "not found after ignore",
			properties: _Numeric,
			data:       []byte("\u200dabc"), // ZWJ + 'a' + 'b' + 'c'
			atEOF:      true,
			want:       notfound,
		},
		{
			name:       "empty data at EOF",
			properties: _Numeric,
			data:       []byte(""),
			atEOF:      true,
			want:       notfound,
		},
		{
			name:       "empty data not at EOF",
			properties: _Numeric,
			data:       []byte(""),
			atEOF:      false,
			want:       more,
		},
		{
			name:       "need more data",
			properties: _Numeric,
			data:       []byte("a"), // Need more data to determine if 'a' is followed by Numeric
			atEOF:      false,
			want:       more,
		},
		{
			name:       "need more data after ignore",
			properties: _Numeric,
			data:       []byte("\u200d"), // ZWJ
			atEOF:      false,
			want:       more,
		},
		{
			name:       "partial rune at end, atEOF",
			properties: _Numeric,
			data:       []byte{0xE2, 0x80}, // Incomplete ZWJ
			atEOF:      true,
			want:       notfound, // Cannot decode, treat as not found
		},
		{
			name:       "partial rune at end, not atEOF",
			properties: _Numeric,
			data:       []byte{0xE2, 0x80}, // Incomplete ZWJ
			atEOF:      false,
			want:       more, // Request more data
		},
		{
			name:       "find after partial rune, not atEOF",
			properties: _Numeric,
			data:       []byte{0xE2, 0x80, '1'}, // Incomplete ZWJ followed by '1'
			atEOF:      false,
			want:       more, // Request more data to complete the rune first
		},
		{
			name:       "find property that is multiple bytes",
			properties: _HebrewLetter,
			data:       []byte("א"), // Aleph
			atEOF:      true,
			want:       2, // width of Aleph
		},
		{
			name:       "find after ignore, property is multiple bytes",
			properties: _HebrewLetter,
			data:       []byte("\u200dא"), // ZWJ + Aleph
			atEOF:      true,
			want:       3 + 2, // width of ZWJ + width of Aleph
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := subsequent(tt.properties, tt.data, tt.atEOF); got != tt.want {
				t.Errorf("subsequent() = %v, want %v", got, tt.want)
			}
		})
	}
}
