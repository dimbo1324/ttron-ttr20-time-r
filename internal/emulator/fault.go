package emulator

import (
	"math/rand"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/config"
	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/frame"
)

type FaultMode struct {
	ResponseDelay     time.Duration
	BadChecksumProb   float64
	FragmentProb      float64
	FragmentDelay     time.Duration
	NoResponse        bool
	CloseAfterRequest bool
}

func FaultModeFromConfig(cfg *config.EmulatorConfig) FaultMode {
	return FaultMode{
		ResponseDelay:     cfg.ResponseDelay,
		BadChecksumProb:   clampProbability(cfg.BadCRCProb),
		FragmentProb:      clampProbability(cfg.FragProb),
		FragmentDelay:     cfg.FragmentDelay,
		NoResponse:        cfg.NoResponse,
		CloseAfterRequest: cfg.CloseAfterRequest,
	}
}

func (f FaultMode) ShouldCorruptChecksum() bool {
	return f.BadChecksumProb > 0 && rand.Float64() < f.BadChecksumProb
}

func (f FaultMode) ShouldFragment() bool {
	return f.FragmentProb > 0 && rand.Float64() < f.FragmentProb
}

func (f FaultMode) ApplyCorruptChecksum(raw []byte, mode string) {
	frame.CorruptChecksum(raw, mode)
}

func clampProbability(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
