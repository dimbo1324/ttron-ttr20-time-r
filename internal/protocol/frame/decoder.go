package frame

import (
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/checksum"
)

func Decode(raw []byte, mode checksum.Mode) (Frame, error) {
	checksumLen, err := mode.ChecksumLength()
	if err != nil {
		return Frame{}, err
	}
	if len(raw) < 3+2+checksumLen+1 {
		return Frame{}, wrapError(ErrFrameTooShort, "got %d bytes", len(raw))
	}
	if raw[0] != StartByte {
		return Frame{}, wrapError(ErrInvalidStartByte, "got 0x%02X", raw[0])
	}
	if raw[2] != StartByte {
		return Frame{}, wrapError(ErrInvalidRepeatedStartByte, "got 0x%02X", raw[2])
	}

	payloadLen := int(raw[1])
	if payloadLen < 2 {
		return Frame{}, wrapError(ErrInvalidLength, "payload length %d", payloadLen)
	}
	wantLen := 3 + payloadLen + checksumLen + 1
	if len(raw) != wantLen {
		return Frame{}, wrapError(ErrInvalidLength, "got %d bytes, want %d", len(raw), wantLen)
	}
	if raw[len(raw)-1] != EndByte {
		return Frame{}, wrapError(ErrInvalidEndByte, "got 0x%02X", raw[len(raw)-1])
	}

	payloadStart := 3
	payloadEnd := payloadStart + payloadLen
	payload := raw[payloadStart:payloadEnd]
	gotChecksum := raw[payloadEnd : payloadEnd+checksumLen]
	if err := checksum.Verify(mode, payload, gotChecksum); err != nil {
		return Frame{}, wrapError(ErrInvalidChecksum, "%v", err)
	}

	return Frame{
		Control: payload[0],
		Address: payload[1],
		Data:    append([]byte(nil), payload[2:]...),
		raw:     append([]byte(nil), raw...),
	}, nil
}

func PayloadData(raw []byte) []byte {
	if len(raw) < 6 {
		return nil
	}
	payloadLen := int(raw[1])
	if payloadLen < 2 {
		return nil
	}
	payloadEnd := 3 + payloadLen
	if payloadEnd > len(raw)-2 {
		return nil
	}
	return append([]byte(nil), raw[5:payloadEnd]...)
}
