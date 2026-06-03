package events

import "testing"

func TestRingKeepsRecentRecords(t *testing.T) {
	ring := NewRing(2)
	ring.Add(FrameRecord{RawHex: "one"})
	ring.Add(FrameRecord{RawHex: "two"})
	ring.Add(FrameRecord{RawHex: "three"})

	got := ring.Snapshot()
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].RawHex != "two" || got[1].RawHex != "three" {
		t.Fatalf("records = %+v", got)
	}
}

func TestRingAssignsStableMonotonicIDsAfterRotation(t *testing.T) {
	ring := NewRing(2)
	ring.Add(FrameRecord{RawHex: "one"})
	ring.Add(FrameRecord{RawHex: "two"})
	firstSnapshot := ring.Snapshot()
	ring.Add(FrameRecord{RawHex: "three"})

	got := ring.Snapshot()
	if got[0].ID != 2 || got[1].ID != 3 {
		t.Fatalf("ids = %d,%d want 2,3; records=%+v", got[0].ID, got[1].ID, got)
	}
	if firstSnapshot[1].ID != got[0].ID {
		t.Fatalf("record ID changed across snapshots: before=%d after=%d", firstSnapshot[1].ID, got[0].ID)
	}
}

func TestRingSnapshotReturnsCopy(t *testing.T) {
	ring := NewRing(2)
	ring.Add(FrameRecord{RawHex: "one"})
	snapshot := ring.Snapshot()
	snapshot[0].RawHex = "mutated"

	got := ring.Snapshot()
	if got[0].RawHex != "one" {
		t.Fatalf("snapshot mutation affected ring: %+v", got)
	}
}
