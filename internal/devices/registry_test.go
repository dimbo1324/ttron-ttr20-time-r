package devices

import (
	"errors"
	"sync"
	"testing"
)

func TestRegistryAddAndGet(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Add(validDevice()); err != nil {
		t.Fatal(err)
	}
	if got := registry.Len(); got != 1 {
		t.Fatalf("Len() = %d, want 1", got)
	}

	device, err := registry.Get("TEKON-01")
	if err != nil {
		t.Fatalf("Get() must accept a non-canonical id: %v", err)
	}
	if device.ID != "tekon-01" {
		t.Fatalf("Get().ID = %q", device.ID)
	}
}

func TestRegistryAddRejectsDuplicates(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Add(validDevice()); err != nil {
		t.Fatal(err)
	}
	if err := registry.Add(validDevice()); !errors.Is(err, ErrDuplicateID) {
		t.Fatalf("Add() error = %v, want %v", err, ErrDuplicateID)
	}
}

func TestRegistryAddRejectsInvalidDevice(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Add(Device{ID: "bad id", Target: "127.0.0.1:1"}); !errors.Is(err, ErrInvalidID) {
		t.Fatalf("Add() error = %v, want %v", err, ErrInvalidID)
	}
	if registry.Len() != 0 {
		t.Fatal("an invalid device must not be stored")
	}
}

func TestRegistryGetMissing(t *testing.T) {
	registry := NewRegistry()
	if _, err := registry.Get("absent"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get() error = %v, want %v", err, ErrNotFound)
	}
}

func TestRegistryUpsertReplaces(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Add(validDevice()); err != nil {
		t.Fatal(err)
	}

	updated := validDevice()
	updated.Target = "10.0.0.9:9100"
	stored, err := registry.Upsert(updated)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Target != "10.0.0.9:9100" {
		t.Fatalf("Upsert() = %+v", stored)
	}
	if registry.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", registry.Len())
	}

	if _, err := registry.Upsert(Device{ID: "", Target: "127.0.0.1:1"}); !errors.Is(err, ErrInvalidID) {
		t.Fatalf("Upsert() must validate: %v", err)
	}
}

func TestRegistryRemove(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Add(validDevice()); err != nil {
		t.Fatal(err)
	}
	if err := registry.Remove("tekon-01"); err != nil {
		t.Fatal(err)
	}
	if registry.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", registry.Len())
	}
	if err := registry.Remove("tekon-01"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Remove() error = %v, want %v", err, ErrNotFound)
	}
}

func TestRegistrySetEnabled(t *testing.T) {
	registry := NewRegistry()
	device := validDevice()
	device.Enabled = false
	if err := registry.Add(device); err != nil {
		t.Fatal(err)
	}

	updated, err := registry.SetEnabled("tekon-01", true)
	if err != nil {
		t.Fatal(err)
	}
	if !updated.Enabled {
		t.Fatal("SetEnabled(true) must enable the device")
	}
	if len(registry.Enabled()) != 1 {
		t.Fatalf("Enabled() = %+v", registry.Enabled())
	}

	if _, err := registry.SetEnabled("absent", true); !errors.Is(err, ErrNotFound) {
		t.Fatalf("SetEnabled() error = %v, want %v", err, ErrNotFound)
	}
}

func TestRegistryListIsSorted(t *testing.T) {
	registry := NewRegistry()
	for _, id := range []string{"zeta", "alpha", "mid"} {
		device := validDevice()
		device.ID = id
		if err := registry.Add(device); err != nil {
			t.Fatal(err)
		}
	}

	list := registry.List()
	want := []string{"alpha", "mid", "zeta"}
	for i, device := range list {
		if device.ID != want[i] {
			t.Fatalf("List()[%d].ID = %q, want %q", i, device.ID, want[i])
		}
	}
}

func TestRegistryEnabledFiltersDisabled(t *testing.T) {
	registry := NewRegistry()
	for i, enabled := range []bool{true, false, true} {
		device := validDevice()
		device.ID = string(rune('a' + i))
		device.Enabled = enabled
		if err := registry.Add(device); err != nil {
			t.Fatal(err)
		}
	}

	enabled := registry.Enabled()
	if len(enabled) != 2 {
		t.Fatalf("Enabled() = %d devices, want 2", len(enabled))
	}
	for _, device := range enabled {
		if !device.Enabled {
			t.Fatalf("Enabled() returned a disabled device: %+v", device)
		}
	}
}

func TestFromSlice(t *testing.T) {
	first := validDevice()
	second := validDevice()
	second.ID = "tekon-02"

	registry, err := FromSlice([]Device{first, second})
	if err != nil {
		t.Fatal(err)
	}
	if registry.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", registry.Len())
	}

	if _, err := FromSlice([]Device{first, first}); !errors.Is(err, ErrDuplicateID) {
		t.Fatalf("FromSlice() error = %v, want %v", err, ErrDuplicateID)
	}
}

func TestRegistryConcurrentAccess(t *testing.T) {
	registry := NewRegistry()
	var wg sync.WaitGroup

	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			device := validDevice()
			device.ID = "device-" + string(rune('a'+index))
			if err := registry.Add(device); err != nil {
				t.Errorf("Add() error = %v", err)
				return
			}
			_ = registry.List()
			_ = registry.Enabled()
			_, _ = registry.Get(device.ID)
		}(i)
	}
	wg.Wait()

	if got := registry.Len(); got != 16 {
		t.Fatalf("Len() = %d, want 16", got)
	}
}
