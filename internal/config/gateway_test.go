package config

import (
	"testing"
	"time"
)

func TestLoadGatewayDefaults(t *testing.T) {
	cfg, err := LoadGateway(nil)
	if err != nil {
		t.Fatalf("LoadGateway defaults failed: %v", err)
	}
	if cfg.Target != "127.0.0.1:9000" || cfg.CRCMode != "sum" || cfg.AdapterAddr != 1 {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if cfg.LogFile != "runtime/logs/ft12-gateway.log" {
		t.Fatalf("LogFile = %q", cfg.LogFile)
	}
}

func TestLoadGatewayModeAlias(t *testing.T) {
	cfg, err := LoadGateway([]string{"-mode", "crc16", "-interval", "1s"})
	if err != nil {
		t.Fatalf("LoadGateway alias failed: %v", err)
	}
	if cfg.CRCMode != "crc16" || cfg.PollInterval != time.Second {
		t.Fatalf("alias/defaults not applied: %+v", cfg)
	}
}

func TestLoadGatewayValidation(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "empty target", args: []string{"-target", ""}},
		{name: "bad target", args: []string{"-target", "bad"}},
		{name: "bad checksum", args: []string{"-crc", "bad"}},
		{name: "bad adapter", args: []string{"-adapter", "256"}},
		{name: "bad interval", args: []string{"-interval", "0s"}},
		{name: "bad request timeout", args: []string{"-timeout", "0s"}},
		{name: "bad connect timeout", args: []string{"-connect-timeout", "0s"}},
		{name: "bad backoff initial", args: []string{"-backoff-initial", "0s"}},
		{name: "bad backoff max", args: []string{"-backoff-max", "0s"}},
		{name: "backoff order", args: []string{"-backoff-initial", "5s", "-backoff-max", "1s"}},
		{name: "bad recent", args: []string{"-recent", "0"}},
		{name: "bad grpc", args: []string{"-grpc-listen", "bad"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := LoadGateway(tt.args); err == nil {
				t.Fatalf("LoadGateway(%v) succeeded, want error", tt.args)
			}
		})
	}
}
