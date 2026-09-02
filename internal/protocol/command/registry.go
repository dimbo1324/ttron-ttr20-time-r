package command

import (
	"fmt"
	"sort"
	"sync"
)

type Descriptor struct {
	ID          ID
	Name        string
	Description string
}

type Registry struct {
	mu     sync.RWMutex
	byID   map[ID]Descriptor
	byName map[string]Descriptor
}

func NewRegistry() *Registry {
	return &Registry{
		byID:   make(map[ID]Descriptor),
		byName: make(map[string]Descriptor),
	}
}

func (r *Registry) Register(d Descriptor) error {
	if d.Name == "" {
		return fmt.Errorf("%w: command 0x%02X has empty name", ErrInvalidDescriptor, byte(d.ID))
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if existing, ok := r.byID[d.ID]; ok {
		return fmt.Errorf("%w: 0x%02X already registered as %q", ErrDuplicateCommand, byte(d.ID), existing.Name)
	}
	if existing, ok := r.byName[d.Name]; ok {
		return fmt.Errorf("%w: name %q already used by 0x%02X", ErrDuplicateCommand, d.Name, byte(existing.ID))
	}
	r.byID[d.ID] = d
	r.byName[d.Name] = d
	return nil
}

func (r *Registry) MustRegister(d Descriptor) {
	if err := r.Register(d); err != nil {
		panic(err)
	}
}

func (r *Registry) Lookup(id ID) (Descriptor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.byID[id]
	return d, ok
}

func (r *Registry) LookupName(name string) (Descriptor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.byName[name]
	return d, ok
}

func (r *Registry) Name(id ID) string {
	if d, ok := r.Lookup(id); ok {
		return d.Name
	}
	return fmt.Sprintf("unknown-0x%02X", byte(id))
}

func (r *Registry) Descriptors() []Descriptor {
	r.mu.RLock()
	out := make([]Descriptor, 0, len(r.byID))
	for _, d := range r.byID {
		out = append(out, d)
	}
	r.mu.RUnlock()
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (r *Registry) Supports(id ID) bool {
	_, ok := r.Lookup(id)
	return ok
}

func DefaultRegistry() *Registry {
	r := NewRegistry()
	r.MustRegister(Descriptor{ID: ReadTime, Name: NameReadTime, Description: "read device date and time"})
	r.MustRegister(Descriptor{ID: ReadIdentity, Name: NameReadIdentity, Description: "read device model, serial and firmware"})
	return r
}
