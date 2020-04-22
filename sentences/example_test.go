package sentences_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/clipperhouse/uax29/sentences"
)

func ExampleScanner_Scan() {
	text := "This is a test. “Is it?”, he wondered."
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
