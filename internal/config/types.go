package config

// Config represents the top-level configuration file structure.
type Config struct {
	CurrentProfile string             `yaml:"current_profile"`
	Profiles       map[string]Profile `yaml:"profiles"`
	Defaults       Defaults           `yaml:"defaults"`
}

// Profile represents connection details for a Jenkins server.
type Profile struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Token    string `yaml:"token"`
	Insecure bool   `yaml:"insecure,omitempty"`
}

// Defaults holds default settings.
type Defaults struct {
	Output string `yaml:"output,omitempty"`
}
