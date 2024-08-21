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

	sc := words.NewScanner(r)
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

	seg := words.NewSegmenter(text)
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

	segments := words.SegmentAll(text)
	fmt.Printf("%q\n", segments)
	// Output: ["Hello" "," " " "世" "界" "." " " "Nice" " " "dog" "!" " " "👍" "🐶"]
}

// In the example below, the hyphen, the leading dot on .com, the leading decimal, the slash on the fraction, the email address
// and the hashtag, would be split into two tokens by default, but are joined into single tokens using joiners.
func ExampleJoiners() {
	text := "Hello, 世界. Tell me about your super-cool .com. I'm .01% interested and 3/4 of a mile away. Email me at foo@example.biz. #winning"
	joiners := &words.Joiners{
		Mid:     []rune("@-/"), // appearing in the middle of a word
		Leading: []rune("#."),  // appearing at the front of a word
	}

	seg := words.NewSegmenter([]byte(text))
	seg.Joiners(joiners)
	seg.Filter(filter.Wordlike)

	for seg.Next() {
		fmt.Println(seg.Text())
	}
	// Output: Hello
	// 世
	// 界
	// Tell
	// me
	// about
	// your
	// super-cool
	// .com
	// I'm
	// .01
	// interested
	// and
	// 3/4
	// of
	// a
	// mile
	// away
	// Email
	// me
	// at
	// foo@example.biz
	// #winning
}
