package events

import (
	"sync"
	"time"
)

type FrameRecord struct {
	Timestamp    time.Time
	Direction    string
	Service      string
	RemoteAddr   string
	RawHex       string
	Command      string
	ChecksumMode string
	Error        string
}

type Ring struct {
	mu       sync.Mutex
	capacity int
	records  []FrameRecord
	next     int
	full     bool
}

func NewRing(capacity int) *Ring {
	if capacity <= 0 {
		capacity = 100
	}
	return &Ring{capacity: capacity, records: make([]FrameRecord, capacity)}
}

func (r *Ring) Add(record FrameRecord) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now()
	}
	r.records[r.next] = record
	r.next = (r.next + 1) % r.capacity
	if r.next == 0 {
		r.full = true
	}
}

func (r *Ring) Snapshot() []FrameRecord {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.full {
		return append([]FrameRecord(nil), r.records[:r.next]...)
	}

	out := make([]FrameRecord, 0, r.capacity)
	out = append(out, r.records[r.next:]...)
	out = append(out, r.records[:r.next]...)
	return out
}

func (r *Ring) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.full {
		return r.capacity
	}
	return r.next
}
