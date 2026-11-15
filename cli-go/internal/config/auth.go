package config

import (
	"fmt"
	"strconv"
	"strings"
)

type EnvLookupFunc func(string) (string, bool)

type FlagValues struct {
	Server   string
	User     string
	Token    string
	Insecure bool

	ServerSet   bool
	UserSet     bool
	TokenSet    bool
	InsecureSet bool
}

func ResolveAuth(flags FlagValues, envLookup EnvLookupFunc, cfg *Config, profileName string) (Profile, error) {
	var base Profile

	pName := profileName
	if pName == "" {
		pName = cfg.CurrentProfile
	}
	if pName != "" {
		p, ok := cfg.Profiles[pName]
		if !ok {
			return Profile{}, fmt.Errorf("profile %q not found", pName)
		}
		base = p
	}

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
		if v, ok := envLookup("JENKINS_READ_ONLY"); ok && v != "" {
			b, err := strconv.ParseBool(v)
			if err == nil {
				base.ReadOnly = b
			}
		}
	}

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

	base.URL = strings.TrimRight(base.URL, "/")

	return base, nil
}
