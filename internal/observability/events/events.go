package events

import (
	"sync"
	"time"
)

type Direction string

const (
	DirectionRX     Direction = "RX"
	DirectionTX     Direction = "TX"
	DirectionError  Direction = "ERR"
	DirectionDrop   Direction = "DROP"
	DirectionSystem Direction = "SYSTEM"
)

type ServiceName string

const (
	ServiceEmulator ServiceName = "emulator"
	ServiceGateway  ServiceName = "gateway"
)

type FrameRecord struct {
	ID           uint64
	Timestamp    time.Time
	Direction    Direction
	Service      ServiceName
	RemoteAddr   string
	RawHex       string
	Command      string
	ChecksumMode string
	Error        string
	Message      string
}

type Ring struct {
	mu       sync.Mutex
	capacity int
	records  []FrameRecord
	next     int
	nextID   uint64
	full     bool
}

func NewRing(capacity int) *Ring {
	if capacity <= 0 {
		capacity = 100
	}
	return &Ring{capacity: capacity, records: make([]FrameRecord, capacity), nextID: 1}
}

func (r *Ring) Add(record FrameRecord) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now()
	}
	record.ID = r.nextID
	r.nextID++
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
