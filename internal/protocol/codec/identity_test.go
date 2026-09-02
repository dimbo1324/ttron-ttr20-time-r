package codec

import (
	"errors"
	"testing"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/checksum"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/command"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/frame"
)

func allModes() []checksum.Mode {
	return []checksum.Mode{checksum.ModeSum, checksum.ModeCRC16}
}

func TestEncodeDecodeReadIdentityRequest(t *testing.T) {
	for _, mode := range allModes() {
		t.Run(string(mode), func(t *testing.T) {
			wire := New(mode, 0x00, 0x01)

			raw, err := wire.EncodeReadIdentityRequest()
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := wire.DecodeReadIdentityRequest(raw)
			if err != nil {
				t.Fatal(err)
			}
			if decoded.Control != 0x00 || decoded.Address != 0x01 {
				t.Fatalf("decoded frame = %+v", decoded)
			}
			if got := decoded.DataBytes(); len(got) != 1 || got[0] != byte(command.ReadIdentity) {
				t.Fatalf("payload = % X", got)
			}
		})
	}
}

func TestEncodeDecodeReadIdentityResponse(t *testing.T) {
	want := command.Identity{Model: "TTR20", Serial: "SN-0000042", Firmware: "1.2.3"}

	for _, mode := range allModes() {
		t.Run(string(mode), func(t *testing.T) {
			wire := New(mode, 0x00, 0x01)

			raw, err := wire.EncodeReadIdentityRequest()
			if err != nil {
				t.Fatal(err)
			}
			request, err := wire.DecodeReadIdentityRequest(raw)
			if err != nil {
				t.Fatal(err)
			}

			response, err := wire.EncodeReadIdentityResponse(request, want)
			if err != nil {
				t.Fatal(err)
			}
			decoded, identity, err := wire.DecodeReadIdentityResponse(response)
			if err != nil {
				t.Fatal(err)
			}
			if decoded.Control != request.Control|0x80 {
				t.Fatalf("response control = 0x%02X, want the request bit set", decoded.Control)
			}
			if identity.Model != want.Model || identity.Serial != want.Serial || identity.Firmware != want.Firmware {
				t.Fatalf("identity = %+v, want %+v", identity, want)
			}
		})
	}
}

func TestDecodeReadIdentityRejectsCorruptFrames(t *testing.T) {
	wire := New(checksum.ModeSum, 0x00, 0x01)
	raw, err := wire.EncodeReadIdentityRequest()
	if err != nil {
		t.Fatal(err)
	}

	corrupt := append([]byte(nil), raw...)
	corrupt[len(corrupt)-2] ^= 0xFF
	if _, err := wire.DecodeReadIdentityRequest(corrupt); !errors.Is(err, frame.ErrInvalidChecksum) {
		t.Fatalf("DecodeReadIdentityRequest() = %v, want %v", err, frame.ErrInvalidChecksum)
	}
	if _, _, err := wire.DecodeReadIdentityResponse(corrupt); !errors.Is(err, frame.ErrInvalidChecksum) {
		t.Fatalf("DecodeReadIdentityResponse() = %v, want %v", err, frame.ErrInvalidChecksum)
	}
}

func TestDecodeReadIdentityRejectsWrongCommand(t *testing.T) {
	wire := New(checksum.ModeSum, 0x00, 0x01)

	timeRequest, err := wire.EncodeReadTimeRequest()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := wire.DecodeReadIdentityRequest(timeRequest); !errors.Is(err, command.ErrUnexpectedCommand) {
		t.Fatalf("DecodeReadIdentityRequest() = %v, want %v", err, command.ErrUnexpectedCommand)
	}

	request, err := wire.DecodeReadTimeRequest(timeRequest)
	if err != nil {
		t.Fatal(err)
	}
	timeResponse, err := wire.EncodeReadTimeResponse(request, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := wire.DecodeReadIdentityResponse(timeResponse); !errors.Is(err, command.ErrUnexpectedCommand) {
		t.Fatalf("DecodeReadIdentityResponse() = %v, want %v", err, command.ErrUnexpectedCommand)
	}
}

func TestEncodeACK(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantCmd byte
	}{
		{name: "known command echoes its id", data: []byte{0x33}, wantCmd: 0x33},
		{name: "empty payload uses the fallback id", data: nil, wantCmd: 0xFF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wire := New(checksum.ModeSum, 0x00, 0x01)
			request := frame.New(0x00, 0x01, tt.data)

			raw, err := wire.EncodeACK(request, tt.data)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := frame.Decode(raw, checksum.ModeSum)
			if err != nil {
				t.Fatal(err)
			}
			payload := decoded.DataBytes()
			if payload[0] != tt.wantCmd {
				t.Fatalf("ack command = 0x%02X, want 0x%02X", payload[0], tt.wantCmd)
			}
			if string(payload[1:]) != "OK" {
				t.Fatalf("ack body = %q, want %q", string(payload[1:]), "OK")
			}
			if decoded.Control != 0x80 {
				t.Fatalf("ack control = 0x%02X", decoded.Control)
			}
		})
	}
}

