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

### Conformance

We use the official [test suite](https://unicode.org/reports/tr41/tr41-26.html#Tests29), thanks to [bleve](https://github.com/blevesearch/segment/blob/master/tables_test.go). Status:

![Go](https://github.com/clipperhouse/uax29/workflows/Go/badge.svg)

### Performance

`uax29` is designed to work in constant memory, regardless of input size. It buffers input and streams tokens. (For example, I am showing a maximum resident size of 8MB when processing a 300MB file.)

Execution time is `O(n)` on input size. It can be I/O bound; I/O performance is determined by the `io.Reader` you pass to `NewScanner`.

In my local testing (Mac laptop), `uax29/words` processes around 14MM tokens per second of [wiki text](https://en.wikipedia.org/w/index.php?title=New_York_City&action=edit).
