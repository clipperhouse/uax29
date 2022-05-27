package words_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/clipperhouse/uax29/iterators/filter"
	"github.com/clipperhouse/uax29/words"
)

func ExampleNewScanner() {
	text := "Hello, 世界. Nice dog! 👍🐶"
	r := strings.NewReader(text)

	scanner := words.NewScanner(r)

	// Scan returns true until error or EOF
	for scanner.Scan() {
		// Do something with the token (segment)
		fmt.Printf("%q\n", scanner.Text())
	}

	// Gotta check the error!
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// You can also choose to filter the returned tokens (segments)
	r2 := strings.NewReader(text)
	filteredScanner := words.NewScanner(r2)
	filteredScanner.Filter(filter.Wordlike)

	// You'll notice that whitespace and punctuation are omitted
	for filteredScanner.Scan() {
		fmt.Printf("%q\n", filteredScanner.Bytes())
	}
	if err := filteredScanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func ExampleNewSegmenter() {
	text := []byte("Hello, 世界. Nice dog! 👍🐶")

	seg := words.NewSegmenter(text)

	// Next returns true until error or end of data
	for seg.Next() {
		// Do something with the token (segment)
		fmt.Printf("%q\n", seg.Bytes())
	}

	// Gotta check the error!
	if err := seg.Err(); err != nil {
		log.Fatal(err)
	}

	// You can also choose to filter the returned tokens (segments)
	filteredSeg := words.NewSegmenter(text)
	filteredSeg.Filter(filter.Wordlike)

	// Notice that whitespace and punctuation are omitted
	for filteredSeg.Next() {
		fmt.Printf("%q\n", filteredSeg.Bytes())
	}
	if err := filteredSeg.Err(); err != nil {
		log.Fatal(err)
	}
}

func ExampleSegmentAll() {
	text := []byte("Hello, 世界. Nice dog! 👍🐶")

	segments := words.SegmentAll(text)
	fmt.Printf("%q\n", segments)
}
