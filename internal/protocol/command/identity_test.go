package command

import (
	"errors"
	"strings"
	"testing"
)

func TestBuildReadIdentityRequest(t *testing.T) {
	got := BuildReadIdentityRequest()
	if len(got) != 1 || got[0] != byte(ReadIdentity) {
		t.Fatalf("BuildReadIdentityRequest() = % X", got)
	}
	if err := ParseReadIdentityRequest(got); err != nil {
		t.Fatalf("ParseReadIdentityRequest() = %v", err)
	}
}

func TestParseReadIdentityRequestRejectsOtherCommands(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr error
	}{
		{name: "empty", data: nil, wantErr: ErrEmptyPayload},
		{name: "read-time", data: []byte{byte(ReadTime)}, wantErr: ErrUnexpectedCommand},
		{name: "unknown", data: []byte{0x7F}, wantErr: ErrUnexpectedCommand},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ParseReadIdentityRequest(tt.data); !errors.Is(err, tt.wantErr) {
				t.Fatalf("ParseReadIdentityRequest() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestIdentityRoundTrip(t *testing.T) {
	want := Identity{Model: "TTR20", Serial: "SN-0000001", Firmware: "1.0.0"}

	payload := BuildReadIdentityResponse(want)
	if payload[0] != byte(ReadIdentity) {
		t.Fatalf("payload command byte = 0x%02X", payload[0])
	}

	got, err := ParseReadIdentityResponse(payload)
	if err != nil {
		t.Fatal(err)
	}
	if got.Model != want.Model || got.Serial != want.Serial || got.Firmware != want.Firmware {
		t.Fatalf("ParseReadIdentityResponse() = %+v, want %+v", got, want)
	}
	if got.Raw != want.String() {
		t.Fatalf("Raw = %q, want %q", got.Raw, want.String())
	}
}

func TestIdentityString(t *testing.T) {
	identity := Identity{Model: "TTR20", Serial: "SN-1", Firmware: "2.0"}
	if got := identity.String(); got != "TTR20|SN-1|2.0" {
		t.Fatalf("String() = %q", got)
	}
}

func TestParseReadIdentityResponseRejectsBadPayloads(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr error
	}{
		{name: "empty", data: nil, wantErr: ErrEmptyPayload},
		{name: "wrong command", data: []byte{byte(ReadTime)}, wantErr: ErrUnexpectedCommand},
		{name: "empty body", data: []byte{byte(ReadIdentity)}, wantErr: ErrInvalidPayload},
		{
			name:    "too long",
			data:    append([]byte{byte(ReadIdentity)}, []byte(strings.Repeat("x", MaxIdentityLength+1))...),
			wantErr: ErrInvalidPayload,
		},
		{
			name:    "too few fields",
			data:    append([]byte{byte(ReadIdentity)}, []byte("TTR20|SN-1")...),
			wantErr: ErrInvalidIdentity,
		},
		{
			name:    "too many fields",
			data:    append([]byte{byte(ReadIdentity)}, []byte("a|b|c|d")...),
			wantErr: ErrInvalidIdentity,
		},
		{
			name:    "blank field",
			data:    append([]byte{byte(ReadIdentity)}, []byte("TTR20|   |1.0")...),
			wantErr: ErrInvalidIdentity,
		},
		{
			name:    "empty leading field",
			data:    append([]byte{byte(ReadIdentity)}, []byte("|SN-1|1.0")...),
			wantErr: ErrInvalidIdentity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseReadIdentityResponse(tt.data)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ParseReadIdentityResponse() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseReadIdentityResponseAcceptsMaximumLength(t *testing.T) {
	body := strings.Repeat("a", 30) + "|" + strings.Repeat("b", 30) + "|" + strings.Repeat("c", 30)
	if len(body) > MaxIdentityLength {
		t.Fatalf("test fixture is %d bytes, above the %d byte limit", len(body), MaxIdentityLength)
	}

	data := append([]byte{byte(ReadIdentity)}, []byte(body)...)
	if _, err := ParseReadIdentityResponse(data); err != nil {
		t.Fatalf("ParseReadIdentityResponse() = %v", err)
	}
}
