An implementation of grapheme cluster boundaries from [Unicode text segmentation](https://unicode.org/reports/tr29/#Grapheme_Cluster_Boundaries) (UAX 29), for Unicode version 12.0.

### Usage

```go
import "github.com/clipperhouse/uax29/graphemes"

text := "This is an example."
reader := strings.NewReader(text)

scanner := graphemes.NewScanner(reader)

// Scan returns true until error or EOF
for scanner.Scan() {
	fmt.Println(scanner.Text())
}

// Gotta check the error (because we depend on I/O).
if err := scanner.Err(); err != nil {
	log.Fatal(err)
}
```

[GoDoc](https://godoc.org/github.com/clipperhouse/uax29/graphemes)

### Performance

`uax29` is designed to work in constant memory, regardless of input size. It buffers input and streams tokens.

Execution time is designed to be `O(n)` on input size. It is I/O-bound. In your code, you control I/O and performance implications by the `io.Reader` you pass to `graphemes.NewScanner`.

### Conformance

We use the official [test suite](https://unicode.org/reports/tr41/tr41-26.html#Tests29), thanks to [bleve](https://github.com/blevesearch/segment/blob/master/tables_test.go). Status:

![Go](https://github.com/clipperhouse/uax29/workflows/Go/badge.svg)
