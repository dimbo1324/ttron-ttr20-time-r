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

	got, err := ParseReadTimeResponse(payload)
	if err != nil {
		t.Fatalf("ParseReadTimeResponse() error = %v", err)
	}
	if got.Raw != wantRaw || !got.Time.Equal(ts) {
		t.Fatalf("ParseReadTimeResponse() = %+v", got)
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
