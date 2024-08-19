package words

type Config struct {
	leadingJoiners map[rune]struct{}
	midJoiners     map[rune]struct{}
}

var empty *Config = nil

func NewConfig() *Config {
	return &Config{}
}

// JoinMiddleCharacters specifies which characters (runes) should
// join words where they would otherwise be split. The joiners
// parameter is interpreted as a set of runes, not a string.
//
// For example, specifying "-" will ensure that hypenated words are treated as a single token;
// "@" will help to keep email addresses as a single token.
func (c *Config) JoinMiddleCharacters(joiners string) *Config {
	if c.midJoiners == nil {
		c.midJoiners = make(map[rune]struct{})
	}
	for _, r := range joiners {
		c.midJoiners[r] = struct{}{}
	}
	return c
}

// JoinLeadingCharacters specifies which characters (runes) should
// join words at the beginning of a words, where they would otherwise be split. The leaders
// parameter is interpreted as a set of runes, not a string.
//
// For example, specifying "#" will ensure that #handles are kept as a single word.
// Specifying "." will ensure preserve decimals like .01.
func (c *Config) JoinLeadingCharacters(leaders string) *Config {
	if c.leadingJoiners == nil {
		c.leadingJoiners = make(map[rune]struct{})
	}
	for _, r := range leaders {
		c.leadingJoiners[r] = struct{}{}
	}
	return c
}
