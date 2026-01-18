package comparative

import (
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
