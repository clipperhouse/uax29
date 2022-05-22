package iterators_test

import (
	"strings"
	"testing"

	"github.com/clipperhouse/uax29/iterators"
	"github.com/clipperhouse/uax29/iterators/filter"
	"github.com/clipperhouse/uax29/words"
)

func TestScanner(t *testing.T) {
	text := "Hello. How are you? 🐶👍"
	r := strings.NewReader(text)

	sc := iterators.NewScanner(r, words.SplitFunc)
	sc.Filters(filter.Wordlike)

	for sc.Scan() {
		t.Logf("%s\n", sc.Bytes())
	}
}
