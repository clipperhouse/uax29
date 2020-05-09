package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/format"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/clipperhouse/uax29/triegen"
)

type prop struct {
	url         string
	packagename string
}

func main() {
	props := []prop{
		{
			url:         "https://www.unicode.org/Public/" + unicode.Version + "/ucd/auxiliary/WordBreakProperty.txt",
			packagename: "words",
		},
		{
			url:         "https://www.unicode.org/Public/" + unicode.Version + "/ucd/auxiliary/GraphemeBreakProperty.txt",
			packagename: "graphemes",
		},
		{
			url:         "https://www.unicode.org/Public/" + unicode.Version + "/ucd/auxiliary/SentenceBreakProperty.txt",
			packagename: "sentences",
		},
		{
			url:         "https://unicode.org/Public/emoji/12.0/emoji-data.txt",
			packagename: "emoji",
		},
	}

	for _, prop := range props {
		err := generate(prop)
		if err != nil {
			panic(err)
		}
	}
}

func generate(prop prop) error {
	fmt.Println(prop.url)
	resp, err := http.Get(prop.url)
	if err != nil {
		return err
	}

	b := bufio.NewReader(resp.Body)

	runesByCategory := map[string][]rune{}
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
		category := strings.TrimSpace(split2[0])

		runesByCategory[category] = append(runesByCategory[category], runes...)
	}

	// Keep the order stable
	categories := make([]string, 0, len(runesByCategory))
	for category := range runesByCategory {
		categories = append(categories, category)
	}
	sort.Strings(categories)

	iotasByCategory := map[string]uint64{}
	for i, category := range categories {
		fmt.Printf("%s: %d\n", category, 1<<i)
		iotasByCategory[category] = 1 << i
	}

	iotasByRune := map[rune]uint64{}
	for category, runes := range runesByCategory {
		for _, r := range runes {
			iotasByRune[r] = iotasByRune[r] | iotasByCategory[category]
		}
	}

	trie := triegen.NewTrie(prop.packagename)

	for r, iotas := range iotasByRune {
		trie.Insert(r, iotas)
	}

	err = write(prop, trie, iotasByCategory)
	if err != nil {
		return err
	}

	return nil
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

func write(prop prop, trie *triegen.Trie, iotasByCategory map[string]uint64) error {
	buf := bytes.Buffer{}

	fmt.Fprintln(&buf, "package "+prop.packagename)
	fmt.Fprintln(&buf, "\n// generated by github.com/clipperhouse/uax29\n// from "+prop.url)

	// Keep the order stable
	categories := make([]string, 0, len(iotasByCategory))
	for category := range iotasByCategory {
		categories = append(categories, category)
	}
	sort.Strings(categories)

	fmt.Fprintln(&buf, "var(")
	for _, category := range categories {
		fmt.Fprintf(&buf, "b%s uint32 = %d\n", category, iotasByCategory[category])
	}
	fmt.Fprintln(&buf, ")")

	triegen.Gen(&buf, prop.packagename, []*triegen.Trie{trie})

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	dst, err := os.Create(prop.packagename + "/trie.go")
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
