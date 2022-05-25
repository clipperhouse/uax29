An implementation of word boundaries from [Unicode text segmentation](https://unicode.org/reports/tr29/#Grapheme_Cluster_Boundaries) (UAX 29), for Unicode version 13.0.0.

## Quick start

```
go get "github.com/clipperhouse/uax29/words"
```

```go
import "github.com/clipperhouse/uax29/words"

text := []byte("Hello, 世界. Nice dog! 👍🐶")

segments := words.NewSegmenter(text)            // A segmenter is an iterator over the words

for segments.Next() {                           // Next() returns true until end of data or error
	fmt.Printf("%q\n", segments.Bytes())        // Do something with the current token
}

if err := segments.Err(); err != nil {          // Check the error
	log.Fatal(err)
}
```

[![Documentation](https://pkg.go.dev/badge/github.com/clipperhouse/uax29/words.svg)](https://pkg.go.dev/github.com/clipperhouse/uax29/words)

_Note: this package will return all tokens, including whitespace and punctuation — it's not strictly “words” in the common sense. If you wish to omit things like whitespace and punctuation, you can use a filter (see below). For our purposes, “segment”, “word”, and “token” are used synonymously._

## Conformance

We use the Unicode [test suite](https://unicode.org/reports/tr41/tr41-26.html#Tests29). Status:

![Go](https://github.com/clipperhouse/uax29/workflows/Go/badge.svg)

## APIs

### If you have a `[]byte`

Use `Segmenter` for bounded memory and best performance:

```go
text := []byte("Hello, 世界. Nice dog! 👍🐶")

segments := words.NewSegmenter(text)            // A segmenter is an iterator over the words

for segments.Next() {                           // Next() returns true until end of data or error
	fmt.Printf("%q\n", segments.Bytes())        // Do something with the current word
}

if err := segments.Err(); err != nil {          // Check the error
	log.Fatal(err)
}
```

Use `SegmentAll()` if you prefer brevity, are not too concerned about allocations, or would be populating a `[][]byte` anyway.

```go
text := []byte("Hello, 世界. Nice dog! 👍🐶")
segments := words.SegmentAll(text)             // Returns a slice of byte slices; each slice is a word

fmt.Println("Graphemes: %q", segments)
```

### If you have an `io.Reader`

Use `Scanner` (which is a [`bufio.Scanner`](https://pkg.go.dev/bufio#Scanner), those docs will tell you what to do).

```go
r := getYourReader()                            // from a file or network maybe
scanner := words.NewScanner(r)

for scanner.Scan() {                            // Scan() returns true until error or EOF
	fmt.Println(scanner.Text())                 // Do something with the current word
}

if err := scanner.Err(); err != nil {           // Check the error
	log.Fatal(err)
}
```

### Performance

On a Mac laptop, we see around 100MB/s, which works out to around 30 million words (word boundaries, really) per second.

You should see approximately constant memory when using `Segmenter` or `Scanner`, independent of data size. When using `SegmentAll()`, expect memory to be `O(n)` on the number of words.

### Invalid inputs

Invalid UTF-8 input is considered undefined behavior. We test to ensure that bad inputs will not cause pathological outcomes, such as a panic or infinite loop. Callers should expect “garbage-in, garbage-out”.

Your pipeline should probably include a call to [`utf8.Valid()`](https://pkg.go.dev/unicode/utf8#Valid).

### Filters

v1.8 adds the ability to filter tokens (segments). This is done by adding a filter to the Scanner or Segmenter.

For example, the Segmenter / Scanner returns _all_ tokens, split by word boundaries. This includes things like whitespace and punctuation, which are not what we think of as "words". By using a filter, you can omit them.

```go
text := []byte("Hello, 世界. Nice dog! 👍🐶")

segments := words.NewSegmenter(text)
segments.Filter(filter.Wordlike)

for segments.Next() {
	// Note that whitespace and punctuation are omitted.
	fmt.Printf("%q\n", segments.Bytes())
}

if err := segments.Err(); err != nil {
	log.Fatal(err)
}
```

You can write your own filters (predicates), with arbitrary logic, by implementing a `func([]byte) bool`.