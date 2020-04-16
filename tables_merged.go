package uax29

import "golang.org/x/text/unicode/rangetable"

var AHLetter = rangetable.Merge(ALetter, Hebrew_Letter)

var MidNumLetQ = rangetable.Merge(MidNumLet, rangetable.New('\''))
