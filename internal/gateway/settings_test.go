package gateway

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/frame"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/schedule"
)

// validSettings is a configuration every rule accepts, so a test can change
// exactly one field and know which rule rejected it.
func validSettings() Settings {
	return Settings{
		ScheduleMode:   string(schedule.ModeAligned),
		PollInterval:   time.Minute,
		PollOffset:     5 * time.Second,
		RequestTimeout: 1500 * time.Millisecond,
		RetryAttempts:  2,
		RetryDelay:     200 * time.Millisecond,
		ClockWarn:      2 * time.Second,
		ClockCritical:  30 * time.Second,
		DegradeAfter:   3,
		OfflineAfter:   10,
		RecoverAfter:   2,
	}
}

func TestSettingsAcceptsAWorkingConfiguration(t *testing.T) {
	if err := validSettings().Validate(); err != nil {
		t.Fatalf("Validate() = %v", err)
	}
}

func TestSettingsRejects(t *testing.T) {
	tests := []struct {
		name  string
		patch func(*Settings)
	}{
		{"an unknown schedule mode", func(s *Settings) { s.ScheduleMode = "cron" }},
		{"a zero interval", func(s *Settings) { s.PollInterval = 0 }},
		{"a negative interval", func(s *Settings) { s.PollInterval = -time.Second }},
		{"an offset at the interval", func(s *Settings) { s.PollOffset = s.PollInterval }},
		{"an offset past the interval", func(s *Settings) { s.PollOffset = 2 * s.PollInterval }},
		{"a negative offset", func(s *Settings) { s.PollOffset = -time.Second }},
		{"a zero request timeout", func(s *Settings) { s.RequestTimeout = 0 }},
		{"a negative request timeout", func(s *Settings) { s.RequestTimeout = -time.Second }},
		// A request that may outlast the interval it is scheduled in is not a
		// schedule; it is a queue that grows until the line gives up.
		{"a timeout at the interval", func(s *Settings) { s.RequestTimeout = s.PollInterval }},
		{"a timeout past the interval", func(s *Settings) { s.RequestTimeout = 2 * s.PollInterval }},
		{"negative retries", func(s *Settings) { s.RetryAttempts = -1 }},
		{"absurdly many retries", func(s *Settings) { s.RetryAttempts = 1000 }},
		{"a negative retry delay", func(s *Settings) { s.RetryDelay = -time.Second }},
		{"a zero warn threshold", func(s *Settings) { s.ClockWarn = 0 }},
		{"a zero critical threshold", func(s *Settings) { s.ClockCritical = 0 }},
		{"a critical threshold below warn", func(s *Settings) { s.ClockCritical = s.ClockWarn / 2 }},
		{"a zero degrade threshold", func(s *Settings) { s.DegradeAfter = 0 }},
		{"a zero offline threshold", func(s *Settings) { s.OfflineAfter = 0 }},
		{"a zero recover threshold", func(s *Settings) { s.RecoverAfter = 0 }},
		{"an offline threshold below degrade", func(s *Settings) { s.OfflineAfter = s.DegradeAfter - 1 }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := validSettings()
			tt.patch(&settings)

			err := settings.Validate()
			if err == nil {
				t.Fatalf("Validate() accepted %+v", settings)
			}
			// One sentinel for everything a caller got wrong, so the control
			// plane can answer "bad request" without matching on strings.
			if !errors.Is(err, ErrInvalidSettings) {
				t.Fatalf("Validate() = %v, want it to wrap ErrInvalidSettings", err)
			}
		})
	}
}

func TestServiceReportsItsStartingSettings(t *testing.T) {
	cfg := testGatewayConfig("127.0.0.1:9000")
	cfg.ScheduleMode = string(schedule.ModeAligned)
	cfg.PollInterval = time.Minute
	cfg.PollOffset = 5 * time.Second
	service := newTestService(t, cfg)

	got := service.Settings()

	if got.ScheduleMode != string(schedule.ModeAligned) || got.PollInterval != time.Minute {
		t.Fatalf("Settings() = %+v", got)
	}
	if got.RequestTimeout != cfg.RequestTimeout || got.RetryAttempts != cfg.RetryAttempts {
		t.Fatalf("Settings() = %+v", got)
	}
	if got.ClockWarn != cfg.ClockWarn || got.DegradeAfter != cfg.DegradeAfter {
		t.Fatalf("Settings() = %+v", got)
	}
	// A settings object read off a gateway must be one it would accept back.
	if err := got.Validate(); err != nil {
		t.Fatalf("a gateway reported settings it would reject: %v", err)
	}
}

