//go:build go1.23
// +build go1.23

package iterators_test

import (
	"errors"
	"os"
	"testing"

	"github.com/clipperhouse/uax29/iterators"
)

func TestScannerIterErr(t *testing.T) {
	file1, err := os.Open("../testdata/sample.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file1.Close()

	e := "hello error"
	split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		return 1, nil, errors.New(e)
	}

	sc := iterators.NewScanner(file1, split)

	for _, err := range sc.All() {
		if err == nil {
			t.Fatal("iter should have returned an error")
		}
		if err.Error() != e {
			t.Fatalf("iter should have returned %q, got %q", e, err)
		}
	}
}
