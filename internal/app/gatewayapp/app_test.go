package gatewayapp

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/platform/lifecycle"
)

func testConfig() *config.GatewayConfig {
	cfg := config.DefaultGateway()
	cfg.Target = "127.0.0.1:9000"
	cfg.LogFile = ""
	cfg.GRPCListen = ""
	cfg.Normalize()
	return &cfg
}

func writeInventory(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "devices.json")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestBuildRuntimeSingleDevice(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	group := lifecycle.NewGroup(logger)

	service, err := buildRuntime(context.Background(), testConfig(), group, logger)
	if err != nil {
		t.Fatal(err)
	}
	if service == nil {
		t.Fatal("buildRuntime() must return a service in single-device mode")
	}
	if got := service.Status().TargetAddress; got != "127.0.0.1:9000" {
		t.Fatalf("TargetAddress = %q", got)
	}
	if service.DeviceID() != "" {
		t.Fatalf("DeviceID() = %q, want empty in single-device mode", service.DeviceID())
	}
}

func TestBuildRuntimeRejectsInvalidSingleDeviceConfig(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	cfg := testConfig()
	cfg.CRCMode = "md5"

	if _, err := buildRuntime(context.Background(), cfg, lifecycle.NewGroup(logger), logger); err == nil {
		t.Fatal("buildRuntime() must reject an unknown checksum mode")
	}
}

func TestBuildRuntimeLoadsDeviceInventory(t *testing.T) {
	path := writeInventory(t, `{"devices":[
      {"id":"beta","target":"127.0.0.1:9002","enabled":true},
      {"id":"alpha","target":"127.0.0.1:9001","enabled":true},
      {"id":"gamma","target":"127.0.0.1:9003","enabled":false}
    ]}`)

	logger := log.New(io.Discard, "", 0)
	cfg := testConfig()
	cfg.DevicesFile = path

	service, err := buildRuntime(context.Background(), cfg, lifecycle.NewGroup(logger), logger)
	if err != nil {
		t.Fatal(err)
	}
	if service == nil {
		t.Fatal("buildRuntime() must return the primary device service")
	}
	if service.DeviceID() != "alpha" {
		t.Fatalf("DeviceID() = %q, want the first device by id", service.DeviceID())
	}
	if got := service.Status().TargetAddress; got != "127.0.0.1:9001" {
		t.Fatalf("TargetAddress = %q", got)
	}
}

func TestBuildRuntimeWithoutEnabledDevices(t *testing.T) {
	path := writeInventory(t, `{"devices":[{"id":"alpha","target":"127.0.0.1:9001","enabled":false}]}`)

	logger := log.New(io.Discard, "", 0)
	cfg := testConfig()
	cfg.DevicesFile = path
	cfg.GRPCListen = ":9200"

	service, err := buildRuntime(context.Background(), cfg, lifecycle.NewGroup(logger), logger)
	if err != nil {
		t.Fatal(err)
	}
	if service != nil {
		t.Fatalf("buildRuntime() = %+v, want no primary service", service)
	}
}

func TestBuildRuntimeRejectsMissingInventory(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	cfg := testConfig()
	cfg.DevicesFile = filepath.Join(t.TempDir(), "absent.json")

	if _, err := buildRuntime(context.Background(), cfg, lifecycle.NewGroup(logger), logger); err == nil {
		t.Fatal("buildRuntime() must fail when the inventory is missing")
	}
}

func TestBuildRuntimeRejectsInvalidInventory(t *testing.T) {
	path := writeInventory(t, `{"devices":[{"id":"A B","target":"127.0.0.1:9001"}]}`)

	logger := log.New(io.Discard, "", 0)
	cfg := testConfig()
	cfg.DevicesFile = path

	if _, err := buildRuntime(context.Background(), cfg, lifecycle.NewGroup(logger), logger); err == nil {
		t.Fatal("buildRuntime() must reject an invalid inventory")
	}
}

func TestRunReportsConfigFailure(t *testing.T) {
	if code := Run([]string{"-schedule", "cron"}); code != 1 {
		t.Fatalf("Run() = %d, want 1", code)
	}
}

func TestRunHelpExitsCleanly(t *testing.T) {
	if code := Run([]string{"-h"}); code != 0 {
		t.Fatalf("Run(-h) = %d, want 0", code)
	}
}

func TestBuildRuntimeRegistersLifecycleTask(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	group := lifecycle.NewGroup(logger)

	if _, err := buildRuntime(context.Background(), testConfig(), group, logger); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	if err := group.Run(ctx); err != nil {
		t.Fatalf("group.Run() = %v", err)
	}
}
