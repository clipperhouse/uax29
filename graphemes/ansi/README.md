Package ansi provides functions for getting ANSI escape sequence lengths for both 7-bit and 8-bit C1 sequences.

[![Documentation](https://pkg.go.dev/badge/github.com/clipperhouse/uax29/v2/ansi.svg)](https://pkg.go.dev/github.com/clipperhouse/uax29/v2/ansi)
![Tests](https://github.com/clipperhouse/uax29/actions/workflows/gotest.yml/badge.svg)
![Fuzz](https://github.com/clipperhouse/uax29/actions/workflows/gofuzz.yml/badge.svg)

## Quick start

```
go get github.com/clipperhouse/uax29/v2/ansi
```

```go
import "github.com/clipperhouse/uax29/v2/ansi"

text := "\x1b[31mHello\x1b[0m"
escLen := ansi.EscapeLength(text)

fmt.Printf("length of ANSI escape sequence is: %d\n", escLen) // Prints out "length of ANSI escape sequence is: 5"
```

## Conformance

We use the Unicode [test suite](https://unicode.org/reports/tr41/tr41-36.html#Tests29).

![Tests](https://github.com/clipperhouse/uax29/actions/workflows/gotest.yml/badge.svg)
![Fuzz](https://github.com/clipperhouse/uax29/actions/workflows/gofuzz.yml/badge.svg)

## APIs

### If you have a 7-bit ANSI escape sequence

```go
text := "\x1b[31mHello\x1b[0m"
escLen := ansi.EscapeLength(text)

fmt.Printf("length of first 7-bit ANSI escape sequence is: %d\n", escLen) // Prints out "length of first 7-bit ANSI escape sequence is: 5"
```

### If you have an 8-bit ANSI escape sequence

0x9B
```go
text := "\x9B31mHello"
escLen := ansi.EscapeLength8Bit(text)

fmt.Printf("length of first 8-bit ANSI escape sequence is: %d\n", escLen) // Prints out "length of first 8-bit ANSI escape sequence is: 4"
```

### ANSI escape sequences

For ESC-initiated (7-bit) control strings, only 7-bit terminators are recognized.
For C1-initiated (8-bit) control strings, only C1 ST (`0x9C`) is recognized as ST.

We implement [ECMA-48](https://ecma-international.org/publications-and-standards/standards/ecma-48/) control codes in both 7-bit and 8-bit representations. 8-bit control codes are not UTF-8 encoded and are not valid UTF-8, caveat emptor.
