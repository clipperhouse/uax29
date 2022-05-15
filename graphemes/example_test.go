package graphemes_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/clipperhouse/uax29/graphemes"
)

func ExampleSegmenter_Next() {
	text := []byte("This is an example.")

	segments := graphemes.NewSegmenter(text)

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
	text := []byte("This is an example.")

	segments := graphemes.SegmentAll(text)
	fmt.Printf("%q\n", segments)
}

func ExampleScanner_Scan() {
	text := "Good dog! 👍🏼🐶"
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
