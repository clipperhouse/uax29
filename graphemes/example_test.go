package graphemes_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/clipperhouse/uax29/graphemes"
)

func ExampleNewSegmenter() {
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
	// Output: "H"
	// "e"
	// "l"
	// "l"
	// "o"
	// ","
	// " "
	// "世"
	// "界"
	// "."
	// " "
	// "N"
	// "i"
	// "c"
	// "e"
	// " "
	// "d"
	// "o"
	// "g"
	// "!"
	// " "
	// "👍"
	// "🐶"
}

func ExampleSegmentAll() {
	text := []byte("Hello, 世界. Nice dog! 👍🐶")

	segments := graphemes.SegmentAll(text)
	fmt.Printf("%q\n", segments)

	// Output: ["H" "e" "l" "l" "o" "," " " "世" "界" "." " " "N" "i" "c" "e" " " "d" "o" "g" "!" " " "👍" "🐶"]
}

func ExampleNewScanner() {
	text := "Hello, 世界. Nice dog! 👍🐶"
	reader := strings.NewReader(text)

	scanner := graphemes.NewScanner(reader)

	// Scan returns true until error or EOF
	for scanner.Scan() {
		fmt.Printf("%q\n", scanner.Text())
	}

	// Gotta check the error!
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	// Output: "H"
	// "e"
	// "l"
	// "l"
	// "o"
	// ","
	// " "
	// "世"
	// "界"
	// "."
	// " "
	// "N"
	// "i"
	// "c"
	// "e"
	// " "
	// "d"
	// "o"
	// "g"
	// "!"
	// " "
	// "👍"
	// "🐶"
}
