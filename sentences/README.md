An implementation of sentence boundaries from [Unicode text segmentation](https://unicode.org/reports/tr29/#Sentence_Boundaries) (UAX 29), for Unicode version 13.0.0.

## Quick start

```
go get "github.com/clipperhouse/uax29/sentences"
```

```go
import "github.com/clipperhouse/uax29/sentences"

text := []byte("Hello, 世界. Nice dog! 👍🐶")

segments := sentences.NewSegmenter(text)        // A segmenter is an iterator over the sentences

for segments.Next() {                           // Next() returns true until end of data or error
	fmt.Printf("%q\n", segments.Bytes())        // Do something with the current sentence
}

if err := segments.Err(); err != nil {          // Check the error
	log.Fatal(err)
}
```

[![Documentation](https://pkg.go.dev/badge/github.com/clipperhouse/uax29/sentences.svg)](https://pkg.go.dev/github.com/clipperhouse/uax29/sentences)

_For our purposes, “segment”, “sentence”, and “token” are used synonymously._

## Conformance

We use the Unicode [test suite](https://unicode.org/reports/tr41/tr41-26.html#Tests29). Status:

![Go](https://github.com/clipperhouse/uax29/actions/workflows/gotest.yml/badge.svg)

## APIs

### If you have a `[]byte`

Use `Segmenter` for bounded memory and best performance:

```go
text := []byte("Hello, 世界. Nice dog! 👍🐶")

segments := sentences.NewSegmenter(text)        // A segmenter is an iterator over the sentences

for segments.Next() {                           // Next() returns true until end of data or error
	fmt.Printf("%q\n", segments.Bytes())        // Do something with the current sentence
}

if err := segments.Err(); err != nil {          // Check the error
	log.Fatal(err)
}
```

Use `SegmentAll()` if you prefer brevity, are not too concerned about allocations, or would be populating a `[][]byte` anyway.

```go
text := []byte("Hello, 世界. Nice dog! 👍🐶")
segments := sentences.SegmentAll(text)          // Returns a slice of byte slices; each slice is a sentence

fmt.Println("Graphemes: %q", segments)
```

### If you have an `io.Reader`

Use `Scanner` (which is a [`bufio.Scanner`](https://pkg.go.dev/bufio#Scanner), those docs will tell you what to do).

```go
r := getYourReader()                            // from a file or network maybe
scanner := sentences.NewScanner(r)

for scanner.Scan() {                            // Scan() returns true until error or EOF
	fmt.Println(scanner.Text())                 // Do something with the current sentence
}

if err := scanner.Err(); err != nil {           // Check the error
	log.Fatal(err)
}
```

### Performance

On a Mac laptop, we see around 35MB/s, which works out to around 180 thousand sentences per second.

You should see approximately constant memory when using `Segmenter` or `Scanner`, independent of data size. When using `SegmentAll()`, expect memory to be `O(n)` on the number of sentences.

### Invalid inputs

Invalid UTF-8 input is considered undefined behavior. We test to ensure that bad inputs will not cause pathological outcomes, such as a panic or infinite loop. Callers should expect “garbage-in, garbage-out”.

Your pipeline should probably include a call to [`utf8.Valid()`](https://pkg.go.dev/unicode/utf8#Valid).
