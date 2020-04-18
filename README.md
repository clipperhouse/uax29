An implementation of word boundaries from [Unicode text segmentation](https://unicode.org/reports/tr29/#Word_Boundaries) (UAX 29), for Unicode version 12.1.

Originally created for use with [jargon](https://github.com/clipperhouse/jargon), a text pipelines package for CLI and Go.

### Usage

```go
text := "This is an example."
reader := strings.NewReader(text)

scanner := words.NewScanner(reader)

// Scan returns true until error or EOF
for scanner.Scan() {
	fmt.Printf("%q\n", scanner.Text())
}

// Gotta check the error (primarily because we depend on I/O).
// The Scan method will return false on error, falling through to here.
if err := scanner.Err(); err != nil {
	log.Fatal(err)
}
```

### Performance

`uax29` is designed to work in constant memory, regardless of input size. It buffers input and streams tokens.

Execution time is designed to be `O(n)` on input size. It is I/O-bound. In your code, you control I/O and performance implications by the `Reader` you pass to `words.NewScanner`.

If I am reading my benchmarks correctly, `uax29/words` processes around 2MM tokens per second on my laptop, when the input is preloaded into memory. By default, it buffers 64K of input at a time.

### Conformance

The [spec](https://unicode.org/reports/tr29/#Word_Boundaries) has many nods to practicality and judgment for the implementer. One place where we vary from the strict spec is to consider underscore `_` a valid mid-word/mid-number character, helpful for things like user_names.

(Counterargument: separate tokens should be preserved for search applications.)

We have not implemented the official [test suite](https://unicode.org/reports/tr41/tr41-26.html#Tests29). Maybe you’d like to help!

### Status

- Most of the [word boundary rules](https://unicode.org/reports/tr29/#Word_Boundaries) have been implemented. We code-gen the Unicode categories relevant to UAX 29 by running `go generate` at the repository root.

- Tests are very basic. We’d like to get the [test suite](https://unicode.org/reports/tr41/tr41-26.html#Tests29) implemented. The folks at bleve [offer an example](https://github.com/blevesearch/segment/blob/master/tables_test.go).

- Support for [grapheme rules](https://unicode.org/reports/tr29/#Grapheme_Cluster_Boundaries) and perhaps [sentence rules](https://unicode.org/reports/tr29/#Sentence_Boundaries) might be useful.

- There is [discussion](https://groups.google.com/d/msg/golang-nuts/_79vJ65KuXc/B_QgeU6rAgAJ) of implementing the above in Go’s [`x/text`](https://godoc.org/golang.org/x/text) package.