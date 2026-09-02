package command

import (
	"errors"
	"testing"
	"time"
)

func TestBuildAndParseReadTimeRequest(t *testing.T) {
	req := BuildReadTimeRequest()
	if len(req) != 1 || req[0] != byte(ReadTime) {
		t.Fatalf("BuildReadTimeRequest() = % X", req)
	}
	if err := ParseReadTimeRequest(req); err != nil {
		t.Fatalf("ParseReadTimeRequest() error = %v", err)
	}
}

func TestParseReadTimeRequestRejectsInvalidPayload(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want error
	}{
		{name: "empty", data: nil, want: ErrEmptyPayload},
		{name: "wrong command", data: []byte{0x02}, want: ErrUnexpectedCommand},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ParseReadTimeRequest(tt.data); !errors.Is(err, tt.want) {
				t.Fatalf("ParseReadTimeRequest() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestBuildAndParseReadTimeResponse(t *testing.T) {
	ts := time.Date(2026, 6, 2, 12, 34, 56, 0, time.UTC)
	payload := BuildReadTimeResponse(ts)
	wantRaw := "2026-06-02 12:34:56"
	if string(payload[1:]) != wantRaw {
		t.Fatalf("BuildReadTimeResponse() timestamp = %q, want %q", string(payload[1:]), wantRaw)
	}

	got, err := ParseReadTimeResponseIn(payload, time.UTC)
	if err != nil {
		t.Fatalf("ParseReadTimeResponseIn() error = %v", err)
	}
	if got.Raw != wantRaw || !got.Time.Equal(ts) {
		t.Fatalf("ParseReadTimeResponseIn() = %+v", got)
	}
}

func TestReadTimeRoundTripIsZoneSymmetric(t *testing.T) {
	locations := []*time.Location{
		time.UTC,
		time.Local,
		time.FixedZone("UTC+3", 3*60*60),
		time.FixedZone("UTC-5", -5*60*60),
	}

	for _, location := range locations {
		t.Run(location.String(), func(t *testing.T) {
			ts := time.Date(2026, 6, 2, 12, 34, 56, 0, location)

			got, err := ParseReadTimeResponseIn(BuildReadTimeResponse(ts), location)
			if err != nil {
				t.Fatal(err)
			}
			if !got.Time.Equal(ts) {
				t.Fatalf("round trip in %s produced %s, want %s", location, got.Time, ts)
			}
		})
	}
}

func TestParseReadTimeResponseInterpretsNaiveTimestampInLocation(t *testing.T) {
	payload := append([]byte{byte(ReadTime)}, []byte("2026-06-02 12:34:56")...)
	zone := time.FixedZone("UTC+3", 3*60*60)

	got, err := ParseReadTimeResponseIn(payload, zone)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 6, 2, 12, 34, 56, 0, zone)
	if !got.Time.Equal(want) {
		t.Fatalf("ParseReadTimeResponseIn() = %s, want %s", got.Time, want)
	}

	inUTC, err := ParseReadTimeResponseIn(payload, time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	if inUTC.Time.Equal(got.Time) {
		t.Fatal("the same naive timestamp must denote different instants in different zones")
	}
}

func TestParseReadTimeResponseDefaultsToLocalZone(t *testing.T) {
	payload := append([]byte{byte(ReadTime)}, []byte("2026-06-02 12:34:56")...)

	got, err := ParseReadTimeResponse(payload)
	if err != nil {
		t.Fatal(err)
	}
	want, err := ParseReadTimeResponseIn(payload, time.Local)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Time.Equal(want.Time) {
		t.Fatalf("ParseReadTimeResponse() = %s, want the local interpretation %s", got.Time, want.Time)
	}
}

func TestParseReadTimeResponseInNilLocationFallsBackToLocal(t *testing.T) {
	payload := append([]byte{byte(ReadTime)}, []byte("2026-06-02 12:34:56")...)

	got, err := ParseReadTimeResponseIn(payload, nil)
	if err != nil {
		t.Fatal(err)
	}
	want, err := ParseReadTimeResponseIn(payload, time.Local)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Time.Equal(want.Time) {
		t.Fatalf("ParseReadTimeResponseIn(nil) = %s, want %s", got.Time, want.Time)
	}
}

func TestParseReadTimeResponseRejectsInvalidPayload(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want error
	}{
		{name: "empty", data: nil, want: ErrEmptyPayload},
		{name: "wrong command", data: []byte{0x02}, want: ErrUnexpectedCommand},
		{name: "short timestamp", data: []byte{byte(ReadTime), '1'}, want: ErrInvalidPayload},
		{name: "long timestamp", data: append([]byte{byte(ReadTime)}, []byte("2026-06-02 12:34:56Z")...), want: ErrInvalidPayload},
		{name: "malformed timestamp", data: append([]byte{byte(ReadTime)}, []byte("2026-99-99 99:99:99")...), want: ErrInvalidTime},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseReadTimeResponse(tt.data)
			if !errors.Is(err, tt.want) {
				t.Fatalf("ParseReadTimeResponse() error = %v, want %v", err, tt.want)
			}
		})
	}
}
