An implementation of grapheme cluster boundaries from [Unicode text segmentation](https://unicode.org/reports/tr29/#Grapheme_Cluster_Boundaries) (UAX 29), for Unicode version 13.0.0.

## Quick start

```
go get "github.com/clipperhouse/uax29/graphemes"
```

```go
import "github.com/clipperhouse/uax29/graphemes"

text := []byte("Hello, 世界. Nice dog! 👍🐶")

segments := graphemes.NewSegmenter(text)        // A segmenter is an iterator over the graphemes

for segments.Next() {                           // Next() returns true until end of data or error
	fmt.Printf("%q\n", segments.Bytes())        // Do something with the current grapheme
}

if err := segments.Err(); err != nil {          // Check the error
	log.Fatal(err)
}
```

[![Documentation](https://pkg.go.dev/badge/github.com/clipperhouse/uax29/graphemes.svg)](https://pkg.go.dev/github.com/clipperhouse/uax29/graphemes)

_A grapheme is a “single visible character”, which might be a simple as a single letter, or a complex emoji that consists of several Unicode code points. For our purposes, “segment”, “grapheme”, and “token” are used synonymously._

## Conformance

We use the Unicode [test suite](https://unicode.org/reports/tr41/tr41-26.html#Tests29). Status:

![Go](https://github.com/clipperhouse/uax29/workflows/Go/badge.svg)

## APIs

### If you have a `[]byte`

Use `Segmenter` for bounded memory and best performance:

```go
text := []byte("Hello, 世界. Nice dog! 👍🐶")

segments := graphemes.NewSegmenter(text)        // A segmenter is an iterator over the graphemes

for segments.Next() {                           // Next() returns true until end of data or error
	fmt.Printf("%q\n", segments.Bytes())        // Do something with the current grapheme
}

if err := segments.Err(); err != nil {          // Check the error
	log.Fatal(err)
}
```

Use `SegmentAll()` if you prefer brevity, are not too concerned about allocations, or would be populating a `[][]byte` anyway.

```go
text := []byte("Hello, 世界. Nice dog! 👍🐶")
segments := graphemes.SegmentAll(text)          // Returns a slice of byte slices; each slice is a grapheme

fmt.Println("Graphemes: %q", segments)
```

### If you have an `io.Reader`

Use `Scanner` (which is a [`bufio.Scanner`](https://pkg.go.dev/bufio#Scanner), those docs will tell you what to do).

```go
r := getYourReader()                            // from a file or network maybe
scanner := graphemes.NewScanner(r)

for scanner.Scan() {                            // Scan() returns true until error or EOF
	fmt.Println(scanner.Text())                 // Do something with the current grapheme
}

if err := scanner.Err(); err != nil {           // Check the error
	log.Fatal(err)
}
```

### Performance

On a Mac laptop, we see around 70MB/s, which works out to around 70 million graphemes per second.

You should see approximately constant memory when using `Segmenter` or `Scanner`, independent of data size. When using `SegmentAll()`, expect memory to be `O(n)` on the number of graphemes.

### Invalid inputs

Invalid UTF-8 input is considered undefined behavior. We test to ensure that bad inputs will not cause pathological outcomes, such as a panic or infinite loop. Callers should expect “garbage-in, garbage-out”.

Your pipeline should probably include a call to [`utf8.Valid()`](https://pkg.go.dev/unicode/utf8#Valid).
