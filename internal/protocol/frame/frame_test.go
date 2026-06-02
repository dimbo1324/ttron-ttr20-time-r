package frame

import (
	"bytes"
	"testing"
)

func TestBuildVerifyExtractSumFrame(t *testing.T) {
	f := AppendChecksum(BuildSkeleton(0x00, 0x01, []byte{0x01}), "sum")
	want := []byte{0x68, 0x03, 0x68, 0x00, 0x01, 0x01, 0x02, 0x16}
	if !bytes.Equal(f, want) {
		t.Fatalf("frame = % X, want % X", f, want)
	}
	if err := Verify(f); err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	var buf bytes.Buffer
	buf.Write(f)
	got, ok := ExtractFrame(&buf)
	if !ok {
		t.Fatal("ExtractFrame() did not find frame")
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("ExtractFrame() = % X, want % X", got, want)
	}
}

func TestPayloadData(t *testing.T) {
	f := AppendChecksum(BuildSkeleton(0x80, 0x01, []byte{0x01, 'O', 'K'}), "sum")
	got := PayloadData(f)
	want := []byte{0x01, 'O', 'K'}
	if !bytes.Equal(got, want) {
		t.Fatalf("PayloadData() = % X, want % X", got, want)
	}
}