func TestUpdateSettingsAppliesEverywhereItIsRead(t *testing.T) {
	service := newTestService(t, testGatewayConfig("127.0.0.1:9000"))

	next := validSettings()
	next.ScheduleMode = string(schedule.ModeInterval)
	next.PollInterval = 2 * time.Second
	next.PollOffset = 0
	next.RequestTimeout = 400 * time.Millisecond
	next.RetryAttempts = 4
	next.ClockWarn = 800 * time.Millisecond
	next.ClockCritical = 9 * time.Second
	next.DegradeAfter = 2
	next.OfflineAfter = 6
	next.RecoverAfter = 1

	applied, err := service.UpdateSettings(next)
	if err != nil {
		t.Fatal(err)
	}

	if applied.PollInterval != 2*time.Second || applied.RetryAttempts != 4 {
		t.Fatalf("applied = %+v", applied)
	}
	if service.Schedule().Interval() != 2*time.Second {
		t.Fatalf("Schedule() = %s", service.Schedule().String())
	}
	if got := service.requestTimeout(); got != 400*time.Millisecond {
		t.Fatalf("requestTimeout() = %s", got)
	}
	if got := service.currentRetry().Attempts; got != 4 {
		t.Fatalf("currentRetry().Attempts = %d", got)
	}

	// The status is what the control plane serves, so it has to move too --
	// a gateway polling every 2s while reporting 40ms is worse than one that
	// never changed.
	status := service.Status()
	if status.Schedule.Mode != string(schedule.ModeInterval) || status.Schedule.Interval != 2*time.Second {
		t.Fatalf("Status().Schedule = %+v", status.Schedule)
	}
	if status.RequestTimeout != 400*time.Millisecond || status.PollingInterval != 2*time.Second {
		t.Fatalf("Status() timings = %s/%s", status.RequestTimeout, status.PollingInterval)
	}
	if status.Retry.Attempts != 4 {
		t.Fatalf("Status().Retry = %+v", status.Retry)
	}
	if status.Clock.WarnThreshold != 800*time.Millisecond || status.Clock.CriticalThreshold != 9*time.Second {
		t.Fatalf("Status().Clock thresholds = %s/%s", status.Clock.WarnThreshold, status.Clock.CriticalThreshold)
	}
	if status.Health.DegradeAfter != 2 || status.Health.OfflineAfter != 6 || status.Health.RecoverAfter != 1 {
		t.Fatalf("Status().Health policy = %+v", status.Health)
	}
}

func TestUpdateSettingsReportsWhatWasActuallyApplied(t *testing.T) {
	service := newTestService(t, testGatewayConfig("127.0.0.1:9000"))

	next := validSettings()
	next.RetryAttempts = 3
	next.RetryDelay = 250 * time.Millisecond

	applied, err := service.UpdateSettings(next)
	if err != nil {
		t.Fatal(err)
	}

	// The retry policy fills in its own ceiling, and a caller redrawing from
	// the request rather than the reply would show a maximum backoff nobody
	// configured.
	if service.Status().Retry.MaxDelay <= 0 {
		t.Fatalf("Retry.MaxDelay = %s, want the policy's own ceiling", service.Status().Retry.MaxDelay)
	}
	if applied != service.Settings() {
		t.Fatalf("UpdateSettings() = %+v, Settings() = %+v", applied, service.Settings())
	}
}

func TestUpdateSettingsChangesNothingWhenRejected(t *testing.T) {
	service := newTestService(t, testGatewayConfig("127.0.0.1:9000"))
	before := service.Settings()
	beforeStatus := service.Status()

	broken := validSettings()
	// Valid on its own, so the clock monitor would accept it -- but the
	// schedule alongside it is not, and an update is all or nothing.
	broken.ClockWarn = 750 * time.Millisecond
	broken.ClockCritical = 8 * time.Second
	broken.PollInterval = 0

	if _, err := service.UpdateSettings(broken); err == nil {
		t.Fatal("UpdateSettings() accepted a broken configuration")
	}

	if service.Settings() != before {
		t.Fatalf("Settings() = %+v, want %+v", service.Settings(), before)
	}
	// The thresholds live in the clock monitor, which has its own lock and
	// would happily have taken them; validating first is what stops it.
	if got := service.Status().Clock.WarnThreshold; got != beforeStatus.Clock.WarnThreshold {
		t.Fatalf("clock warn threshold moved to %s on a rejected update", got)
	}
	if got := service.Schedule().Interval(); got != beforeStatus.Schedule.Interval {
		t.Fatalf("schedule moved to %s on a rejected update", got)
	}
}

