package config

import (
	"strings"
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
	// With no flags at all the gateway does the thing this project was asked
	// for: the fifth second of every minute.
	if cfg.ScheduleMode != "aligned" || cfg.PollInterval != time.Minute || cfg.PollOffset != 5*time.Second {
		t.Fatalf("default schedule = %s/%s/+%s", cfg.ScheduleMode, cfg.PollInterval, cfg.PollOffset)
	}
}

func TestLoadGatewayModeAlias(t *testing.T) {
	cfg, err := LoadGateway([]string{"-mode", "crc16", "-schedule", "interval", "-interval", "1s"})
	if err != nil {
		t.Fatalf("LoadGateway alias failed: %v", err)
	}
	if cfg.CRCMode != "crc16" || cfg.PollInterval != time.Second {
		t.Fatalf("alias/defaults not applied: %+v", cfg)
	}
}

func TestLoadGatewayExplainsAnOffsetThatNoLongerFits(t *testing.T) {
	// The default schedule is aligned at +5s, so passing -interval on its own
	// is both the first thing anyone does and the first way to make the pair
	// invalid. The message has to name the two flags involved, or the reader
	// goes hunting for a knob they never touched.
	_, err := LoadGateway([]string{"-interval", "1s"})
	if err == nil {
		t.Fatal("LoadGateway accepted an offset outside its interval")
	}
	for _, want := range []string{"-poll-offset", "-interval", "-schedule interval"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error = %q, want it to mention %q", err, want)
		}
	}
}

func TestLoadGatewayAcceptsAFixedRate(t *testing.T) {
	cfg, err := LoadGateway([]string{"-schedule", "interval", "-interval", "1s"})
	if err != nil {
		t.Fatal(err)
	}
	// An offset means nothing to a fixed rate, so the default one is simply
	// ignored rather than being a reason to refuse.
	if cfg.PollInterval != time.Second || cfg.ScheduleMode != "interval" {
		t.Fatalf("cfg = %+v", cfg)
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
