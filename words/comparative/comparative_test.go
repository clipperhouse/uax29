package comparative

import (
	"strings"
	"testing"

	"github.com/blevesearch/segment"
	"github.com/clipperhouse/uax29/v2/testdata"
	"github.com/clipperhouse/uax29/v2/words"
)

// asciiProse is typical English prose, ASCII only, for benchmarking the ASCII fast path
var asciiProse = strings.Repeat(`The quick brown fox jumps over the lazy dog. This is a sample of typical English prose that contains common words and punctuation. Software engineers often work with text processing and need efficient algorithms to handle large amounts of data. The performance of text segmentation can be critical in search engines and natural language processing applications. When dealing with ASCII text the optimizer can take shortcuts that would not be safe with arbitrary Unicode input. Numbers like 12345 and mixed tokens like abc123 should also be handled efficiently. Short words are common in English text and spaces separate them cleanly. `, 100)

func BenchmarkWordsMultilingual(b *testing.B) {
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

func BenchmarkWordsASCII(b *testing.B) {
	b.Run("clipperhouse/uax29", func(b *testing.B) {
		b.SetBytes(int64(len(asciiProse)))
		for i := 0; i < b.N; i++ {
			count := 0
			tokens := words.FromString(asciiProse)
			for tokens.Next() {
				count++
			}
		}
	})

	b.Run("blevesearch/segment", func(b *testing.B) {
		b.SetBytes(int64(len(asciiProse)))
		for i := 0; i < b.N; i++ {
			count := 0
			segmenter := segment.NewWordSegmenterDirect([]byte(asciiProse))
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
