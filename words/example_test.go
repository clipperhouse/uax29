package words_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/clipperhouse/uax29/words"
)

func ExampleScanner_Scan() {
	text := "Hello, 世界. Nice dog! 👍🐶"
	reader := strings.NewReader(text)

	scanner := words.NewScanner(reader)

	// Scan returns true until error or EOF
	for scanner.Scan() {
		fmt.Printf("%q\n", scanner.Text())
	}

	// Gotta check the error!
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func ExampleSegmenter_Next() {
	text := []byte("Hello, 世界. Nice dog! 👍🐶")

	segments := words.NewSegmenter(text)

	// Next returns true until error or end of data
	for segments.Next() {
		fmt.Printf("%q\n", segments.Bytes())
	}

	// Gotta check the error!
	if err := segments.Err(); err != nil {
		log.Fatal(err)
	}
}

func ExampleSegmentAll() {
	text := []byte("Hello, 世界. Nice dog! 👍🐶")

	segments := words.SegmentAll(text)
	fmt.Printf("%q\n", segments)
}
