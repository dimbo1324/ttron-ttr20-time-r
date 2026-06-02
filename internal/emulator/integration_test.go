package emulator

import (
	"errors"
	"io"
	"log"
	"net"
	"testing"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/checksum"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/codec"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/frame"
)

func TestEmulatorReadTimeIntegration(t *testing.T) {
	srv, errCh := startTestServer(t, &config.EmulatorConfig{Listen: "127.0.0.1:0", CRCMode: "sum", AdapterAddr: 1, ReadTimeoutDuration: time.Second, WriteTimeoutDuration: time.Second, RecentSize: 10})
	defer stopTestServer(t, srv, errCh)

	conn := dialTestServer(t, srv)
	defer conn.Close()

	wire := codec.New(checksum.ModeSum, 0x00, 0x01)
	req, err := wire.EncodeReadTimeRequest()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Write(req); err != nil {
		t.Fatal(err)
	}

	raw := readOneFrame(t, conn, checksum.ModeSum, time.Second)
	_, resp, err := wire.DecodeReadTimeResponse(raw)
	if err != nil {
		t.Fatalf("DecodeReadTimeResponse() error = %v", err)
	}
	if resp.Raw == "" {
		t.Fatal("empty response timestamp")
	}
	if status := srv.Service().Status(); status.TotalRequests != 1 || status.TotalResponses != 1 || status.RecentFramesCount == 0 {
		t.Fatalf("status = %+v", status)
	}
}

func TestEmulatorBadChecksumFault(t *testing.T) {
	srv, errCh := startTestServer(t, &config.EmulatorConfig{Listen: "127.0.0.1:0", CRCMode: "sum", AdapterAddr: 1, BadCRCProb: 1, ReadTimeoutDuration: time.Second, WriteTimeoutDuration: time.Second, RecentSize: 10})
	defer stopTestServer(t, srv, errCh)

	conn := dialTestServer(t, srv)
	defer conn.Close()

	wire := codec.New(checksum.ModeSum, 0x00, 0x01)
	req, _ := wire.EncodeReadTimeRequest()
	if _, err := conn.Write(req); err != nil {
		t.Fatal(err)
	}
	raw := readRawExact(t, conn, checksum.ModeSum, time.Second)
	if _, err := frame.Decode(raw, checksum.ModeSum); !errors.Is(err, frame.ErrInvalidChecksum) {
		t.Fatalf("Decode() error = %v, want ErrInvalidChecksum", err)
	}
}

func TestEmulatorFragmentFault(t *testing.T) {
	srv, errCh := startTestServer(t, &config.EmulatorConfig{Listen: "127.0.0.1:0", CRCMode: "crc16", AdapterAddr: 1, FragProb: 1, FragmentDelay: time.Millisecond, ReadTimeoutDuration: time.Second, WriteTimeoutDuration: time.Second, RecentSize: 10})
	defer stopTestServer(t, srv, errCh)

	conn := dialTestServer(t, srv)
	defer conn.Close()

	wire := codec.New(checksum.ModeCRC16, 0x00, 0x01)
	req, _ := wire.EncodeReadTimeRequest()
	if _, err := conn.Write(req); err != nil {
		t.Fatal(err)
	}
	raw := readOneFrame(t, conn, checksum.ModeCRC16, time.Second)
	if _, _, err := wire.DecodeReadTimeResponse(raw); err != nil {
		t.Fatalf("DecodeReadTimeResponse() error = %v", err)
	}
}

func TestEmulatorNoResponseFault(t *testing.T) {
	srv, errCh := startTestServer(t, &config.EmulatorConfig{Listen: "127.0.0.1:0", CRCMode: "sum", AdapterAddr: 1, NoResponse: true, ReadTimeoutDuration: time.Second, WriteTimeoutDuration: time.Second, RecentSize: 10})
	defer stopTestServer(t, srv, errCh)

	conn := dialTestServer(t, srv)
	defer conn.Close()

	wire := codec.New(checksum.ModeSum, 0x00, 0x01)
	req, _ := wire.EncodeReadTimeRequest()
	if _, err := conn.Write(req); err != nil {
		t.Fatal(err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(150 * time.Millisecond))
	buf := make([]byte, 16)
	if _, err := conn.Read(buf); err == nil {
		t.Fatal("expected timeout")
	}
}

func startTestServer(t *testing.T, cfg *config.EmulatorConfig) (*Server, chan error) {
	t.Helper()
	if cfg.FragmentDelay == 0 {
		cfg.FragmentDelay = time.Millisecond
	}
	srv := NewServer(cfg, log.New(io.Discard, "", 0))
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Start() }()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if srv.Addr() != nil {
			return srv, errCh
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("server did not start")
	return nil, nil
}

func stopTestServer(t *testing.T, srv *Server, errCh chan error) {
	t.Helper()
	srv.Stop()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("server stopped with error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not stop")
	}
}

func dialTestServer(t *testing.T, srv *Server) net.Conn {
	t.Helper()
	conn, err := net.DialTimeout("tcp", srv.Addr().String(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	return conn
}

func readOneFrame(t *testing.T, conn net.Conn, mode checksum.Mode, timeout time.Duration) []byte {
	t.Helper()
	raw := readRawFrame(t, conn, mode, timeout)
	if _, err := frame.Decode(raw, mode); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	return raw
}

func readRawFrame(t *testing.T, conn net.Conn, mode checksum.Mode, timeout time.Duration) []byte {
	t.Helper()
	parser := frame.NewStreamParser(mode)
	buf := make([]byte, 1024)
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_ = conn.SetReadDeadline(deadline)
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		result := parser.Push(buf[:n])
		if len(result.Frames) > 0 {
			return result.Frames[0].RawBytes()
		}
	}
	t.Fatal("no frame received")
	return nil
}

func readRawExact(t *testing.T, conn net.Conn, mode checksum.Mode, timeout time.Duration) []byte {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	header := make([]byte, 3)
	if _, err := io.ReadFull(conn, header); err != nil {
		t.Fatal(err)
	}
	checksumLen, err := mode.ChecksumLength()
	if err != nil {
		t.Fatal(err)
	}
	tail := make([]byte, int(header[1])+checksumLen+1)
	if _, err := io.ReadFull(conn, tail); err != nil {
		t.Fatal(err)
	}
	return append(header, tail...)
}

func TestFaultModeClamp(t *testing.T) {
	cfg := &config.EmulatorConfig{BadCRCProb: 2, FragProb: -1}
	f := FaultModeFromConfig(cfg)
	if f.BadChecksumProb != 1 || f.FragmentProb != 0 {
		t.Fatalf("fault = %+v", f)
	}
}
