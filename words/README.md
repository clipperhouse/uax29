An implementation of word boundaries from [Unicode text segmentation](https://unicode.org/reports/tr29/#Word_Boundaries) (UAX 29), for Unicode version 12.0.

### Usage

```go
import "github.com/clipperhouse/uax29/words"

text := "This is an example."
reader := strings.NewReader(text)

scanner := words.NewScanner(reader)

// Scan returns true until error or EOF
for scanner.Scan() {
	fmt.Printf("%q\n", scanner.Text())
}

// Gotta check the error (because we depend on I/O).
if err := scanner.Err(); err != nil {
	log.Fatal(err)
}
```

[GoDoc](https://godoc.org/github.com/clipperhouse/uax29/words)

### Performance

`uax29` is designed to work in constant memory, regardless of input size. It buffers input and streams tokens.

Execution time is designed to be `O(n)` on input size. It is I/O-bound. In your code, you control I/O and performance implications by the `Reader` you pass to `words.NewScanner`.

If I am reading my benchmarks correctly, `uax29/words` processes around 2MM tokens per second on my laptop, when the input is preloaded into memory. By default, it buffers 64K of input at a time.

### Conformance

We use the official [test suite](https://unicode.org/reports/tr41/tr41-26.html#Tests29), thanks to [bleve](https://github.com/blevesearch/segment/blob/master/tables_test.go). Status:

![Go](https://github.com/clipperhouse/uax29/workflows/Go/badge.svg)

The [spec](https://unicode.org/reports/tr29/#Word_Boundaries) has many nods to practicality and judgment for the implementer. One place where we vary from the strict spec is to consider underscore `_` a valid mid-word/mid-number character, helpful for things like user_names.