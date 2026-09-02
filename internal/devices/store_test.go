package devices

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeInventory(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "devices.json")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadInventory(t *testing.T) {
	path := writeInventory(t, `{
      "devices": [
        {
          "id": "tekon-01",
          "name": "Feeder 1",
          "target": "127.0.0.1:9000",
          "checksumMode": "sum",
          "adapterAddr": 1,
          "scheduleMode": "aligned",
          "pollInterval": "1m0s",
          "pollOffset": "5s",
          "requestTimeout": "1.5s",
          "connectTimeout": "2s",
          "clockWarn": "2s",
          "clockCritical": "30s",
          "degradeAfter": 3,
          "offlineAfter": 10,
          "recoverAfter": 2,
          "enabled": true
        }
      ]
    }`)

	registry, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	device, err := registry.Get("tekon-01")
	if err != nil {
		t.Fatal(err)
	}
	if device.PollInterval != time.Minute || device.PollOffset != 5*time.Second {
		t.Fatalf("schedule = (%s, %s)", device.PollInterval, device.PollOffset)
	}
	if device.RequestTimeout != 1500*time.Millisecond {
		t.Fatalf("RequestTimeout = %s", device.RequestTimeout)
	}
	if device.ClockCritical != 30*time.Second || !device.Enabled {
		t.Fatalf("device = %+v", device)
	}
}

func TestLoadAppliesDefaultsForOmittedFields(t *testing.T) {
	path := writeInventory(t, `{"devices":[{"id":"a","target":"127.0.0.1:9000","enabled":true}]}`)

	registry, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	device, err := registry.Get("a")
	if err != nil {
		t.Fatal(err)
	}
	if device.PollInterval != DefaultPollInterval || device.ChecksumMode != "sum" {
		t.Fatalf("device = %+v", device)
	}
}

func TestLoadRejectsBadInput(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "malformed json", body: `{"devices":`},
		{name: "unknown field", body: `{"devices":[{"id":"a","target":"127.0.0.1:1","colour":"red"}]}`},
		{name: "bad duration", body: `{"devices":[{"id":"a","target":"127.0.0.1:1","pollInterval":"5 seconds"}]}`},
		{name: "negative duration", body: `{"devices":[{"id":"a","target":"127.0.0.1:1","pollInterval":"-5s"}]}`},
		{name: "invalid id", body: `{"devices":[{"id":"A B","target":"127.0.0.1:1"}]}`},
		{name: "invalid target", body: `{"devices":[{"id":"a","target":"nope"}]}`},
		{name: "duplicate ids", body: `{"devices":[{"id":"a","target":"127.0.0.1:1"},{"id":"a","target":"127.0.0.1:2"}]}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeInventory(t, tt.body)
			if _, err := Load(path); err == nil {
				t.Fatal("Load() = nil error, want a failure")
			}
		})
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "absent.json")); err == nil {
		t.Fatal("Load() must fail for a missing file")
	}
}

func TestLoadOrEmpty(t *testing.T) {
	registry, err := LoadOrEmpty("")
	if err != nil {
		t.Fatal(err)
	}
	if registry.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", registry.Len())
	}

	registry, err = LoadOrEmpty(filepath.Join(t.TempDir(), "absent.json"))
	if err != nil {
		t.Fatalf("LoadOrEmpty() must tolerate a missing file: %v", err)
	}
	if registry.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", registry.Len())
	}

	path := writeInventory(t, `{"devices":[{"id":"a","target":"127.0.0.1:9000","enabled":true}]}`)
	registry, err = LoadOrEmpty(path)
	if err != nil {
		t.Fatal(err)
	}
	if registry.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", registry.Len())
	}
}

func TestSaveAndReload(t *testing.T) {
	registry := NewRegistry()
	device := validDevice()
	device.ScheduleMode = "aligned"
	device.PollInterval = time.Minute
	device.PollOffset = 5 * time.Second
	if err := registry.Add(device); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "nested", "devices.json")
	if err := Save(path, registry); err != nil {
		t.Fatal(err)
	}

	reloaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	got, err := reloaded.Get("tekon-01")
	if err != nil {
		t.Fatal(err)
	}
	want, err := registry.Get("tekon-01")
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("round trip changed the device:\n got = %+v\nwant = %+v", got, want)
	}
}

func TestSaveRejectsEmptyPath(t *testing.T) {
	if err := Save("", NewRegistry()); err == nil {
		t.Fatal("Save() must reject an empty path")
	}
}

func TestSaveLeavesNoTemporaryFile(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "devices.json")
	registry := NewRegistry()
	if err := registry.Add(validDevice()); err != nil {
		t.Fatal(err)
	}
	if err := Save(path, registry); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "devices.json" {
		t.Fatalf("directory contains %d entries, want only the inventory", len(entries))
	}
}

func TestSaveOverwritesExistingInventory(t *testing.T) {
	path := writeInventory(t, `{"devices":[{"id":"old","target":"127.0.0.1:9000","enabled":true}]}`)

	registry := NewRegistry()
	if err := registry.Add(validDevice()); err != nil {
		t.Fatal(err)
	}
	if err := Save(path, registry); err != nil {
		t.Fatal(err)
	}

	reloaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := reloaded.Get("old"); !errors.Is(err, ErrNotFound) {
		t.Fatal("Save() must replace the previous inventory")
	}
}

func TestParseOptionalDuration(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    time.Duration
		wantErr bool
	}{
		{name: "empty", value: "", want: 0},
		{name: "seconds", value: "5s", want: 5 * time.Second},
		{name: "fractional", value: "1.5s", want: 1500 * time.Millisecond},
		{name: "invalid", value: "soon", wantErr: true},
		{name: "negative", value: "-1s", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOptionalDuration(tt.value, "field")
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseOptionalDuration(%q) error = %v, wantErr %t", tt.value, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("parseOptionalDuration(%q) = %s, want %s", tt.value, got, tt.want)
			}
		})
	}
}

func TestShippedExampleInventoryLoads(t *testing.T) {
	registry, err := Load(filepath.Join("..", "..", "examples", "devices.example.json"))
	if err != nil {
		t.Fatalf("the shipped example inventory must stay loadable: %v", err)
	}
	if registry.Len() == 0 {
		t.Fatal("the example inventory must declare devices")
	}
	for _, device := range registry.List() {
		if err := device.Validate(); err != nil {
			t.Fatalf("example device %q is invalid: %v", device.ID, err)
		}
		if _, err := device.Schedule(); err != nil {
			t.Fatalf("example device %q has an unusable schedule: %v", device.ID, err)
		}
	}
}