func TestEncodeReadTimeResponseRejectsInvalidMode(t *testing.T) {
	wire := New(checksum.Mode("md5"), 0x00, 0x01)

	if _, err := wire.EncodeReadTimeRequest(); err == nil {
		t.Fatal("EncodeReadTimeRequest() must reject an unknown checksum mode")
	}
	if _, err := wire.EncodeReadIdentityRequest(); err == nil {
		t.Fatal("EncodeReadIdentityRequest() must reject an unknown checksum mode")
	}
	if _, err := wire.EncodeACK(frame.New(0x00, 0x01, nil), nil); err == nil {
		t.Fatal("EncodeACK() must reject an unknown checksum mode")
	}
	if _, err := wire.EncodeReadIdentityResponse(frame.New(0x00, 0x01, nil), command.Identity{Model: "a", Serial: "b", Firmware: "c"}); err == nil {
		t.Fatal("EncodeReadIdentityResponse() must reject an unknown checksum mode")
	}
	if _, err := wire.EncodeReadTimeResponse(frame.New(0x00, 0x01, nil), time.Now()); err == nil {
		t.Fatal("EncodeReadTimeResponse() must reject an unknown checksum mode")
	}
}

func TestCodecRoundTripIsZoneSymmetric(t *testing.T) {
	zones := []*time.Location{
		time.UTC,
		time.Local,
		time.FixedZone("UTC+3", 3*60*60),
		time.FixedZone("UTC-7", -7*60*60),
	}

	for _, zone := range zones {
		t.Run(zone.String(), func(t *testing.T) {
			wire := New(checksum.ModeSum, 0x00, 0x01).WithLocation(zone)

			raw, err := wire.EncodeReadTimeRequest()
			if err != nil {
				t.Fatal(err)
			}
			request, err := wire.DecodeReadTimeRequest(raw)
			if err != nil {
				t.Fatal(err)
			}

			sent := time.Date(2026, 6, 2, 12, 34, 56, 0, time.UTC)
			response, err := wire.EncodeReadTimeResponse(request, sent)
			if err != nil {
				t.Fatal(err)
			}
			_, parsed, err := wire.DecodeReadTimeResponse(response)
			if err != nil {
				t.Fatal(err)
			}
			if !parsed.Time.Equal(sent.Truncate(time.Second)) {
				t.Fatalf("round trip in %s = %s, want %s", zone, parsed.Time, sent)
			}
		})
	}
}

func TestCodecDefaultsToLocalZone(t *testing.T) {
	wire := New(checksum.ModeSum, 0x00, 0x01)
	if wire.Location != time.Local {
		t.Fatalf("New() Location = %v, want time.Local", wire.Location)
	}

	zeroValue := Codec{Mode: checksum.ModeSum}
	if zeroValue.location() != time.Local {
		t.Fatal("a zero-value codec must fall back to the local zone")
	}
	if got := wire.WithLocation(nil); got.Location != time.Local {
		t.Fatal("WithLocation(nil) must keep the existing zone")
	}
}

func TestCodecSkewSurvivesZoneMismatchOnlyWhenAligned(t *testing.T) {
	device := New(checksum.ModeSum, 0x00, 0x01).WithLocation(time.FixedZone("UTC+3", 3*60*60))
	gateway := New(checksum.ModeSum, 0x00, 0x01).WithLocation(time.UTC)

	raw, err := device.EncodeReadTimeRequest()
	if err != nil {
		t.Fatal(err)
	}
	request, err := device.DecodeReadTimeRequest(raw)
	if err != nil {
		t.Fatal(err)
	}

	sent := time.Date(2026, 6, 2, 9, 0, 0, 0, time.UTC)
	response, err := device.EncodeReadTimeResponse(request, sent)
	if err != nil {
		t.Fatal(err)
	}

	_, mismatched, err := gateway.DecodeReadTimeResponse(response)
	if err != nil {
		t.Fatal(err)
	}
	if mismatched.Time.Equal(sent) {
		t.Fatal("a zone mismatch must be observable, not silently correct")
	}
	if delta := mismatched.Time.Sub(sent); delta != 3*time.Hour {
		t.Fatalf("zone mismatch = %s, want 3h", delta)
	}

	_, aligned, err := device.DecodeReadTimeResponse(response)
	if err != nil {
		t.Fatal(err)
	}
	if !aligned.Time.Equal(sent) {
		t.Fatalf("matched zones = %s, want %s", aligned.Time, sent)
	}
}
