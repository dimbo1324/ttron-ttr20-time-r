package gateway

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/checksum"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/codec"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/command"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/frame"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/schedule"
)

type deviceReply struct {
	raw   []byte
	close bool
}

type scriptedDevice struct {
	t        *testing.T
	mode     checksum.Mode
	listener net.Listener
	handler  func(request frame.Frame, index int) deviceReply

	mu          sync.Mutex
	requests    int
	connections int
}

func startScriptedDevice(t *testing.T, modeName string, handler func(request frame.Frame, index int) deviceReply) *scriptedDevice {
	t.Helper()
	mode, err := checksum.ParseMode(modeName)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	device := &scriptedDevice{
		t:        t,
		mode:     mode,
		listener: listener,
		handler:  handler,
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			device.mu.Lock()
			device.connections++
			device.mu.Unlock()
			go device.serve(conn)
		}
	}()

	t.Cleanup(func() {
		_ = listener.Close()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Error("scripted device did not stop")
		}
	})
	return device
}

func (d *scriptedDevice) serve(conn net.Conn) {
	defer conn.Close()
	parser := frame.NewStreamParser(d.mode)
	buffer := make([]byte, 1024)

	for {
		count, err := conn.Read(buffer)
		if count > 0 {
			result := parser.Push(buffer[:count])
			for _, request := range result.Frames {
				d.mu.Lock()
				index := d.requests
				d.requests++
				d.mu.Unlock()

				reply := d.handler(request, index)
				if len(reply.raw) > 0 {
					if _, writeErr := conn.Write(reply.raw); writeErr != nil {
						return
					}
				}
				if reply.close {
					return
				}
			}
		}
		if err != nil {
			if !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) {
				return
			}
			return
		}
	}
}

func (d *scriptedDevice) addr() string {
	return d.listener.Addr().String()
}

func (d *scriptedDevice) requestCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.requests
}

func (d *scriptedDevice) connectionCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.connections
}

func testWire(t *testing.T, modeName string) codec.Codec {
	t.Helper()
	mode, err := checksum.ParseMode(modeName)
	if err != nil {
		t.Fatal(err)
	}
	return codec.New(mode, 0x00, 0x01)
}

func timeReply(t *testing.T, wire codec.Codec, request frame.Frame, at time.Time) deviceReply {
	t.Helper()
	raw, err := wire.EncodeReadTimeResponse(request, at)
	if err != nil {
		t.Errorf("EncodeReadTimeResponse() error = %v", err)
		return deviceReply{}
	}
	return deviceReply{raw: raw}
}

func corruptTimeReply(t *testing.T, wire codec.Codec, request frame.Frame, at time.Time) deviceReply {
	t.Helper()
	reply := timeReply(t, wire, request, at)
	if len(reply.raw) == 0 {
		return reply
	}
	corrupt := append([]byte(nil), reply.raw...)
	corrupt[len(corrupt)-2] ^= 0xFF
	return deviceReply{raw: corrupt}
}

func identityReply(t *testing.T, wire codec.Codec, request frame.Frame, identity command.Identity) deviceReply {
	t.Helper()
	raw, err := wire.EncodeReadIdentityResponse(request, identity)
	if err != nil {
		t.Errorf("EncodeReadIdentityResponse() error = %v", err)
		return deviceReply{}
	}
	return deviceReply{raw: raw}
}

func ackReply(t *testing.T, wire codec.Codec, request frame.Frame) deviceReply {
	t.Helper()
	raw, err := wire.EncodeACK(request, request.DataBytes())
	if err != nil {
		t.Errorf("EncodeACK() error = %v", err)
		return deviceReply{}
	}
	return deviceReply{raw: raw}
}

func requestedCommand(request frame.Frame) command.ID {
	id, err := command.ParseID(request.DataBytes())
	if err != nil {
		return command.ID(0xFF)
	}
	return id
}

func testGatewayConfig(target string) *config.GatewayConfig {
	cfg := config.DefaultGateway()
	cfg.Target = target
	cfg.CRCMode = "sum"
	cfg.AdapterAddr = 1
	// A fixed rate, stated rather than inherited: these tests poll every 40ms
	// to stay fast, and the shipped default is a calendar schedule whose +5s
	// offset cannot fit inside an interval that short.
	cfg.ScheduleMode = string(schedule.ModeInterval)
	cfg.PollOffset = 0
	cfg.PollInterval = 40 * time.Millisecond
	cfg.RequestTimeout = 300 * time.Millisecond
	cfg.ConnectTimeout = 500 * time.Millisecond
	cfg.BackoffInitial = 10 * time.Millisecond
	cfg.BackoffMax = 30 * time.Millisecond
	cfg.RecentSize = 64
	cfg.RetryAttempts = 2
	cfg.RetryDelay = 5 * time.Millisecond
	cfg.IdentityProbe = false
	cfg.LogFile = ""
	cfg.GRPCListen = ""
	cfg.Normalize()
	return &cfg
}

func newTestService(t *testing.T, cfg *config.GatewayConfig) *Service {
	t.Helper()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("test config must be valid: %v", err)
	}
	service, err := NewService(cfg, log.New(io.Discard, "", 0))
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func runService(t *testing.T, service *Service) func() {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- service.Run(ctx) }()

	stopped := false
	stop := func() {
		if stopped {
			return
		}
		stopped = true
		cancel()
		select {
		case err := <-errCh:
			if err != nil {
				t.Errorf("service stopped with error: %v", err)
			}
		case <-time.After(3 * time.Second):
			t.Error("service did not stop")
		}
	}
	t.Cleanup(stop)
	return stop
}
