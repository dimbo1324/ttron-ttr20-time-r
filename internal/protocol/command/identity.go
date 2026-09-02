package command

import (
	"fmt"
	"strings"
)

const (
	IdentitySeparator = "|"
	IdentityFields    = 3
	MaxIdentityLength = 96
)

type Identity struct {
	Model    string
	Serial   string
	Firmware string
	Raw      string
}

func (i Identity) String() string {
	return strings.Join([]string{i.Model, i.Serial, i.Firmware}, IdentitySeparator)
}

func BuildReadIdentityRequest() []byte {
	return []byte{byte(ReadIdentity)}
}

func ParseReadIdentityRequest(data []byte) error {
	return Expect(data, ReadIdentity)
}

func BuildReadIdentityResponse(identity Identity) []byte {
	body := identity.String()
	payload := make([]byte, 0, 1+len(body))
	payload = append(payload, byte(ReadIdentity))
	payload = append(payload, []byte(body)...)
	return payload
}

func ParseReadIdentityResponse(data []byte) (Identity, error) {
	if err := Expect(data, ReadIdentity); err != nil {
		return Identity{}, err
	}
	body := data[1:]
	if len(body) == 0 {
		return Identity{}, fmt.Errorf("%w: empty identity body", ErrInvalidPayload)
	}
	if len(body) > MaxIdentityLength {
		return Identity{}, fmt.Errorf("%w: identity body is %d bytes, max %d", ErrInvalidPayload, len(body), MaxIdentityLength)
	}

	raw := string(body)
	parts := strings.Split(raw, IdentitySeparator)
	if len(parts) != IdentityFields {
		return Identity{}, fmt.Errorf("%w: identity has %d fields, want %d", ErrInvalidIdentity, len(parts), IdentityFields)
	}
	for index, part := range parts {
		if strings.TrimSpace(part) == "" {
			return Identity{}, fmt.Errorf("%w: identity field %d is empty", ErrInvalidIdentity, index)
		}
	}
	return Identity{
		Model:    parts[0],
		Serial:   parts[1],
		Firmware: parts[2],
		Raw:      raw,
	}, nil
}
