package words_test

import "github.com/clipperhouse/uax29/v2/words"

var joinersInput = []byte("Hello, 世界. Tell me about your super-cool .com. I'm .01% interested and 3/4 of a mile away. Email me at foo@example.biz. #winning")
var joiners = &words.Joiners{
	Middle:  []rune("@-/"),
	Leading: []rune("#."),
}

type joinersTest struct {
	input string
	// word should be found in standard iterator
	found1 bool
	// word should be found in iterator with joiners
	found2 bool
}

var joinersTests = []joinersTest{
	{"Hello", true, true},
	{"世", true, true},
	{"super", true, false},
	{"-", true, false},
	{"cool", true, false},
	{"super-cool", false, true},
	{"com", true, false}, // ".com" should usually be split, but joined with config
	{".com", false, true},
	{"01", true, false},
	{".01", false, true},
	{"3", true, false},
	{"3/4", false, true},
	{"foo", true, false},
	{"@", true, false},
	{"example.biz", true, false},
	{"foo@example.biz", false, true},
	{"#", true, false},
	{"winning", true, false},
	{"#winning", false, true},
}
