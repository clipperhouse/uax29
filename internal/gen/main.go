// Package main generates tries of Unicode properties by calling go generate as the repository root
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/format"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/v2/internal/gen/triegen"
	"golang.org/x/text/unicode/rangetable"
)

func main() {
	props := []prop{
		// make sure emoji goes first, subsequent props need it
		{
			name: "Emoji",
			url:  "https://www.unicode.org/Public/" + unicode.Version + "/ucd/emoji/emoji-data.txt",
		},
		{
			name: "Word",
		},
		{
			name: "Phrase",
		},
		{
			name: "Grapheme",
		},
		{
			name: "Sentence",
		},
	}

	for _, p := range props {
		if err := p.generateTrie(); err != nil {
			panic(err)
		}

		if err := p.generateTests(); err != nil {
			panic(err)
		}
	}
}

const baseURL = "https://www.unicode.org/Public/" + unicode.Version + "/ucd/auxiliary"

type prop struct {
	name string
	url  string
}

func (p prop) URL() string {
	if p.url != "" {
		return p.url
	}

	if p.name == "Phrase" {
		p.name = "Word"
	}

	return fmt.Sprintf("%s/%sBreakProperty.txt", baseURL, p.name)
}

func (p prop) TestURL() string {
	if p.name == "Emoji" {
		panic("no tests for emoji")
	}
	return fmt.Sprintf("%s/%sBreakTest.txt", baseURL, p.name)
}

func (p prop) PackageName() string {
	return strings.ToLower(p.name) + "s"
}

var extendedPictographic []rune

func (p prop) generateTrie() error {
	fmt.Println(p.URL())
	resp, err := http.Get(p.URL())
	if err != nil {
		return err
	}

	b := bufio.NewReader(resp.Body)

	runesByProperty := map[string][]rune{}
	for {
		s, err := b.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if len(s) == 0 {
			continue
		}

		if s[0] == '\n' || s[0] == '#' {
			continue
		}

		parts := strings.Split(s, ";")
		runes, err := getRuneRange(parts[0])
		if err != nil {
			return err
		}

		split2 := strings.Split(parts[1], "#")
		property := strings.TrimSpace(split2[0])

		runesByProperty[property] = append(runesByProperty[property], runes...)
	}

	// Words and graphemes need Extended_Pictographic property
	const key = "Extended_Pictographic"
	if p.name == "Emoji" {
		extendedPictographic = runesByProperty[key]
		// We don't need to generate emoji package
		return nil
	}
	if p.name == "Word" || p.name == "Phrase" || p.name == "Grapheme" {
		if len(extendedPictographic) == 0 {
			panic("didn't get emoji data; make sure it's loaded first")
		}
		runesByProperty[key] = extendedPictographic
	}

	if p.name == "Word" {
		// Concatenate UAX 29 definition of Katakana with Han and Hiragana
		// The rangetable unicode.Katakana isn't complete for
		// our purposes, see https://www.unicode.org/reports/tr29/tr29-37.html#Katakana
		table := rangetable.Merge(unicode.Han, unicode.Hiragana)
		var ideo []rune
		rangetable.Visit(table, func(r rune) {
			ideo = append(ideo, r)
		})
		runesByProperty["BleveIdeographic"] = append(runesByProperty["Katakana"], ideo...)
	}

	// Keep the order stable
	properties := make([]string, 0, len(runesByProperty))
	for property := range runesByProperty {
		properties = append(properties, property)
	}
	sort.Strings(properties)

	iotasByProperty := map[string]uint64{}
	for i, property := range properties {
		iotasByProperty[property] = 1 << i
	}

	iotasByRune := map[rune]uint64{}
	for property, runes := range runesByProperty {
		for _, r := range runes {
			iotasByRune[r] = iotasByRune[r] | iotasByProperty[property]
		}
	}

	trie := triegen.NewTrie(p.PackageName())

	for r, iotas := range iotasByRune {
		trie.Insert(r, iotas)
	}

	err = writeTrie(p, trie, iotasByProperty)
	if err != nil {
		return err
	}

	return nil
}

type unicodeTest struct {
	input    []byte
	expected [][]byte
	comment  string
}

