package phrases_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/clipperhouse/uax29/iterators/filter"
	"github.com/clipperhouse/uax29/phrases"
)

func ExampleNewScanner() {
	text := "Hello, 世界. Nice dog! 👍🐶"
	r := strings.NewReader(text)

	sc := phrases.NewScanner(r)
	sc.Filter(filter.Wordlike) // let's exclude whitespace & punctuation

	// Scan returns true until error or EOF
	for sc.Scan() {
		// Do something with the token (segment)
		fmt.Printf("%q\n", sc.Text())
	}

	// Gotta check the error!
	if err := sc.Err(); err != nil {
		log.Fatal(err)
	}
	// Output: "Hello"
	//"世"
	//"界"
	//"Nice"
	//"dog"
	//"👍"
	//"🐶"
}

func ExampleNewSegmenter() {
	text := []byte("Hello, 世界. Nice dog! 👍🐶")

	seg := phrases.NewSegmenter(text)
	seg.Filter(filter.Wordlike) // let's exclude whitespace & punctuation

	// Next returns true until error or end of data
	for seg.Next() {
		// Do something with the token (segment)
		fmt.Printf("%q\n", seg.Bytes())
	}

	// Gotta check the error!
	if err := seg.Err(); err != nil {
		log.Fatal(err)
	}
	// Output: "Hello"
	//"世"
	//"界"
	//"Nice"
	//"dog"
	//"👍"
	//"🐶"
}

func ExampleSegmentAll() {
	text := []byte("Hello, 世界. Nice dog! 👍🐶")

	segments := phrases.SegmentAll(text)
	fmt.Printf("%q\n", segments)
	// Output: ["Hello" "," " " "世" "界" "." " " "Nice" " " "dog" "!" " " "👍" "🐶"]
}
