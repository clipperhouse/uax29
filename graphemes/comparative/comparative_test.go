package comparative

import (
	"strings"
	"testing"

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
