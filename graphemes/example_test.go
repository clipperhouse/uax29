package graphemes_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/clipperhouse/uax29/graphemes"
)

func ExampleSegmenter_Next() {
	text := []byte("Hello, 世界. Nice dog! 👍🐶")

	segments := graphemes.NewSegmenter(text)

	// Next() returns true until end of data or error
	for segments.Next() {
		fmt.Printf("%q\n", segments.Bytes())
	}

	// Should check the error
	if err := segments.Err(); err != nil {
		log.Fatal(err)
	}
}

func ExampleSegmentAll() {
	text := []byte("Hello, 世界. Nice dog! 👍🐶")

	segments := graphemes.SegmentAll(text)
	fmt.Printf("%q\n", segments)
}

func ExampleScanner_Scan() {
	text := "Hello, 世界. Nice dog! 👍🐶"
	reader := strings.NewReader(text)

	scanner := graphemes.NewScanner(reader)

	// Scan returns true until error or EOF
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	// Gotta check the error!
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
