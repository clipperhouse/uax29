This package tokenizes words, sentences and graphemes, based on [Unicode text segmentation](https://unicode.org/reports/tr29/#Word_Boundaries) (UAX 29), for Unicode version 12.0.

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

### Conformance

We use the official [test suites](https://unicode.org/reports/tr41/tr41-26.html#Tests29), thanks to [bleve](https://github.com/blevesearch/segment/blob/master/tables_test.go). Status:

![Go](https://github.com/clipperhouse/uax29/workflows/Go/badge.svg)

### Performance

`uax29` is designed to work in constant memory, regardless of input size. It buffers input and streams tokens. (For example, I am showing a maximum resident size of 8MB when processing a 300MB file.)

Execution time is `O(n)` on input size. It can be I/O bound; I/O performance is determined by the `io.Reader` you pass to `NewScanner`.

In my local benchmarking (Mac laptop), `uax29/words` processes around 16MM tokens per second, or 40MB/s, of [wiki text](https://en.wikipedia.org/w/index.php?title=New_York_City&action=edit).

### Status

- The [word boundary rules](https://unicode.org/reports/tr29/#Word_Boundaries) have been implemented in the `words` package

- The [sentence boundary rules](https://unicode.org/reports/tr29/#Sentence_Boundaries) have been implemented in the `sentences` package

- The [grapheme cluster rules](https://unicode.org/reports/tr29/#Grapheme_Cluster_Boundaries) have been implemented in the `graphemes` package

- The official [test suite](https://unicode.org/reports/tr41/tr41-26.html#Tests29) passes for words, sentences, and graphemes

- We code-gen the Unicode categories relevant to UAX 29 by running `go generate` at the repository root

- There is [discussion](https://groups.google.com/d/msg/golang-nuts/_79vJ65KuXc/B_QgeU6rAgAJ) of implementing the above in Go’s [`x/text`](https://godoc.org/golang.org/x/text) package

### See also

[jargon](https://github.com/clipperhouse/jargon), a text pipelines package for CLI and Go, which consumes this package.

### Prior art

[blevesearch/segment](https://github.com/blevesearch/segment)

[rivo/uniseg](https://github.com/rivo/uniseg)
