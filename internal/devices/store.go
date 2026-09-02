package devices

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type inventoryFile struct {
	Devices []deviceDocument `json:"devices"`
}

type deviceDocument struct {
	ID             string `json:"id"`
	Name           string `json:"name,omitempty"`
	Target         string `json:"target"`
	ChecksumMode   string `json:"checksumMode,omitempty"`
	AdapterAddr    int    `json:"adapterAddr,omitempty"`
	ScheduleMode   string `json:"scheduleMode,omitempty"`
	PollInterval   string `json:"pollInterval,omitempty"`
	PollOffset     string `json:"pollOffset,omitempty"`
	RequestTimeout string `json:"requestTimeout,omitempty"`
	ConnectTimeout string `json:"connectTimeout,omitempty"`
	ClockWarn      string `json:"clockWarn,omitempty"`
	ClockCritical  string `json:"clockCritical,omitempty"`
	DegradeAfter   int    `json:"degradeAfter,omitempty"`
	OfflineAfter   int    `json:"offlineAfter,omitempty"`
	RecoverAfter   int    `json:"recoverAfter,omitempty"`
	Enabled        bool   `json:"enabled"`
}

func Load(path string) (*Registry, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read device inventory: %w", err)
	}

	var file inventoryFile
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&file); err != nil {
		return nil, fmt.Errorf("parse device inventory %s: %w", path, err)
	}

	registry := NewRegistry()
	for index, document := range file.Devices {
		device, err := document.toDevice()
		if err != nil {
			return nil, fmt.Errorf("device inventory entry %d: %w", index, err)
		}
		if err := registry.Add(device); err != nil {
			return nil, fmt.Errorf("device inventory entry %d: %w", index, err)
		}
	}
	return registry, nil
}

func LoadOrEmpty(path string) (*Registry, error) {
	if path == "" {
		return NewRegistry(), nil
	}
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return NewRegistry(), nil
		}
		return nil, fmt.Errorf("read device inventory: %w", err)
	}
	return Load(path)
}

func Save(path string, registry *Registry) error {
	if path == "" {
		return fmt.Errorf("device inventory path must not be empty")
	}
	devices := registry.List()
	file := inventoryFile{Devices: make([]deviceDocument, 0, len(devices))}
	for _, device := range devices {
		file.Devices = append(file.Devices, newDeviceDocument(device))
	}

	payload, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return fmt.Errorf("encode device inventory: %w", err)
	}
	payload = append(payload, '\n')

	if directory := filepath.Dir(path); directory != "" && directory != "." {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			return fmt.Errorf("create device inventory directory: %w", err)
		}
	}

	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, payload, 0o644); err != nil {
		return fmt.Errorf("write device inventory: %w", err)
	}
	if err := os.Rename(temporary, path); err != nil {
		_ = os.Remove(temporary)
		return fmt.Errorf("replace device inventory: %w", err)
	}
	return nil
}

func (d deviceDocument) toDevice() (Device, error) {
	device := Device{
		ID:           d.ID,
		Name:         d.Name,
		Target:       d.Target,
		ChecksumMode: d.ChecksumMode,
		AdapterAddr:  d.AdapterAddr,
		ScheduleMode: d.ScheduleMode,
		DegradeAfter: d.DegradeAfter,
		OfflineAfter: d.OfflineAfter,
		RecoverAfter: d.RecoverAfter,
		Enabled:      d.Enabled,
	}

	fields := []struct {
		name   string
		value  string
		target *time.Duration
	}{
		{"pollInterval", d.PollInterval, &device.PollInterval},
		{"pollOffset", d.PollOffset, &device.PollOffset},
		{"requestTimeout", d.RequestTimeout, &device.RequestTimeout},
		{"connectTimeout", d.ConnectTimeout, &device.ConnectTimeout},
		{"clockWarn", d.ClockWarn, &device.ClockWarn},
		{"clockCritical", d.ClockCritical, &device.ClockCritical},
	}
	for _, field := range fields {
		parsed, err := parseOptionalDuration(field.value, field.name)
		if err != nil {
			return Device{}, err
		}
		*field.target = parsed
	}
	return device, nil
}

func newDeviceDocument(device Device) deviceDocument {
	return deviceDocument{
		ID:             device.ID,
		Name:           device.Name,
		Target:         device.Target,
		ChecksumMode:   device.ChecksumMode,
		AdapterAddr:    device.AdapterAddr,
		ScheduleMode:   device.ScheduleMode,
		PollInterval:   device.PollInterval.String(),
		PollOffset:     device.PollOffset.String(),
		RequestTimeout: device.RequestTimeout.String(),
		ConnectTimeout: device.ConnectTimeout.String(),
		ClockWarn:      device.ClockWarn.String(),
		ClockCritical:  device.ClockCritical.String(),
		DegradeAfter:   device.DegradeAfter,
		OfflineAfter:   device.OfflineAfter,
		RecoverAfter:   device.RecoverAfter,
		Enabled:        device.Enabled,
	}
}

func parseOptionalDuration(value, name string) (time.Duration, error) {
	if value == "" {
		return 0, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("field %s: %w", name, err)
	}
	if parsed < 0 {
		return 0, fmt.Errorf("field %s must not be negative", name)
	}
	return parsed, nil
}
