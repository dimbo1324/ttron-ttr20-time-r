package config

import (
	"testing"
	"time"
)

func TestLoadEmulatorDefaults(t *testing.T) {
	cfg, err := LoadEmulator(nil)
	if err != nil {
		t.Fatalf("LoadEmulator defaults failed: %v", err)
	}
	if cfg.ListenAddress() != "127.0.0.1:9000" || cfg.CRCMode != "sum" || cfg.AdapterAddr != 1 {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if cfg.LogFile != "runtime/logs/ft12-emulator.log" {
		t.Fatalf("LogFile = %q", cfg.LogFile)
	}
}

func TestLoadEmulatorAliasesAndNormalization(t *testing.T) {
	cfg, err := LoadEmulator([]string{"-mode", "crc16", "-delay", "25", "-readtimeout", "7", "-bad-checksum", "-fragment-response"})
	if err != nil {
		t.Fatalf("LoadEmulator aliases failed: %v", err)
	}
	if cfg.CRCMode != "crc16" {
		t.Fatalf("CRCMode = %q, want crc16", cfg.CRCMode)
	}
	if cfg.ResponseDelay != 25*time.Millisecond {
		t.Fatalf("ResponseDelay = %s", cfg.ResponseDelay)
	}
	if cfg.ReadTimeoutDuration != 7*time.Second {
		t.Fatalf("ReadTimeoutDuration = %s", cfg.ReadTimeoutDuration)
	}
	if cfg.BadCRCProb != 1 || cfg.FragProb != 1 {
		t.Fatalf("fault aliases not applied: %+v", cfg)
	}
}

func TestLoadEmulatorValidation(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "bad listen", args: []string{"-listen", "bad"}},
		{name: "bad adapter", args: []string{"-adapter", "-1"}},
		{name: "bad checksum", args: []string{"-crc", "bad"}},
		{name: "bad badcrc", args: []string{"-badcrc", "1.1"}},
		{name: "bad fragment", args: []string{"-fragment", "-0.1"}},
		{name: "bad recent", args: []string{"-recent", "0"}},
		{name: "bad read timeout", args: []string{"-read-timeout", "0s"}},
		{name: "bad write timeout", args: []string{"-write-timeout", "0s"}},
		{name: "bad grpc", args: []string{"-grpc-listen", "bad"}},
		{name: "bad response delay", args: []string{"-response-delay", "-1s"}},
		{name: "bad fragment delay", args: []string{"-fragment-delay", "-1s"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := LoadEmulator(tt.args); err == nil {
				t.Fatalf("LoadEmulator(%v) succeeded, want error", tt.args)
			}
		})
	}
}
