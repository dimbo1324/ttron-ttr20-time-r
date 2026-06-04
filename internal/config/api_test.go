package config

import (
	"testing"
	"time"
)

func TestLoadAPIDefaults(t *testing.T) {
	cfg, err := LoadAPI(nil)
	if err != nil {
		t.Fatalf("LoadAPI defaults failed: %v", err)
	}
	if cfg.HTTPListen != ":8080" || cfg.EmulatorGRPC != "127.0.0.1:9100" || cfg.GatewayGRPC != "127.0.0.1:9200" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if cfg.RequestTimeout != 3*time.Second {
		t.Fatalf("timeout = %s", cfg.RequestTimeout)
	}
}

func TestLoadAPIFlags(t *testing.T) {
	cfg, err := LoadAPI([]string{
		"-http-listen", "127.0.0.1:18080",
		"-emulator-grpc", "127.0.0.1:19100",
		"-gateway-grpc", "127.0.0.1:19200",
		"-timeout", "5s",
		"-cors-origin", "*",
		"-log", "api.log",
	})
	if err != nil {
		t.Fatalf("LoadAPI flags failed: %v", err)
	}
	if cfg.HTTPListen != "127.0.0.1:18080" || cfg.EmulatorGRPC != "127.0.0.1:19100" || cfg.GatewayGRPC != "127.0.0.1:19200" {
		t.Fatalf("flags not applied: %+v", cfg)
	}
	if cfg.RequestTimeout != 5*time.Second || cfg.CORSOrigin != "*" || cfg.LogFile != "api.log" {
		t.Fatalf("flags not applied: %+v", cfg)
	}
}

func TestLoadAPIValidation(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "bad http listen", args: []string{"-http-listen", "bad"}},
		{name: "empty emulator", args: []string{"-emulator-grpc", ""}},
		{name: "bad emulator", args: []string{"-emulator-grpc", "bad"}},
		{name: "empty gateway", args: []string{"-gateway-grpc", ""}},
		{name: "bad gateway", args: []string{"-gateway-grpc", "bad"}},
		{name: "bad timeout", args: []string{"-timeout", "0s"}},
		{name: "bad cors", args: []string{"-cors-origin", "localhost:5173/path"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := LoadAPI(tt.args); err == nil {
				t.Fatalf("LoadAPI(%v) succeeded, want error", tt.args)
			}
		})
	}
}
