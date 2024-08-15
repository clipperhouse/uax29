package words

type Config struct {
	joiners map[byte]struct{}
}

var empty *Config = nil

func NewConfig(joiners ...byte) *Config {
	if len(joiners) == 0 {
		return empty
	}

	js := make(map[byte]struct{})
	for _, b := range joiners {
		js[b] = struct{}{}
	}
	return &Config{joiners: js}
}