func TestUpdateSettingsWhileTheGatewayIsPolling(t *testing.T) {
	wire := testWire(t, "sum")
	device := startScriptedDevice(t, "sum", func(request frame.Frame, _ int) deviceReply {
		return timeReply(t, wire, request, time.Now())
	})

	cfg := testGatewayConfig(device.addr())
	cfg.PollInterval = 30 * time.Millisecond
	service := newTestService(t, cfg)
	runService(t, service)

	waitFor(t, 3*time.Second, func() bool { return service.Status().SuccessfulReads > 0 })

	// Hammered from several goroutines at once, under -race: the poll loop
	// reads the schedule, the retry budget and the request timeout on its own
	// goroutine, and this is the test that says so.
	var wg sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for i := 0; i < 25; i++ {
				next := validSettings()
				next.ScheduleMode = string(schedule.ModeInterval)
				next.PollInterval = time.Duration(20+worker+i) * time.Millisecond
				next.PollOffset = 0
				next.RequestTimeout = 10 * time.Millisecond
				if _, err := service.UpdateSettings(next); err != nil {
					t.Errorf("UpdateSettings() = %v", err)
					return
				}
				_ = service.Status()
			}
		}(worker)
	}
	wg.Wait()

	if service.Settings().ScheduleMode != string(schedule.ModeInterval) {
		t.Fatalf("Settings() = %+v", service.Settings())
	}
	// Still polling after a hundred reconfigurations, which is the only proof
	// that rescheduling a live session does not wedge it.
	before := service.Status().SuccessfulReads
	waitFor(t, 3*time.Second, func() bool { return service.Status().SuccessfulReads > before })
}

func TestUpdateSettingsSurvivesAStopAndStart(t *testing.T) {
	wire := testWire(t, "sum")
	device := startScriptedDevice(t, "sum", func(request frame.Frame, _ int) deviceReply {
		return timeReply(t, wire, request, time.Now())
	})

	cfg := testGatewayConfig(device.addr())
	service := newTestService(t, cfg)

	next := validSettings()
	next.ScheduleMode = string(schedule.ModeInterval)
	next.PollInterval = 40 * time.Millisecond
	next.PollOffset = 0
	next.RequestTimeout = 20 * time.Millisecond
	if _, err := service.UpdateSettings(next); err != nil {
		t.Fatal(err)
	}

	// Configured while stopped, honoured when started: the settings live on
	// the service, not on the session, so a gateway reconfigured between runs
	// does not quietly come back up on its startup flags.
	runService(t, service)
	waitFor(t, 3*time.Second, func() bool { return service.Status().SuccessfulReads > 0 })

	if got := service.Status().Schedule.Interval; got != 40*time.Millisecond {
		t.Fatalf("Schedule.Interval = %s, want the configured value", got)
	}
	if got := service.Settings().RequestTimeout; got != 20*time.Millisecond {
		t.Fatalf("RequestTimeout = %s", got)
	}
}

func TestUpdateSettingsWakesASessionParkedOnALongWait(t *testing.T) {
	wire := testWire(t, "sum")
	device := startScriptedDevice(t, "sum", func(request frame.Frame, _ int) deviceReply {
		return timeReply(t, wire, request, time.Now())
	})

	cfg := testGatewayConfig(device.addr())
	// Long enough that the test would time out if the change had to wait for
	// the interval already in progress.
	cfg.PollInterval = 30 * time.Second
	cfg.ScheduleMode = string(schedule.ModeInterval)
	service := newTestService(t, cfg)
	runService(t, service)

	waitFor(t, 3*time.Second, func() bool { return service.Status().SuccessfulReads > 0 })
	before := service.Status().SuccessfulReads

	next := validSettings()
	next.ScheduleMode = string(schedule.ModeInterval)
	next.PollInterval = 30 * time.Millisecond
	next.PollOffset = 0
	next.RequestTimeout = 10 * time.Millisecond
	if _, err := service.UpdateSettings(next); err != nil {
		t.Fatal(err)
	}

	// An operator who changes a thirty-second interval to thirty milliseconds
	// must not wait out the thirty seconds already in progress.
	waitFor(t, 3*time.Second, func() bool { return service.Status().SuccessfulReads > before })
}
