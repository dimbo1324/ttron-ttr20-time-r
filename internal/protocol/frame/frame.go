package frame

const (
	StartByte byte = 0x68
	EndByte   byte = 0x16
)

type Frame struct {
	Control byte
	Address byte
	Data    []byte
	raw     []byte
}

func New(control, address byte, data []byte) Frame {
	return Frame{
		Control: control,
		Address: address,
		Data:    append([]byte(nil), data...),
	}
}

func (f Frame) DataBytes() []byte {
	return append([]byte(nil), f.Data...)
}

func (f Frame) RawBytes() []byte {
	return append([]byte(nil), f.raw...)
}

func (f Frame) PayloadBytes() []byte {
	payload := make([]byte, 0, 2+len(f.Data))
	payload = append(payload, f.Control, f.Address)
	payload = append(payload, f.Data...)
	return payload
}
