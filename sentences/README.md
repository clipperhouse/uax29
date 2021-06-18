An implementation of sentence boundaries from [Unicode text segmentation](https://unicode.org/reports/tr29/#Sentence_Boundaries) (UAX 29), for Unicode version 13.0.0.

### Usage

```go
import "github.com/clipperhouse/uax29/sentences"

text := "This is a test. “Is it?”, he wondered."
reader := strings.NewReader(text)

scanner := sentences.NewScanner(reader)

// Scan returns true until error or EOF
for scanner.Scan() {
	fmt.Printf("%s\n", scanner.Text())
}

// Gotta check the error (because we depend on I/O).
if err := scanner.Err(); err != nil {
	log.Fatal(err)
}
```

[GoDoc](https://godoc.org/github.com/clipperhouse/uax29/sentences)

### Conformance

We use the official [test suite](https://unicode.org/reports/tr41/tr41-26.html#Tests29), thanks to [bleve](https://github.com/blevesearch/segment/blob/master/tables_test.go). Status:

![Go](https://github.com/clipperhouse/uax29/workflows/Go/badge.svg)

### Performance

`uax29` is designed to work in constant memory, regardless of input size. It buffers input and streams tokens.

Execution time is `O(n)` on input size. It can be I/O bound; you can control I/O and performance implications by the `io.Reader` you pass to `NewScanner`.

### Invalid inputs

Invalid UTF-8 input is considered undefined behavior. That said, we’ve worked to ensure that such inputs will not cause pathological outcomes, such as a panic or infinite loop. Callers should expect “garbage-in, garbage-out”.

There are two basic tests in each package, called `TestInvalidUTF8` and `TestRandomBytes`. Those tests pass, returning the invalid bytes verbatim, without a guarantee as to how they will be segmented.
