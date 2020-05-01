This package tokenizes text based on [Unicode text segmentation](https://unicode.org/reports/tr29/#Word_Boundaries) (UAX 29), for Unicode version 12.0.

### Usage

```go
import "github.com/clipperhouse/uax29/words"

text := "It’s not “obvious” (IMHO) what comprises a word, a sentence, or a grapheme. 👍🏼🐶!"
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

### Why tokenize?

Any time our code operates on individual words, we are tokenizing. Often, we do it ad hoc, such as splitting on spaces, which gives inconsistent results. Best to do it consistently.

### Performance

`uax29` is designed to work in constant memory, regardless of input size. It buffers input and streams tokens.

Execution time is designed to be `O(n)` on input size. It is I/O-bound. In your code, you control I/O and performance implications by the `io.Reader` you pass to `NewScanner`.

If I am reading my benchmarks correctly, `uax29/words` processes around 2MM tokens per second on my laptop, when the input is preloaded into memory. By default, it buffers 64K of input at a time.

### Conformance

We use the official [test suites](https://unicode.org/reports/tr41/tr41-26.html#Tests29), thanks to [bleve](https://github.com/blevesearch/segment/blob/master/tables_test.go). Status:

![Go](https://github.com/clipperhouse/uax29/workflows/Go/badge.svg)

### Status

- The [word boundary rules](https://unicode.org/reports/tr29/#Word_Boundaries) have been implemented in the `words` package

- The [sentence boundary rules](https://unicode.org/reports/tr29/#Sentence_Boundaries) have been implemented in the `sentences` package

- The [grapheme cluster rules](https://unicode.org/reports/tr29/#Grapheme_Cluster_Boundaries) have been implemented in the `graphemes` package

- The official [test suite](https://unicode.org/reports/tr41/tr41-26.html#Tests29) passes for words, sentences, and graphemes

- We code-gen the Unicode categories relevant to UAX 29 by running `go generate` at the repository root

- There is [discussion](https://groups.google.com/d/msg/golang-nuts/_79vJ65KuXc/B_QgeU6rAgAJ) of implementing the above in Go’s [`x/text`](https://godoc.org/golang.org/x/text) package

### See also

[jargon](https://github.com/clipperhouse/jargon), a text pipelines package for CLI and Go, which consumes this package.
