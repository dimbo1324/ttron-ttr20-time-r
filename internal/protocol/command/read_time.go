package command

import (
	"fmt"
	"time"
)

const ReadTimeLayout = "2006-01-02 15:04:05"

type ReadTimeResponse struct {
	Time time.Time
	Raw  string
}

func BuildReadTimeRequest() []byte {
	return []byte{byte(ReadTime)}
}

func ParseReadTimeRequest(data []byte) error {
	return Expect(data, ReadTime)
}

func BuildReadTimeResponse(t time.Time) []byte {
	payload := make([]byte, 0, 1+len(ReadTimeLayout))
	payload = append(payload, byte(ReadTime))
	payload = append(payload, []byte(t.Format(ReadTimeLayout))...)
	return payload
}

func ParseReadTimeResponse(data []byte) (ReadTimeResponse, error) {
	if err := Expect(data, ReadTime); err != nil {
		return ReadTimeResponse{}, err
	}
	if len(data) != 1+len(ReadTimeLayout) {
		return ReadTimeResponse{}, fmt.Errorf("%w: got %d bytes, want %d", ErrInvalidPayload, len(data), 1+len(ReadTimeLayout))
	}

	raw := string(data[1:])
	parsed, err := time.Parse(ReadTimeLayout, raw)
	if err != nil {
		return ReadTimeResponse{}, fmt.Errorf("%w: %v", ErrInvalidTime, err)
	}
	return ReadTimeResponse{Time: parsed, Raw: raw}, nil
}
