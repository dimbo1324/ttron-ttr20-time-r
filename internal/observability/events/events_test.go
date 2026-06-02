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
