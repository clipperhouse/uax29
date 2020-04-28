package uax29_test

import (
	"bufio"
	"reflect"
	"strings"
	"testing"

	"github.com/clipperhouse/uax29"
	"github.com/clipperhouse/uax29/words"
)

func TestSplitFunc(t *testing.T) {
	original := "This is a test."

	r1 := strings.NewReader(original)
	scanner1 := uax29.NewScanner(r1, words.BreakFunc)

	var got1 []string
	for scanner1.Scan() {
		got1 = append(got1, scanner1.Text())
	}
	if err := scanner1.Err(); err != nil {
		t.Error(err)
	}

	r2 := strings.NewReader(original)
	scanner2 := bufio.NewScanner(r2)
	scanner2.Split(words.SplitFunc)
	var got2 []string
	for scanner2.Scan() {
		got2 = append(got2, scanner2.Text())
	}
	if err := scanner2.Err(); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(got1, got2) {
		t.Errorf("got1:\n%q\ngot2:\n%q\n", got1, got2)
	}
}
