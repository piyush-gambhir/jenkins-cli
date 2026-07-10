package cmd

import "testing"

func TestEnvFlagEnabled(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  bool
	}{{"true", true}, {"TRUE", true}, {"1", true}, {"false", false}, {"0", false}, {"anything", false}} {
		t.Setenv("JENKINS_TEST_FLAG", tc.value)
		if got := envFlagEnabled("JENKINS_TEST_FLAG"); got != tc.want {
			t.Errorf("envFlagEnabled(%q) = %v, want %v", tc.value, got, tc.want)
		}
	}
}