func (p prop) generateTests() error {
	if p.name == "Emoji" {
		return nil
	}
	if p.name == "Phrase" {
		return nil
	}
	fmt.Println(p.TestURL())
	resp, err := http.Get(p.TestURL())
	if err != nil {
		return err
	}

	sc := bufio.NewScanner(resp.Body) // defaults to ScanLines

	var unicodeTests []unicodeTest
	for sc.Scan() {
		line := sc.Text()
		if line[0] == '#' {
			// comment line, ignore
			continue
		}

		test := unicodeTest{}

		parts := strings.Split(line, "#")

		test.comment = strings.TrimSpace(parts[1])

		data := parts[0]

		segments := strings.Split(data, "÷")

		for _, segment := range segments {
			if len(segment) < 4 {
				continue
			}

			vals := strings.Split(segment, " ")

			var expected []byte
			for _, val := range vals {
				if len(val) < 4 {
					continue
				}

				hex := "0x" + val
				r64, err := strconv.ParseInt(hex, 0, 32)
				if err != nil {
					return err
				}

				r := rune(r64)
				rb := make([]byte, utf8.RuneLen(r))
				utf8.EncodeRune(rb, r)
				test.input = append(test.input, rb...)
				expected = append(expected, rb...)
			}
			test.expected = append(test.expected, expected)
		}

		unicodeTests = append(unicodeTests, test)
	}

	if err := sc.Err(); err != nil {
		return err
	}

	return p.writeTests(unicodeTests)
}

func getRuneRange(s string) ([]rune, error) {
	s = strings.TrimSpace(s)
	hilo := strings.Split(s, "..")
	lo64, err := strconv.ParseInt("0x"+hilo[0], 0, 32)
	if err != nil {
		return nil, err
	}

	lo := rune(lo64)
	runes := []rune{lo}

	if len(hilo) == 1 {
		return runes, nil
	}

	hi64, err := strconv.ParseInt("0x"+hilo[1], 0, 32)
	if err != nil {
		return nil, err
	}

	hi := rune(hi64)
	if hi == lo {
		return runes, nil
	}

	// Skip first, inclusive of last
	for r := lo + 1; r <= hi; r++ {
		runes = append(runes, r)
	}

	return runes, nil
}

func (p prop) writeTests(tests []unicodeTest) error {
	buf := bytes.Buffer{}

	fmt.Fprintf(&buf, `package %s_test

	// generated by github.com/clipperhouse/uax29/v2
	// from %s
`, p.PackageName(), p.TestURL())

	fmt.Fprintf(&buf, `
type unicodeTest struct {
	input    []byte
	expected [][]byte
	comment  string
}

var unicodeTests = [%d]unicodeTest {
`, len(tests))

	for _, t := range tests {
		fmt.Fprintf(&buf, `{
	input: %#v,
	expected: %#v,
	comment: %#v,
},
`, t.input, t.expected, t.comment)
	}

	fmt.Fprintf(&buf, "}\n\n")

	byts := buf.Bytes()
	byts = bytes.ReplaceAll(byts, []byte("[]uint8{0x"), []byte("{0x"))
	byts = bytes.ReplaceAll(byts, []byte("[]uint8{"), []byte("[]byte{"))

	formatted, err := format.Source(byts)
	if err != nil {
		return err
	}

	f := filepath.Join("../../", p.PackageName(), "unicode_test.go")
	dst, err := os.Create(f)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = dst.Write(formatted)
	if err != nil {
		return err
	}

	return nil
}

func writeTrie(prop prop, trie *triegen.Trie, iotasByProperty map[string]uint64) error {
	buf := bytes.Buffer{}

	fmt.Fprintln(&buf, "package "+prop.PackageName())
	fmt.Fprintln(&buf, "\n// generated by github.com/clipperhouse/uax29/v2\n// from "+prop.URL())
	fmt.Fprintln(&buf)

	// Keep the order stable
	properties := make([]string, 0, len(iotasByProperty))
	for property := range iotasByProperty {
		properties = append(properties, property)
	}
	sort.Strings(properties)

	inttype := ""
	len := len(properties)
	switch {
	case len < 8:
		inttype = "uint8"
	case len < 16:
		inttype = "uint16"
	case len < 32:
		inttype = "uint32"
	default:
		inttype = "uint64"
	}

	fmt.Fprintf(&buf, "type property %s\n\n", inttype)

	fmt.Fprintln(&buf, "const (")
	for i, property := range properties {
		name := strings.ReplaceAll(property, "_", "")
		if i == 0 {
			fmt.Fprintf(&buf, "_%s property = 1 << iota\n", name)
			continue
		}
		fmt.Fprintf(&buf, "_%s\n", name)
	}
	fmt.Fprintln(&buf, ")")

	_, err := triegen.Gen(&buf, prop.PackageName(), []*triegen.Trie{trie})
	if err != nil {
		return err
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	f := filepath.Join("../../", prop.PackageName(), "trie.go")

	dst, err := os.Create(f)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = dst.Write(formatted)
	if err != nil {
		return err
	}

	return nil
}
