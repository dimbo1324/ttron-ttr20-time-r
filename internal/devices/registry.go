package devices

import (
	"fmt"
	"sort"
	"sync"
)

type Registry struct {
	mu      sync.RWMutex
	devices map[string]Device
}

func NewRegistry() *Registry {
	return &Registry{devices: make(map[string]Device)}
}

func FromSlice(items []Device) (*Registry, error) {
	r := NewRegistry()
	for _, item := range items {
		if err := r.Add(item); err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (r *Registry) Add(device Device) error {
	normalized := device.Normalize()
	if err := normalized.Validate(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.devices[normalized.ID]; exists {
		return fmt.Errorf("%w: %q", ErrDuplicateID, normalized.ID)
	}
	r.devices[normalized.ID] = normalized
	return nil
}

func (r *Registry) Upsert(device Device) (Device, error) {
	normalized := device.Normalize()
	if err := normalized.Validate(); err != nil {
		return Device{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.devices[normalized.ID] = normalized
	return normalized, nil
}

func (r *Registry) Get(id string) (Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	device, ok := r.devices[normalizeID(id)]
	if !ok {
		return Device{}, fmt.Errorf("%w: %q", ErrNotFound, id)
	}
	return device, nil
}

func (r *Registry) Remove(id string) error {
	key := normalizeID(id)
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.devices[key]; !ok {
		return fmt.Errorf("%w: %q", ErrNotFound, id)
	}
	delete(r.devices, key)
	return nil
}

func (r *Registry) SetEnabled(id string, enabled bool) (Device, error) {
	key := normalizeID(id)
	r.mu.Lock()
	defer r.mu.Unlock()
	device, ok := r.devices[key]
	if !ok {
		return Device{}, fmt.Errorf("%w: %q", ErrNotFound, id)
	}
	device.Enabled = enabled
	r.devices[key] = device
	return device, nil
}

func (r *Registry) List() []Device {
	r.mu.RLock()
	out := make([]Device, 0, len(r.devices))
	for _, device := range r.devices {
		out = append(out, device)
	}
	r.mu.RUnlock()
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (r *Registry) Enabled() []Device {
	all := r.List()
	out := make([]Device, 0, len(all))
	for _, device := range all {
		if device.Enabled {
			out = append(out, device)
		}
	}
	return out
}

func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.devices)
}

func normalizeID(id string) string {
	return Device{ID: id}.Normalize().ID
}
