package gateway

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"syscall"
	"testing"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/command"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/frame"
)

type fakeNetError struct {
	timeout bool
}

func (e fakeNetError) Error() string { return "fake net error" }
func (e fakeNetError) Timeout() bool { return e.timeout }
func (e fakeNetError) Temporary() bool {
	return e.timeout
}

func TestIsProtocolError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "invalid checksum", err: frame.ErrInvalidChecksum, want: true},
		{name: "wrapped invalid checksum", err: fmt.Errorf("decode: %w", frame.ErrInvalidChecksum), want: true},
		{name: "frame too short", err: frame.ErrFrameTooShort, want: true},
		{name: "invalid start byte", err: frame.ErrInvalidStartByte, want: true},
		{name: "invalid repeated start byte", err: frame.ErrInvalidRepeatedStartByte, want: true},
		{name: "invalid length", err: frame.ErrInvalidLength, want: true},
		{name: "invalid end byte", err: frame.ErrInvalidEndByte, want: true},
		{name: "frame too large", err: frame.ErrFrameTooLarge, want: true},
		{name: "unsupported payload length", err: frame.ErrUnsupportedPayloadLength, want: true},
		{name: "invalid control address", err: frame.ErrInvalidControlAddressBytes, want: true},
		{name: "empty command payload", err: command.ErrEmptyPayload, want: true},
		{name: "unexpected command", err: command.ErrUnexpectedCommand, want: true},
		{name: "invalid command payload", err: command.ErrInvalidPayload, want: true},
		{name: "invalid timestamp", err: command.ErrInvalidTime, want: true},
		{name: "invalid identity", err: command.ErrInvalidIdentity, want: true},
		{name: "transport error", err: io.EOF, want: false},
		{name: "unrelated", err: errors.New("boom"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isProtocolError(tt.err); got != tt.want {
				t.Fatalf("isProtocolError(%v) = %t, want %t", tt.err, got, tt.want)
			}
		})
	}
}

func TestIsTimeoutError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "deadline exceeded", err: os.ErrDeadlineExceeded, want: true},
		{name: "wrapped deadline", err: fmt.Errorf("read: %w", os.ErrDeadlineExceeded), want: true},
		{name: "net timeout", err: fakeNetError{timeout: true}, want: true},
		{name: "net non timeout", err: fakeNetError{}, want: false},
		{name: "plain error", err: errors.New("boom"), want: false},
		{name: "eof", err: io.EOF, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTimeoutError(tt.err); got != tt.want {
				t.Fatalf("isTimeoutError(%v) = %t, want %t", tt.err, got, tt.want)
			}
		})
	}
}

func TestIsConnectionLost(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "eof", err: io.EOF, want: true},
		{name: "unexpected eof", err: io.ErrUnexpectedEOF, want: true},
		{name: "wrapped eof", err: fmt.Errorf("read response: %w", io.EOF), want: true},
		{name: "net closed", err: net.ErrClosed, want: true},
		{name: "connection reset", err: syscall.ECONNRESET, want: true},
		{name: "connection aborted", err: syscall.ECONNABORTED, want: true},
		{name: "timeout", err: os.ErrDeadlineExceeded, want: false},
		{name: "protocol", err: frame.ErrInvalidChecksum, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isConnectionLost(tt.err); got != tt.want {
				t.Fatalf("isConnectionLost(%v) = %t, want %t", tt.err, got, tt.want)
			}
		})
	}
}

func TestRetryableOnSameConnection(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "no response", err: ErrNoResponse, want: true},
		{name: "wrapped no response", err: fmt.Errorf("%w: deadline", ErrNoResponse), want: true},
		{name: "read timeout", err: os.ErrDeadlineExceeded, want: true},
		{name: "bad checksum", err: fmt.Errorf("decode: %w", frame.ErrInvalidChecksum), want: true},
		{name: "unexpected command", err: command.ErrUnexpectedCommand, want: true},
		{name: "eof tears the session down", err: io.EOF, want: false},
		{name: "closed connection tears the session down", err: net.ErrClosed, want: false},
		{name: "connection reset tears the session down", err: syscall.ECONNRESET, want: false},
		{name: "unknown error is treated as fatal", err: errors.New("boom"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := retryableOnSameConnection(tt.err); got != tt.want {
				t.Fatalf("retryableOnSameConnection(%v) = %t, want %t", tt.err, got, tt.want)
			}
		})
	}
}

func TestSentinelErrorsAreDistinct(t *testing.T) {
	if errors.Is(ErrNoResponse, ErrPollExhausted) || errors.Is(ErrPollExhausted, ErrNoResponse) {
		t.Fatal("gateway sentinels must not alias each other")
	}
}
