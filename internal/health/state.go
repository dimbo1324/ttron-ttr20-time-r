package health

import "fmt"

type State string

const (
	StateUnknown  State = "unknown"
	StateOnline   State = "online"
	StateDegraded State = "degraded"
	StateOffline  State = "offline"
)

const (
	DefaultDegradeAfter = 3
	DefaultOfflineAfter = 10
	DefaultRecoverAfter = 2
	DefaultWindowSize   = 128
)

type Policy struct {
	DegradeAfter int
	OfflineAfter int
	RecoverAfter int
	WindowSize   int
}

func DefaultPolicy() Policy {
	return Policy{
		DegradeAfter: DefaultDegradeAfter,
		OfflineAfter: DefaultOfflineAfter,
		RecoverAfter: DefaultRecoverAfter,
		WindowSize:   DefaultWindowSize,
	}
}

func (p Policy) Normalize() Policy {
	if p.DegradeAfter <= 0 {
		p.DegradeAfter = DefaultDegradeAfter
	}
	if p.OfflineAfter <= 0 {
		p.OfflineAfter = DefaultOfflineAfter
	}
	if p.RecoverAfter <= 0 {
		p.RecoverAfter = DefaultRecoverAfter
	}
	if p.WindowSize <= 0 {
		p.WindowSize = DefaultWindowSize
	}
	if p.OfflineAfter < p.DegradeAfter {
		p.OfflineAfter = p.DegradeAfter
	}
	return p
}

func (p Policy) Validate() error {
	if p.DegradeAfter <= 0 {
		return fmt.Errorf("degrade threshold must be positive")
	}
	if p.OfflineAfter <= 0 {
		return fmt.Errorf("offline threshold must be positive")
	}
	if p.RecoverAfter <= 0 {
		return fmt.Errorf("recover threshold must be positive")
	}
	if p.OfflineAfter < p.DegradeAfter {
		return fmt.Errorf("offline threshold must not be below degrade threshold")
	}
	if p.WindowSize <= 0 {
		return fmt.Errorf("health window size must be positive")
	}
	return nil
}

func (s State) Healthy() bool {
	return s == StateOnline
}

func (s State) Reachable() bool {
	return s == StateOnline || s == StateDegraded
}
