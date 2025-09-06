An implementation of “phrase boundaries”, a variation on words boundaries from [Unicode text segmentation](https://unicode.org/reports/tr29/#Word_Boundaries) (UAX 29).

“Phrases” are not a Unicode standard, it is our definition that we think may be useful. We define it as “a series of words separated only by spaces”. Punctuation breaks phrases. Emojis are treated as words.

## Quick start

```
go get "github.com/clipperhouse/uax29/phrases"
```

```go
text := []byte("Hello, 世界. Nice — and totally adorable — dog; perhaps the “best one”! 🏆 🐶")

phrase := phrases.NewSegmenter(text)

// Next returns true until error or end of data
for phrase.Next() {
	// Do something with the phrase
	fmt.Printf("%q\n", phrase.Bytes())
}

// Gotta check the error!
if err := phrase.Err(); err != nil {
	log.Fatal(err)
}
// Output: "Hello"
// ","
// " "
// "世"
// "界"
// "."
// " Nice "
// "—"
// " and totally adorable "
// "—"
// " dog"
// ";"
// " perhaps the "
// "“"
// "best one"
// "”"
// "!"
// " 🏆 🐶"
```

[![Documentation](https://pkg.go.dev/badge/github.com/clipperhouse/uax29/phrases.svg)](https://pkg.go.dev/github.com/clipperhouse/uax29/phrases)

_Note: this package will return all tokens, including punctuation — it's not strictly “phrases” in the common sense. If you wish to omit things certain tokens, use a filter (see below). For our purposes, “segment”, “phrase”, and “token” are used synonymously._

## APIs

#### If you have a `[]byte`

Use `Segmenter` for bounded memory and best performance:

```go
text := []byte("Hello, 世界. Nice dog! 👍🐶")

segments := phrases.NewSegmenter(text)            // A segmenter is an iterator over the phrases

for segments.Next() {                           // Next() returns true until end of data or error
	fmt.Printf("%q\n", segments.Bytes())        // Do something with the current phrase
}

if segments.Err() != nil {                      // Check the error
	log.Fatal(segments.Err())
}
```

Use `SegmentAll()` if you prefer brevity, and are not too concerned about allocations.

```go
segments := phrases.SegmentAll(text)             // Returns a slice of byte slices; each slice is a phrase

fmt.Println("phrases: %q", segments)
```

#### If you have an `io.Reader`

Use `Scanner`

```go
r := getYourReader()                            // from a file or network maybe
scanner := phrases.NewScanner(r)

for scanner.Scan() {                            // Scan() returns true until error or EOF
	fmt.Println(scanner.Text())                 // Do something with the current phrase
}

if scanner.Err() != nil {                       // Check the error
	log.Fatal(scanner.Err())
}
```

### Performance

On a Mac M2 laptop, we see around 240MB/s, which works out to around 30 million phrases (tokens, really) per second.

You should see approximately constant memory when using `Segmenter` or `Scanner`, independent of data size. When using `SegmentAll()`, expect memory to be `O(n)` on the number of phrases (one slice per phrase).

### Uses

The uax29 module has 4 tokenizers. In decreasing granularity: sentences → phrases → words → graphemes.

For best results, you may wish to first [split sentences](https://github.com/clipperhouse/uax29/tree/master/sentences), and _then_ split phrases within those sentences.

If you're doing embeddings, the definition of “meaningful unit” will depend on your application. You might choose sentences, phrases, words, or all of the above. You can tokenize the tokens of other tokenizers.

### Invalid inputs

Invalid UTF-8 input is considered undefined behavior. We test to ensure that bad inputs will not cause pathological outcomes, such as a panic or infinite loop. Callers should expect “garbage-in, garbage-out”.

Your pipeline should probably include a call to [`utf8.Valid()`](https://pkg.go.dev/unicode/utf8#Valid).

### Filters

You can add a filter to a `Scanner` or `Segmenter`.

For example, the tokenizer returns _all_ tokens, split by phrase boundaries. This may includes things like punctuation, which may not be what one means by “phrases”. By using a filter, you can omit them.

```go
text := []byte("Hello, 世界. Nice dog! 👍🐶")

segments := phrases.NewSegmenter(text)
segments.Filter(filter.Wordlike)

for segments.Next() {
	fmt.Printf("%q\n", segments.Bytes())
}

if segments.Err() != nil {
	log.Fatal(segments.Err())
}
```

You can write your own filters (predicates), with arbitrary logic, by implementing a `func([]byte) bool`. You can also create a filter based on Unicode categories with the [`filter.Contains`](https://pkg.go.dev/github.com/clipperhouse/uax29/iterators/filter#Contains) and [`filter.Entirely`](https://pkg.go.dev/github.com/clipperhouse/uax29/iterators/filter#Entirely) methods.

### Limitations

This package follows derives from the basic UAX #29 specification. For more idiomatic treatment of phrases across languages, there is more that can be done, scroll down to the [“Notes:” section of the standard](https://unicode.org/reports/tr29/#Word_Boundary_Rules):

> It is not possible to provide a uniform set of rules that resolves all issues across languages or that handles all ambiguous situations within a given language. The goal for the specification presented in this annex is to provide a workable default; tailored implementations can be more sophisticated.

I also found [this article](https://www.hathitrust.org/blogs/large-scale-search/multilingual-issues-part-1-word-segmentation) helpful.
