package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLoadConfig_Empty(t *testing.T) {
	// Point config to a temp dir that has no config file
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Profiles == nil {
		t.Fatal("expected non-nil Profiles map")
	}
	if len(cfg.Profiles) != 0 {
		t.Errorf("expected empty profiles, got %d", len(cfg.Profiles))
	}
	if cfg.CurrentProfile != "" {
		t.Errorf("expected empty current_profile, got %q", cfg.CurrentProfile)
	}
}

func TestLoadConfig_Existing(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	configDir := filepath.Join(tmpDir, "jenkins-cli")
	os.MkdirAll(configDir, 0o700)

	yamlContent := `current_profile: prod
profiles:
  prod:
    url: https://jenkins.prod.example.com
    username: deploy-user
    token: prod-token-123
  staging:
    url: https://jenkins.staging.example.com
    username: test-user
    token: staging-token-456
    insecure: true
defaults:
  output: json
`
	err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(yamlContent), 0o600)
	if err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.CurrentProfile != "prod" {
		t.Errorf("expected current_profile 'prod', got %q", cfg.CurrentProfile)
	}
	if len(cfg.Profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(cfg.Profiles))
	}

	prod := cfg.Profiles["prod"]
	if prod.URL != "https://jenkins.prod.example.com" {
		t.Errorf("prod URL = %q", prod.URL)
	}
	if prod.Username != "deploy-user" {
		t.Errorf("prod Username = %q", prod.Username)
	}
	if prod.Token != "prod-token-123" {
		t.Errorf("prod Token = %q", prod.Token)
	}

	staging := cfg.Profiles["staging"]
	if !staging.Insecure {
		t.Error("expected staging to be insecure")
	}

	if cfg.Defaults.Output != "json" {
		t.Errorf("expected defaults.output 'json', got %q", cfg.Defaults.Output)
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg := &Config{
		CurrentProfile: "local",
		Profiles: map[string]Profile{
			"local": {
				URL:      "http://localhost:8080",
				Username: "admin",
				Token:    "admin-token",
			},
		},
	}

	err := Save(cfg)
	if err != nil {
		t.Fatalf("unexpected error saving: %v", err)
	}

	// Read back
	loaded, err := Load()
	if err != nil {
		t.Fatalf("unexpected error loading: %v", err)
	}

	if loaded.CurrentProfile != "local" {
		t.Errorf("expected current_profile 'local', got %q", loaded.CurrentProfile)
	}
	p, ok := loaded.Profiles["local"]
	if !ok {
		t.Fatal("expected 'local' profile to exist")
	}
	if p.URL != "http://localhost:8080" {
		t.Errorf("expected URL 'http://localhost:8080', got %q", p.URL)
	}
	if p.Username != "admin" {
		t.Errorf("expected username 'admin', got %q", p.Username)
	}
	if p.Token != "admin-token" {
		t.Errorf("expected token 'admin-token', got %q", p.Token)
	}
}

func TestAddProfile(t *testing.T) {
	cfg := &Config{
		Profiles: make(map[string]Profile),
	}

	SetProfile(cfg, "new-profile", Profile{
		URL:      "http://new-jenkins:8080",
		Username: "new-user",
		Token:    "new-token",
	})

	if len(cfg.Profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(cfg.Profiles))
	}
	p, ok := cfg.Profiles["new-profile"]
	if !ok {
		t.Fatal("expected 'new-profile' to exist")
	}
	if p.URL != "http://new-jenkins:8080" {
		t.Errorf("expected URL 'http://new-jenkins:8080', got %q", p.URL)
	}
	if p.Username != "new-user" {
		t.Errorf("expected username 'new-user', got %q", p.Username)
	}

	// Test overwrite
	SetProfile(cfg, "new-profile", Profile{
		URL:      "http://updated:8080",
		Username: "updated-user",
		Token:    "updated-token",
	})
	p = cfg.Profiles["new-profile"]
	if p.URL != "http://updated:8080" {
		t.Errorf("expected updated URL, got %q", p.URL)
	}
}

func TestResolveAuth_FlagsWin(t *testing.T) {
	cfg := &Config{
		CurrentProfile: "default",
		Profiles: map[string]Profile{
			"default": {
				URL:      "http://config-url:8080",
				Username: "config-user",
				Token:    "config-token",
			},
		},
	}

	envLookup := func(key string) (string, bool) {
		switch key {
		case "JENKINS_URL":
			return "http://env-url:8080", true
		case "JENKINS_USER":
			return "env-user", true
		case "JENKINS_TOKEN":
			return "env-token", true
		}
		return "", false
	}

	flags := FlagValues{
		Server:    "http://flag-url:8080",
		User:      "flag-user",
		Token:     "flag-token",
		ServerSet: true,
		UserSet:   true,
		TokenSet:  true,
	}

	profile, err := ResolveAuth(flags, envLookup, cfg, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if profile.URL != "http://flag-url:8080" {
		t.Errorf("expected flag URL, got %q", profile.URL)
	}
	if profile.Username != "flag-user" {
		t.Errorf("expected flag user, got %q", profile.Username)
	}
	if profile.Token != "flag-token" {
		t.Errorf("expected flag token, got %q", profile.Token)
	}
}

func TestResolveAuth_EnvWins(t *testing.T) {
	cfg := &Config{
		CurrentProfile: "default",
		Profiles: map[string]Profile{
			"default": {
				URL:      "http://config-url:8080",
				Username: "config-user",
				Token:    "config-token",
			},
		},
	}

	envLookup := func(key string) (string, bool) {
		switch key {
		case "JENKINS_URL":
			return "http://env-url:8080", true
		case "JENKINS_USER":
			return "env-user", true
		case "JENKINS_TOKEN":
			return "env-token", true
		}
		return "", false
	}

	// No flags set
	flags := FlagValues{}

	profile, err := ResolveAuth(flags, envLookup, cfg, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if profile.URL != "http://env-url:8080" {
		t.Errorf("expected env URL, got %q", profile.URL)
	}
	if profile.Username != "env-user" {
		t.Errorf("expected env user, got %q", profile.Username)
	}
	if profile.Token != "env-token" {
		t.Errorf("expected env token, got %q", profile.Token)
	}
}

// TestSaveConfig_RoundTrip verifies YAML serialization round-trip fidelity.
func TestSaveConfig_RoundTrip(t *testing.T) {
	original := &Config{
		CurrentProfile: "test",
		Profiles: map[string]Profile{
			"test": {
				URL:      "http://example.com",
				Username: "user",
				Token:    "tok",
				Insecure: true,
			},
		},
		Defaults: Defaults{Output: "yaml"},
	}

	data, err := yaml.Marshal(original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var loaded Config
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if loaded.CurrentProfile != original.CurrentProfile {
		t.Errorf("CurrentProfile mismatch: %q vs %q", loaded.CurrentProfile, original.CurrentProfile)
	}
	p := loaded.Profiles["test"]
	if p.URL != "http://example.com" {
		t.Errorf("URL mismatch: %q", p.URL)
	}
	if !p.Insecure {
		t.Error("expected Insecure=true after round-trip")
	}
	if loaded.Defaults.Output != "yaml" {
		t.Errorf("Defaults.Output mismatch: %q", loaded.Defaults.Output)
	}
}
