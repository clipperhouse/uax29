This package tokenizes (splits) words, sentences and graphemes, based on [Unicode text segmentation](https://unicode.org/reports/tr29/) (UAX #29), for Unicode version 15.0.0. Details and usage are in the respective packages:

[uax29/words](https://github.com/clipperhouse/uax29/tree/master/words)

[uax29/sentences](https://github.com/clipperhouse/uax29/tree/master/sentences)

[uax29/graphemes](https://github.com/clipperhouse/uax29/tree/master/graphemes)

### Why tokenize?

Any time our code operates on individual words, we are tokenizing. Often, we do it ad hoc, such as splitting on spaces, which gives inconsistent results. The Unicode standard is better: it is multi-lingual, and handles punctuation, special characters, etc.

### Conformance

We use the official [test suites](https://unicode.org/reports/tr41/tr41-26.html#Tests29). Status:

![Go](https://github.com/clipperhouse/uax29/actions/workflows/gotest.yml/badge.svg)

### See also

[jargon](https://github.com/clipperhouse/jargon), a text pipelines package for CLI and Go, which consumes this package.

### Prior art

[blevesearch/segment](https://github.com/blevesearch/segment)

[rivo/uniseg](https://github.com/rivo/uniseg)

### Other language implementations

[C#](https://github.com/clipperhouse/uax29.net) (also by me)

[JavaScript](https://github.com/tc39/proposal-intl-segmenter)

[Rust](https://unicode-rs.github.io/unicode-segmentation/unicode_segmentation/trait.UnicodeSegmentation.html)

[Java](https://lucene.apache.org/core/6_5_0/core/org/apache/lucene/analysis/standard/StandardTokenizer.html)

[Python](https://uniseg-python.readthedocs.io/en/latest/)
