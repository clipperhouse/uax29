package phrases_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/clipperhouse/uax29/phrases"
)

func ExampleNewScanner() {
	text := "Hello, 世界. Nice — and adorable — dog; perhaps the “best one”! 🏆 🐶"
	r := strings.NewReader(text)

	sc := phrases.NewScanner(r)

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
	// ","
	// " "
	// "世"
	// "界"
	// "."
	// " Nice "
	// "—"
	// " and adorable "
	// "—"
	// " dog"
	// ";"
	// " perhaps the "
	// "“"
	// "best one"
	// "”"
	// "!"
	// " 🏆 🐶"
}

func ExampleNewSegmenter() {
	text := []byte("Hello, 世界. Nice — and adorable — dog; perhaps the “best one”! 🏆 🐶")

	phrase := phrases.NewSegmenter(text)

	// Next returns true until error or end of data
	for phrase.Next() {
		// Do something with the phrase
		fmt.Printf("%q\n", phrase.Bytes())
	}

	// Gotta check the error!
	if err := phrase.Err(); err != nil {
		log.Fatal(err)
	}
	// Output: "Hello"
	// ","
	// " "
	// "世"
	// "界"
	// "."
	// " Nice "
	// "—"
	// " and adorable "
	// "—"
	// " dog"
	// ";"
	// " perhaps the "
	// "“"
	// "best one"
	// "”"
	// "!"
	// " 🏆 🐶"
}

func ExampleSegmentAll() {
	text := []byte("Hello, 世界. Nice — and adorable — dog; perhaps the best one! 👍🐶")

	segments := phrases.SegmentAll(text)
	fmt.Printf("%q\n", segments)
	// Output: ["Hello" "," " " "世" "界" "." " Nice " "—" " and adorable " "—" " dog" ";" " perhaps the best one" "!" " 👍🐶"]
}
