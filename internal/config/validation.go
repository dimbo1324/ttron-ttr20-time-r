package config

import (
	"fmt"
	"time"

	"github.com/dimbo1324/ttron-ttr20-time-r/internal/protocol/checksum"
)

func validateChecksumMode(mode string) error {
	_, err := checksum.ParseMode(mode)
	return err
}

func validateAdapterAddr(adapter int) error {
	if adapter < 0 || adapter > 255 {
		return fmt.Errorf("adapter address must be in range 0..255")
	}
	return nil
}

func validateProbability(value float64, name string) error {
	if value < 0 || value > 1 {
		return fmt.Errorf("%s must be in range 0..1", name)
	}
	return nil
}

func validatePositiveDuration(value time.Duration, name string) error {
	if value <= 0 {
		return fmt.Errorf("%s must be positive", name)
	}
	return nil
}

func validateNonNegativeDuration(value time.Duration, name string) error {
	if value < 0 {
		return fmt.Errorf("%s must be non-negative", name)
	}
	return nil
}

func validateRecentSize(value int) error {
	if value <= 0 {
		return fmt.Errorf("recent size must be positive")
	}
	return nil
}
