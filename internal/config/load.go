package config

// Load reads and validates the YAML config at path. Decoding is strict
// (KnownFields, D-33): an unknown key rejects the whole document; an empty
// document is a validate-time refusal — an explicit -config must yield a
// working config, never silent defaults.
func Load(path string) (*Config, error) {
	return &Config{}, nil // RED stub: the corpus pins the strict decode contract.
}
