package sentences_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/clipperhouse/uax29/sentences"
)

func ExampleSegmenter_Next() {
	text := []byte("Hello, 世界. “Nice dog! 👍🐶”, they said.")

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
	text := []byte("Hello, 世界. “Nice dog! 👍🐶”, they said.")

	segments := sentences.SegmentAll(text)
	fmt.Printf("%q\n", segments)
}

func ExampleScanner_Scan() {
	text := "Hello, 世界. “Nice dog! 👍🐶”, they said."
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
