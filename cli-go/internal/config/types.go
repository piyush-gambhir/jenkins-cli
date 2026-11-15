package config

type Config struct {
	CurrentProfile string             `yaml:"current_profile"`
	Profiles       map[string]Profile `yaml:"profiles"`
	Defaults       Defaults           `yaml:"defaults"`
}

type Profile struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Token    string `yaml:"token"`
	Insecure bool   `yaml:"insecure,omitempty"`
	ReadOnly bool   `yaml:"read_only,omitempty"`
}

type Defaults struct {
	Output string `yaml:"output,omitempty"`
}
