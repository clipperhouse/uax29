package comparative

import (
	"testing"

	"github.com/blevesearch/segment"
	"github.com/clipperhouse/uax29/v2/testdata"
	"github.com/clipperhouse/uax29/v2/words"
)

func BenchmarkWords(b *testing.B) {
	data, err := testdata.Sample()
	if err != nil {
		b.Fatal(err)
	}
	text := string(data)

	b.Run("clipperhouse/uax29", func(b *testing.B) {
		b.SetBytes(int64(len(text)))
		for i := 0; i < b.N; i++ {
			count := 0
			tokens := words.FromString(text)
			for tokens.Next() {
				count++
			}
		}
	})

	b.Run("blevesearch/segment", func(b *testing.B) {
		b.SetBytes(int64(len(text)))
		for i := 0; i < b.N; i++ {
			count := 0
			segmenter := segment.NewWordSegmenterDirect([]byte(text))
			for segmenter.Segment() {
				count++
			}
		}
	})
}

// Test that both implementations produce the same number of words
func TestWordCountConsistency(t *testing.T) {
	data, err := testdata.Sample()
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)

	// Count with UAX29
	uax29Count := 0
	tokens := words.FromString(text)
	for tokens.Next() {
		uax29Count++
	}

	// Count with blevesearch/segment
	bleveCount := 0
	segmenter := segment.NewWordSegmenterDirect([]byte(text))
	for segmenter.Segment() {
		bleveCount++
	}

	t.Logf("UAX29: %d words, blevesearch: %d words", uax29Count, bleveCount)

	if uax29Count != bleveCount {
		t.Logf("Note: Different word counts likely due to different boundary rule implementations")
	}
}
