An implementation of "phrase boundaries", a variation on words boundaries from [Unicode text segmentation](https://unicode.org/reports/tr29/#Word_Boundaries) (UAX 29).

"Phrases" are not a Unicode standard, it is our definition that we think may be useful. We define it as "a series of words separated only by spaces". Punctuation breaks phrases. Emojis are treated as words.

## Quick start

```
go get "github.com/clipperhouse/uax29/phrases"
```

```go
import "github.com/clipperhouse/uax29/phrases"

text := "Hello, 世界. Nice — and totally adorable — dog; perhaps the "best one"! 🏆 🐶"

tokens := phrases.FromString(text)

for tokens.Next() {                         // Next() returns true until end of data
	fmt.Printf("%q\n", tokens.Text())       // Do something with the current phrase
}
```

[![Documentation](https://pkg.go.dev/badge/github.com/clipperhouse/uax29/phrases.svg)](https://pkg.go.dev/github.com/clipperhouse/uax29/phrases)

_Note: this package will return all tokens, including punctuation — it's not strictly "phrases" in the common sense. For our purposes, "segment", "phrase", and "token" are used synonymously._

## APIs

### If you have a `string`

Use `FromString`:

```go
text := "Hello, 世界. Nice dog! 👍🐶"

tokens := phrases.FromString(text)

for tokens.Next() {                         // Next() returns true until end of data
	fmt.Printf("%q\n", tokens.Text())       // Do something with the current phrase
}
```

### If you have an `io.Reader`

Use `FromReader`. It embeds a [`bufio.Scanner`](https://pkg.go.dev/bufio#Scanner), so you can use those methods.

```go
r := getYourReader()                            // from a file or network maybe
tokens := phrases.FromReader(r)

for tokens.Scan() {                             // Scan() returns true until error or EOF
	fmt.Println(tokens.Text())                  // Do something with the current phrase
}

if tokens.Err() != nil {                        // Check the error
	log.Fatal(tokens.Err())
}
```

### If you have a `[]byte`

Use `FromBytes`.

```go
b := []byte("Hello, 世界. Nice dog! 👍🐶")

tokens := phrases.FromBytes(b)

for tokens.Next() {                            // Next() returns true until end of data
	fmt.Printf("%q\n", tokens.Bytes())         // Do something with the current phrase
}
```

### Performance

On a Mac M2 laptop, we see around 240MB/s, which works out to around 30 million phrases (tokens, really) per second.

You should see approximately constant memory, independent of data size. We iterate tokens instead of collecting them into a slice.

### Invalid inputs

Invalid UTF-8 input is considered undefined behavior. We test to ensure that bad inputs will not cause pathological outcomes, such as a panic or infinite loop. Callers should expect "garbage-in, garbage-out".

Your pipeline should probably include a call to [`utf8.Valid()`](https://pkg.go.dev/unicode/utf8#Valid).
