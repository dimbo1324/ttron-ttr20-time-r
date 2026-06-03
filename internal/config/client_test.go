package config

import "testing"

func TestLoadClientDefaults(t *testing.T) {
	cfg, err := LoadClient(nil)
	if err != nil {
		t.Fatalf("LoadClient defaults failed: %v", err)
	}
	if cfg.Host != "127.0.0.1" || cfg.Port != 9000 || cfg.CRCMode != "sum" || cfg.AdapterAddr != 1 {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestLoadClientValidation(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "empty host", args: []string{"-host", ""}},
		{name: "bad port", args: []string{"-port", "70000"}},
		{name: "bad adapter", args: []string{"-adapter", "256"}},
		{name: "bad checksum", args: []string{"-crc", "bad"}},
		{name: "bad timeout", args: []string{"-timeout", "0"}},
		{name: "bad retries", args: []string{"-retries", "-1"}},
		{name: "bad poll step", args: []string{"-pollstep", "0"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := LoadClient(tt.args); err == nil {
				t.Fatalf("LoadClient(%v) succeeded, want error", tt.args)
			}
		})
	}
}
