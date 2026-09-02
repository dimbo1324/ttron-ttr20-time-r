package emulator

import (
	"io"
	"log"
	"testing"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/checksum"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/codec"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/command"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/frame"
)

func newHandlerService(t *testing.T, mutate func(*config.EmulatorConfig)) *Service {
	t.Helper()
	cfg := config.DefaultEmulator()
	cfg.LogFile = ""
	if mutate != nil {
		mutate(&cfg)
	}
	cfg.Normalize()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("test config must be valid: %v", err)
	}
	service, err := NewService(&cfg, log.New(io.Discard, "", 0))
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func TestBuildResponseServesReadTime(t *testing.T) {
	service := newHandlerService(t, nil)
	wire := codec.New(checksum.ModeSum, 0x00, 0x01)
	request := frame.New(0x00, 0x01, command.BuildReadTimeRequest())

	raw, name, handled, err := service.BuildResponse(request)
	if err != nil {
		t.Fatal(err)
	}
	if !handled || name != command.NameReadTime {
		t.Fatalf("BuildResponse() = (%q, %t)", name, handled)
	}

	_, parsed, err := wire.DecodeReadTimeResponse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if delta := time.Since(parsed.Time); delta > 5*time.Second || delta < -5*time.Second {
		t.Fatalf("device time = %s, want about now", parsed.Time)
	}
}

func TestBuildResponseServesReadIdentity(t *testing.T) {
	service := newHandlerService(t, func(cfg *config.EmulatorConfig) {
		cfg.IdentityModel = "TTR20"
		cfg.IdentitySerial = "SN-0000042"
		cfg.IdentityFirmware = "1.2.3"
	})
	wire := codec.New(checksum.ModeSum, 0x00, 0x01)
	request := frame.New(0x00, 0x01, command.BuildReadIdentityRequest())

	raw, name, handled, err := service.BuildResponse(request)
	if err != nil {
		t.Fatal(err)
	}
	if !handled || name != command.NameReadIdentity {
		t.Fatalf("BuildResponse() = (%q, %t)", name, handled)
	}

	_, identity, err := wire.DecodeReadIdentityResponse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if identity.Model != "TTR20" || identity.Serial != "SN-0000042" || identity.Firmware != "1.2.3" {
		t.Fatalf("identity = %+v", identity)
	}
	if service.Identity().Serial != "SN-0000042" {
		t.Fatalf("Identity() = %+v", service.Identity())
	}
}

func TestBuildResponseFallsBackToAck(t *testing.T) {
	service := newHandlerService(t, nil)

	tests := []struct {
		name     string
		data     []byte
		wantName string
	}{
		{name: "unknown command", data: []byte{0x7F}, wantName: "unknown-0x7F"},
		{name: "empty payload", data: nil, wantName: "ack"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, name, handled, err := service.BuildResponse(frame.New(0x00, 0x01, tt.data))
			if err != nil {
				t.Fatal(err)
			}
			if handled {
				t.Fatal("an unknown command must not be reported as handled")
			}
			if name != tt.wantName {
				t.Fatalf("BuildResponse() name = %q, want %q", name, tt.wantName)
			}
			decoded, err := frame.Decode(raw, checksum.ModeSum)
			if err != nil {
				t.Fatal(err)
			}
			if string(decoded.DataBytes()[1:]) != "OK" {
				t.Fatalf("ack body = %q", string(decoded.DataBytes()[1:]))
			}
		})
	}
}

func TestBuildResponseNamesKnownCommandsThroughTheRegistry(t *testing.T) {
	service := newHandlerService(t, nil)

	if got := service.commandName(command.BuildReadTimeRequest()); got != command.NameReadTime {
		t.Fatalf("commandName() = %q, want %q", got, command.NameReadTime)
	}
	if got := service.commandName(command.BuildReadIdentityRequest()); got != command.NameReadIdentity {
		t.Fatalf("commandName() = %q, want %q", got, command.NameReadIdentity)
	}
	if got := service.commandName(nil); got != "ack" {
		t.Fatalf("commandName(nil) = %q, want %q", got, "ack")
	}
	if !service.Commands().Supports(command.ReadIdentity) {
		t.Fatal("the emulator must advertise the identity command")
	}
}

func TestEmulatorAppliesClockOffsetToReadTime(t *testing.T) {
	const offset = 90 * time.Second
	service := newHandlerService(t, func(cfg *config.EmulatorConfig) {
		cfg.ClockOffset = offset
	})
	wire := codec.New(checksum.ModeSum, 0x00, 0x01)

	raw, _, _, err := service.BuildResponse(frame.New(0x00, 0x01, command.BuildReadTimeRequest()))
	if err != nil {
		t.Fatal(err)
	}
	_, parsed, err := wire.DecodeReadTimeResponse(raw)
	if err != nil {
		t.Fatal(err)
	}

	skew := parsed.Time.Sub(time.Now())
	if delta := skew - offset; delta > 2*time.Second || delta < -2*time.Second {
		t.Fatalf("device clock skew = %s, want about %s", skew, offset)
	}
	if got := service.Status().FaultMode.ClockOffset; got != offset {
		t.Fatalf("Status().FaultMode.ClockOffset = %s, want %s", got, offset)
	}
}

func TestEmulatorSetFaultModeReconfiguresClock(t *testing.T) {
	service := newHandlerService(t, nil)

	applied := service.SetFaultMode(FaultMode{ClockOffset: 30 * time.Second, ClockDriftPerDay: 12 * time.Second})
	if applied.ClockOffset != 30*time.Second || applied.ClockDriftPerDay != 12*time.Second {
		t.Fatalf("SetFaultMode() = %+v", applied)
	}

	offset, drift := service.clock.Settings()
	if offset != 30*time.Second || drift != 12*time.Second {
		t.Fatalf("clock settings = (%s, %s)", offset, drift)
	}

	skew := service.DeviceTime().Sub(time.Now())
	if delta := skew - 30*time.Second; delta > 2*time.Second || delta < -2*time.Second {
		t.Fatalf("DeviceTime() skew = %s, want about 30s", skew)
	}
	if got := service.FaultMode().ClockOffset; got != 30*time.Second {
		t.Fatalf("FaultMode().ClockOffset = %s", got)
	}
}

func TestFaultModeFromConfigCarriesClockFaults(t *testing.T) {
	cfg := config.DefaultEmulator()
	cfg.ClockOffset = 5 * time.Second
	cfg.ClockDrift = -10 * time.Second

	fault := FaultModeFromConfig(&cfg)
	if fault.ClockOffset != 5*time.Second || fault.ClockDriftPerDay != -10*time.Second {
		t.Fatalf("FaultModeFromConfig() = %+v", fault)
	}
}

func TestEmulatorIdentityDefaultsAreUsable(t *testing.T) {
	service := newHandlerService(t, func(cfg *config.EmulatorConfig) {
		cfg.IdentityModel = "  "
		cfg.IdentitySerial = ""
		cfg.IdentityFirmware = ""
	})

	identity := service.Identity()
	if identity.Model == "" || identity.Serial == "" || identity.Firmware == "" {
		t.Fatalf("Identity() = %+v, want normalized defaults", identity)
	}

	payload := command.BuildReadIdentityResponse(identity)
	if _, err := command.ParseReadIdentityResponse(payload); err != nil {
		t.Fatalf("default identity must be encodable: %v", err)
	}
}
