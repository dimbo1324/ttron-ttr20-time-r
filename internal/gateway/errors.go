package gateway

import (
	"errors"
	"io"
	"net"
	"os"
	"syscall"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/command"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/frame"
)

var (
	ErrNoResponse    = errors.New("no response frame received")
	ErrPollExhausted = errors.New("read attempts exhausted")
)

var protocolErrors = []error{
	frame.ErrFrameTooShort,
	frame.ErrInvalidStartByte,
	frame.ErrInvalidRepeatedStartByte,
	frame.ErrInvalidLength,
	frame.ErrInvalidChecksum,
	frame.ErrInvalidEndByte,
	frame.ErrFrameTooLarge,
	frame.ErrUnsupportedPayloadLength,
	frame.ErrInvalidControlAddressBytes,
	command.ErrEmptyPayload,
	command.ErrUnexpectedCommand,
	command.ErrInvalidPayload,
	command.ErrInvalidTime,
	command.ErrInvalidIdentity,
}

func isProtocolError(err error) bool {
	for _, sentinel := range protocolErrors {
		if errors.Is(err, sentinel) {
			return true
		}
	}
	return false
}

func isTimeoutError(err error) bool {
	if errors.Is(err, os.ErrDeadlineExceeded) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}
	return false
}

func isConnectionLost(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	if errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ECONNABORTED) {
		return true
	}
	return false
}

func retryableOnSameConnection(err error) bool {
	if err == nil {
		return false
	}
	if isConnectionLost(err) {
		return false
	}
	if errors.Is(err, ErrNoResponse) || isTimeoutError(err) {
		return true
	}
	return isProtocolError(err)
}
