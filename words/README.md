An implementation of word boundaries from [Unicode text segmentation](https://unicode.org/reports/tr29/#Word_Boundaries) (UAX 29), for Unicode version 13.0.0.

### Usage

```go
import "github.com/kevwang/uax29/words"

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

[GoDoc](https://godoc.org/github.com/kevwang/uax29/words)

### Conformance

We use the official [test suite](https://unicode.org/reports/tr41/tr41-26.html#Tests29), thanks to [bleve](https://github.com/blevesearch/segment/blob/master/tables_test.go). Status:

![Go](https://github.com/kevwang/uax29/workflows/Go/badge.svg)

### Performance

`uax29` is designed to work in constant memory, regardless of input size. It buffers input and streams tokens. (For example, I am showing a maximum resident size of 8MB when processing a 300MB file.)

Execution time is `O(n)` on input size. It can be I/O bound; I/O performance is determined by the `io.Reader` you pass to `NewScanner`.

In my local benchmarking (Mac laptop), [`uax29/words`](https://github.com/kevwang/uax29/tree/master/words) processes around 25MM tokens per second, or 90MB/s, of [multi-lingual prose](https://github.com/kevwang/uax29/blob/master/words/testdata/sample.txt).

### Invalid inputs

Invalid UTF-8 input is undefined behavior. That said, we’ve worked to ensure that such inputs will not cause pathological outcomes, such as a panic or infinite loop. Callers should expect “garbage-in, garbage-out”.

There are two tests in each package, called `TestInvalidUTF8` and `TestRandomBytes`. Those tests pass, returning the invalid bytes verbatim, without a guarantee as to how they will be segmented.
