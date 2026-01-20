package comparative

import (
	"strings"
	"testing"

	"github.com/clipperhouse/uax29/v2/graphemes"
	"github.com/clipperhouse/uax29/v2/testdata"
	"github.com/rivo/uniseg"
)

func BenchmarkGraphemes(b *testing.B) {
	data, err := testdata.Sample()
	if err != nil {
		b.Fatal(err)
	}
	text := string(data)

	b.Run("clipperhouse/uax29", func(b *testing.B) {
		b.SetBytes(int64(len(text)))
		for i := 0; i < b.N; i++ {
			count := 0
			tokens := graphemes.FromString(text)
			for tokens.Next() {
				count++
			}
		}
	})

	b.Run("rivo/uniseg", func(b *testing.B) {
		b.SetBytes(int64(len(text)))
		for i := 0; i < b.N; i++ {
			count := 0
			gr := uniseg.NewGraphemes(text)
			for gr.Next() {
				count++
			}
		}
	})
}

func BenchmarkGraphemesASCII(b *testing.B) {
	// Pure ASCII text - should benefit from ASCII hot path
	ascii := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 100)

	b.Run("clipperhouse/uax29", func(b *testing.B) {
		b.SetBytes(int64(len(ascii)))
		for i := 0; i < b.N; i++ {
			count := 0
			tokens := graphemes.FromString(ascii)
			for tokens.Next() {
				count++
			}
		}
	})

	b.Run("rivo/uniseg", func(b *testing.B) {
		b.SetBytes(int64(len(ascii)))
		for i := 0; i < b.N; i++ {
			count := 0
			gr := uniseg.NewGraphemes(ascii)
			for gr.Next() {
				count++
			}
		}
	})
}

func BenchmarkGraphemesBytes(b *testing.B) {
	data, err := testdata.Sample()
	if err != nil {
		b.Fatal(err)
	}

	b.Run("clipperhouse/uax29", func(b *testing.B) {
		b.SetBytes(int64(len(data)))
		for i := 0; i < b.N; i++ {
			count := 0
			tokens := graphemes.FromBytes(data)
			for tokens.Next() {
				count++
			}
		}
	})
}

func BenchmarkGraphemesBytesASCII(b *testing.B) {
	// Pure ASCII text - should benefit from ASCII hot path
	ascii := []byte(strings.Repeat("The quick brown fox jumps over the lazy dog. ", 100))

	b.Run("clipperhouse/uax29", func(b *testing.B) {
		b.SetBytes(int64(len(ascii)))
		for i := 0; i < b.N; i++ {
			count := 0
			tokens := graphemes.FromBytes(ascii)
			for tokens.Next() {
				count++
			}
		}
	})
}
