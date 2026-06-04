package dto

import "testing"

func TestMergeEventsSortsNewestFirstAndLimits(t *testing.T) {
	t1 := "2026-06-04T10:00:00Z"
	t2 := "2026-06-04T10:00:01Z"
	got := MergeEvents(
		[]EventDTO{{ID: 1, Timestamp: &t1, Source: "emulator"}},
		[]EventDTO{{ID: 2, Timestamp: &t2, Source: "gateway"}, {ID: 3, Timestamp: &t1, Source: "gateway"}},
		2,
	)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].ID != 2 || got[1].ID != 3 {
		t.Fatalf("events = %+v", got)
	}
}
