package codec

import (
	"testing"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/checksum"
)

func TestEncodeDecodeReadTimeRequest(t *testing.T) {
	for _, mode := range []checksum.Mode{checksum.ModeSum, checksum.ModeCRC16} {
		t.Run(string(mode), func(t *testing.T) {
			c := New(mode, 0x00, 0x01)
			raw, err := c.EncodeReadTimeRequest()
			if err != nil {
				t.Fatalf("EncodeReadTimeRequest() error = %v", err)
			}
			f, err := c.DecodeReadTimeRequest(raw)
			if err != nil {
				t.Fatalf("DecodeReadTimeRequest() error = %v", err)
			}
			if f.Control != 0x00 || f.Address != 0x01 {
				t.Fatalf("frame = %+v", f)
			}
		})
	}
}

func TestEncodeDecodeReadTimeResponse(t *testing.T) {
	for _, mode := range []checksum.Mode{checksum.ModeSum, checksum.ModeCRC16} {
		t.Run(string(mode), func(t *testing.T) {
			c := New(mode, 0x00, 0x01)
			reqRaw, err := c.EncodeReadTimeRequest()
			if err != nil {
				t.Fatal(err)
			}
			req, err := c.DecodeReadTimeRequest(reqRaw)
			if err != nil {
				t.Fatal(err)
			}

			ts := time.Date(2026, 6, 2, 12, 34, 56, 0, time.UTC)
			respRaw, err := c.EncodeReadTimeResponse(req, ts)
			if err != nil {
				t.Fatalf("EncodeReadTimeResponse() error = %v", err)
			}
			respFrame, resp, err := c.DecodeReadTimeResponse(respRaw)
			if err != nil {
				t.Fatalf("DecodeReadTimeResponse() error = %v", err)
			}
			if respFrame.Control != 0x80 || respFrame.Address != 0x01 || !resp.Time.Equal(ts) {
				t.Fatalf("response = frame %+v payload %+v", respFrame, resp)
			}
		})
	}
}
