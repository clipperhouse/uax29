package words_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/clipperhouse/uax29/words"
)

func ExampleScanner_Scan() {
	text := "This is an example."
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
