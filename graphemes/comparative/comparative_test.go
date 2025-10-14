package comparative

import (
	"testing"

	"github.com/clipperhouse/uax29/v2/graphemes"
	"github.com/clipperhouse/uax29/v2/testdata"
	"github.com/rivo/uniseg"
)

func BenchmarkUAX29Graphemes(b *testing.B) {
	data, err := testdata.Sample()
	if err != nil {
		b.Fatal(err)
	}
	text := string(data)

	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count := 0
		tokens := graphemes.FromString(text)
		for tokens.Next() {
			count++
		}
	}
}

func BenchmarkUnisegGraphemes(b *testing.B) {
	data, err := testdata.Sample()
	if err != nil {
		b.Fatal(err)
	}
	text := string(data)

	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count := 0
		gr := uniseg.NewGraphemes(text)
		for gr.Next() {
			count++
		}
	}
}

// Test that both implementations produce the same number of graphemes
func TestGraphemeCountConsistency(t *testing.T) {
	data, err := testdata.Sample()
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)

	// Count with UAX29
	uax29Count := 0
	tokens := graphemes.FromString(text)
	for tokens.Next() {
		uax29Count++
	}

	// Count with uniseg
	unisegCount := 0
	gr := uniseg.NewGraphemes(text)
	for gr.Next() {
		unisegCount++
	}

	if uax29Count != unisegCount {
		t.Errorf("Grapheme count mismatch: UAX29=%d, uniseg=%d", uax29Count, unisegCount)
	}

	t.Logf("Both implementations found %d graphemes", uax29Count)
}
