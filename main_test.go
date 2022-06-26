package main_test

import (
	"testing"

	"github.com/macrat/ayd-ntp-probe"
	"github.com/macrat/ayd/lib-ayd"
)

func TestParseTarget(t *testing.T) {
	tests := []struct {
		Input  string
		Output string
	}{
		{"ntp:example.com", "ntp://example.com"},
		{"ntp://example.com", "ntp://example.com"},
		{"ntp:example.com:1230", "ntp://example.com:1230"},
		{"ntp://foo:bar@example.com:1230/path/to#abc?def=ghi", "ntp://example.com:1230"},
	}

	for _, tt := range tests {
		t.Run(tt.Input, func(t *testing.T) {
			u, err := ayd.ParseURL(tt.Input)
			if err != nil {
				t.Fatalf("failed to parse input url: %s", err)
			}

			u = main.NormalizeTarget(u)

			if u.String() != tt.Output {
				t.Errorf("unexpected output: %s", u)
			}
		})
	}
}
