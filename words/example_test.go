package words_test

import (
	"fmt"

	"github.com/clipperhouse/uax29/words"
)

func ExampleSegmentAll() {
	text := []byte("Hello, 世界. Nice dog! 👍🐶")

	segments := words.SegmentAll(text)
	fmt.Printf("%q\n", segments)
	// Output: ["Hello" "," " " "世" "界" "." " " "Nice" " " "dog" "!" " " "👍" "🐶"]
}
