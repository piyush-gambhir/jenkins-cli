package config

import (
	"strconv"
	"strings"
)

// EnvLookupFunc is a function that looks up an environment variable.
type EnvLookupFunc func(string) (string, bool)

// FlagValues holds values passed via CLI flags.
type FlagValues struct {
	Server   string
	User     string
	Token    string
	Insecure bool

	// Whether flags were explicitly set
	ServerSet   bool
	UserSet     bool
	TokenSet    bool
	InsecureSet bool
}

// ResolveAuth resolves authentication details with priority: flags > env > config profile.
func ResolveAuth(flags FlagValues, envLookup EnvLookupFunc, cfg *Config, profileName string) (Profile, error) {
	// Start with config profile as base
	var base Profile

	pName := profileName
	if pName == "" {
		pName = cfg.CurrentProfile
	}
	if pName != "" {
		if p, ok := cfg.Profiles[pName]; ok {
			base = p
		}
	}

	// Layer env vars
	if envLookup != nil {
		if v, ok := envLookup("JENKINS_URL"); ok && v != "" {
			base.URL = v
		}
		if v, ok := envLookup("JENKINS_USER"); ok && v != "" {
			base.Username = v
		}
		if v, ok := envLookup("JENKINS_TOKEN"); ok && v != "" {
			base.Token = v
		}
		if v, ok := envLookup("JENKINS_INSECURE"); ok && v != "" {
			b, err := strconv.ParseBool(v)
			if err == nil {
				base.Insecure = b
			}
		}
	}

	// Layer flags (highest priority)
	if flags.ServerSet && flags.Server != "" {
		base.URL = flags.Server
	}
	if flags.UserSet && flags.User != "" {
		base.Username = flags.User
	}
	if flags.TokenSet && flags.Token != "" {
		base.Token = flags.Token
	}
	if flags.InsecureSet {
		base.Insecure = flags.Insecure
	}

	// Normalize URL
	base.URL = strings.TrimRight(base.URL, "/")

	return base, nil
}
