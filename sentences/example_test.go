package sentences_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/clipperhouse/uax29/sentences"
)

func ExampleSegmenter_Next() {
	text := []byte("This is an example.")

	segments := sentences.NewSegmenter(text)

	// Scan returns true until error or EOF
	for segments.Next() {
		fmt.Printf("%q\n", segments.Bytes())
	}

	// Gotta check the error!
	if err := segments.Err(); err != nil {
		log.Fatal(err)
	}
}

func ExampleSegmentAll() {
	text := []byte("This is an example. Followed by a second sentence.")

	segments := sentences.SegmentAll(text)
	fmt.Printf("%q\n", segments)
}

func ExampleScanner_Scan() {
	text := "This is a test. “Is it?”, he wondered."
	reader := strings.NewReader(text)

	scanner := sentences.NewScanner(reader)

	// Scan returns true until error or EOF
	for scanner.Scan() {
		fmt.Printf("%s\n", scanner.Text())
	}

	// Gotta check the error!
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
