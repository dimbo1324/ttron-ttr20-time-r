package command

import (
	"errors"
	"sync"
	"testing"
)

func TestDefaultRegistryKnowsBuiltInCommands(t *testing.T) {
	registry := DefaultRegistry()

	tests := []struct {
		id   ID
		name string
	}{
		{id: ReadTime, name: NameReadTime},
		{id: ReadIdentity, name: NameReadIdentity},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			descriptor, ok := registry.Lookup(tt.id)
			if !ok {
				t.Fatalf("Lookup(0x%02X) not found", byte(tt.id))
			}
			if descriptor.Name != tt.name {
				t.Fatalf("Name = %q, want %q", descriptor.Name, tt.name)
			}
			if descriptor.Description == "" {
				t.Fatal("built-in commands must carry a description")
			}
			if !registry.Supports(tt.id) {
				t.Fatalf("Supports(0x%02X) = false", byte(tt.id))
			}
			if got := registry.Name(tt.id); got != tt.name {
				t.Fatalf("Name(0x%02X) = %q, want %q", byte(tt.id), got, tt.name)
			}
			byName, ok := registry.LookupName(tt.name)
			if !ok || byName.ID != tt.id {
				t.Fatalf("LookupName(%q) = (%+v, %t)", tt.name, byName, ok)
			}
		})
	}
}

func TestRegistryNameForUnknownCommand(t *testing.T) {
	registry := DefaultRegistry()

	if got := registry.Name(ID(0x7F)); got != "unknown-0x7F" {
		t.Fatalf("Name() = %q, want %q", got, "unknown-0x7F")
	}
	if registry.Supports(ID(0x7F)) {
		t.Fatal("Supports() must be false for an unregistered command")
	}
	if _, ok := registry.LookupName("absent"); ok {
		t.Fatal("LookupName() must not resolve an unknown name")
	}
}

func TestRegistryRegisterRejectsDuplicates(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Register(Descriptor{ID: ID(0x10), Name: "custom"}); err != nil {
		t.Fatal(err)
	}

	err := registry.Register(Descriptor{ID: ID(0x10), Name: "other"})
	if !errors.Is(err, ErrDuplicateCommand) {
		t.Fatalf("Register() error = %v, want %v", err, ErrDuplicateCommand)
	}

	err = registry.Register(Descriptor{ID: ID(0x11), Name: "custom"})
	if !errors.Is(err, ErrDuplicateCommand) {
		t.Fatalf("Register() error = %v, want %v", err, ErrDuplicateCommand)
	}
}

func TestRegistryRegisterRejectsEmptyName(t *testing.T) {
	registry := NewRegistry()
	err := registry.Register(Descriptor{ID: ID(0x20)})
	if !errors.Is(err, ErrInvalidDescriptor) {
		t.Fatalf("Register() error = %v, want %v", err, ErrInvalidDescriptor)
	}
}

func TestRegistryMustRegisterPanicsOnDuplicate(t *testing.T) {
	registry := NewRegistry()
	registry.MustRegister(Descriptor{ID: ID(0x30), Name: "first"})

	defer func() {
		if recover() == nil {
			t.Fatal("MustRegister() must panic on a duplicate")
		}
	}()
	registry.MustRegister(Descriptor{ID: ID(0x30), Name: "second"})
}

func TestRegistryDescriptorsAreSortedByID(t *testing.T) {
	registry := NewRegistry()
	for _, descriptor := range []Descriptor{
		{ID: ID(0x30), Name: "c"},
		{ID: ID(0x10), Name: "a"},
		{ID: ID(0x20), Name: "b"},
	} {
		if err := registry.Register(descriptor); err != nil {
			t.Fatal(err)
		}
	}

	descriptors := registry.Descriptors()
	want := []ID{0x10, 0x20, 0x30}
	if len(descriptors) != len(want) {
		t.Fatalf("Descriptors() = %d entries, want %d", len(descriptors), len(want))
	}
	for i, descriptor := range descriptors {
		if descriptor.ID != want[i] {
			t.Fatalf("Descriptors()[%d].ID = 0x%02X, want 0x%02X", i, byte(descriptor.ID), byte(want[i]))
		}
	}
}

func TestRegistryConcurrentUse(t *testing.T) {
	registry := DefaultRegistry()
	var wg sync.WaitGroup

	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			if err := registry.Register(Descriptor{ID: ID(0x40 + index), Name: string(rune('a' + index))}); err != nil {
				t.Errorf("Register() error = %v", err)
				return
			}
			_ = registry.Descriptors()
			_ = registry.Name(ReadTime)
			_ = registry.Supports(ReadIdentity)
		}(i)
	}
	wg.Wait()

	if got := len(registry.Descriptors()); got != 18 {
		t.Fatalf("Descriptors() = %d entries, want 18", got)
	}
}
