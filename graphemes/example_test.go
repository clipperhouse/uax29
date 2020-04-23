package graphemes_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/clipperhouse/uax29/graphemes"
)

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
