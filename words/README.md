An implementation of word boundaries from [Unicode text segmentation](https://unicode.org/reports/tr29/#Word_Boundaries) (UAX 29), for Unicode version 13.0.0.

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

if segments.Err() != nil {                      // Check the error
	log.Fatal(segments.Err())
}
```

[![Documentation](https://pkg.go.dev/badge/github.com/clipperhouse/uax29/words.svg)](https://pkg.go.dev/github.com/clipperhouse/uax29/words)

_Note: this package will return all tokens, including whitespace and punctuation — it's not strictly “words” in the common sense. If you wish to omit things like whitespace and punctuation, you can use a filter (see below). For our purposes, “segment”, “word”, and “token” are used synonymously._

## Conformance

We use the Unicode [test suite](https://unicode.org/reports/tr41/tr41-26.html#Tests29). Status:

![Go](https://github.com/clipperhouse/uax29/workflows/Go/badge.svg)

## APIs

#### If you have a `[]byte`

Use `Segmenter` for bounded memory and best performance:

```go
text := []byte("Hello, 世界. Nice dog! 👍🐶")

segments := words.NewSegmenter(text)            // A segmenter is an iterator over the words

for segments.Next() {                           // Next() returns true until end of data or error
	fmt.Printf("%q\n", segments.Bytes())        // Do something with the current word
}

if segments.Err() != nil {                      // Check the error
	log.Fatal(segments.Err())
}
```

Use `SegmentAll()` if you prefer brevity, and are not too concerned about allocations.

```go
segments := words.SegmentAll(text)             // Returns a slice of byte slices; each slice is a word

fmt.Println("Words: %q", segments)
```

#### If you have an `io.Reader`

Use `Scanner`

```go
r := getYourReader()                            // from a file or network maybe
scanner := words.NewScanner(r)

for scanner.Scan() {                            // Scan() returns true until error or EOF
	fmt.Println(scanner.Text())                 // Do something with the current word
}

if scanner.Err() != nil {                       // Check the error
	log.Fatal(scanner.Err())
}
```

### Performance

On a Mac laptop, we see around 100MB/s, which works out to around 30 million words (tokens, really) per second.

You should see approximately constant memory when using `Segmenter` or `Scanner`, independent of data size. When using `SegmentAll()`, expect memory to be `O(n)` on the number of words (one slice per word, 24 bytes).

### Invalid inputs

Invalid UTF-8 input is considered undefined behavior. We test to ensure that bad inputs will not cause pathological outcomes, such as a panic or infinite loop. Callers should expect “garbage-in, garbage-out”.

Your pipeline should probably include a call to [`utf8.Valid()`](https://pkg.go.dev/unicode/utf8#Valid).

### Filters

You can add a filter to a `Scanner` or `Segmenter`.

For example, the Segmenter / Scanner returns _all_ tokens, split by word boundaries. This includes things like whitespace and punctuation, which may not be what one means by “words”. By using a filter, you can omit them.

```go
text := []byte("Hello, 世界. Nice dog! 👍🐶")

segments := words.NewSegmenter(text)
segments.Filter(filter.Wordlike)

for segments.Next() {
	// Note that whitespace and punctuation are omitted.
	fmt.Printf("%q\n", segments.Bytes())
}

if segments.Err() != nil {
	log.Fatal(segments.Err())
}
```

You can write your own filters (predicates), with arbitrary logic, by implementing a `func([]byte) bool`. You can also create a filter based on Unicode categories with the [`filter.Contains`](https://pkg.go.dev/github.com/clipperhouse/uax29/iterators/filter#Contains) and [`filter.Entirely`](https://pkg.go.dev/github.com/clipperhouse/uax29/iterators/filter#Entirely) methods.

### Transforms

Tokens can be modified by adding a transformer to a `Scanner` or `Segmenter`.

You might wish to lowercase all the words, for example:

```go
text := []byte("Hello, 世界. Nice dog! 👍🐶")

segments := words.NewSegmenter(text)
segments.Transform(transformer.Lower)

for segments.Next() {
	// Note that tokens come out lowercase
	fmt.Printf("%q\n", segments.Bytes())
}

if segments.Err() != nil {
	log.Fatal(segments.Err())
}
```

Here are a [few more examples](https://pkg.go.dev/github.com/clipperhouse/uax29/iterators/transformer).

We use the [`x/text/transform`](https://pkg.go.dev/golang.org/x/text/transform) package. We can accept anything that implements the `transform.Transformer` interface. Many things in `x/text` do that, such as [runes](https://pkg.go.dev/golang.org/x/text/runes), [normalization](https://pkg.go.dev/golang.org/x/text/unicode/norm), [casing](https://pkg.go.dev/golang.org/x/text/cases), and [encoding](https://pkg.go.dev/golang.org/x/text/encoding).

### Limitations

This package follows the basic UAX #29 specification. For more effective or idiomatic treatment of words across languages, there is more that can be done, scroll down to the [“Notes:” section of the standard](https://unicode.org/reports/tr29/#Word_Boundary_Rules):

> It is not possible to provide a uniform set of rules that resolves all issues across languages or that handles all ambiguous situations within a given language. The goal for the specification presented in this annex is to provide a workable default; tailored implementations can be more sophisticated.

I also found [this article](https://www.hathitrust.org/blogs/large-scale-search/multilingual-issues-part-1-word-segmentation) helpful.