package emulator

import (
	"sync"
	"time"
)

const nanosecondsPerDay = float64(24 * time.Hour)

type DeviceClock struct {
	mu sync.RWMutex

	source      func() time.Time
	origin      time.Time
	offset      time.Duration
	driftPerDay time.Duration
}

func NewDeviceClock(source func() time.Time, offset, driftPerDay time.Duration) *DeviceClock {
	if source == nil {
		source = time.Now
	}
	return &DeviceClock{
		source:      source,
		origin:      source(),
		offset:      offset,
		driftPerDay: driftPerDay,
	}
}

func (c *DeviceClock) Now() time.Time {
	c.mu.RLock()
	source := c.source
	origin := c.origin
	offset := c.offset
	driftPerDay := c.driftPerDay
	c.mu.RUnlock()

	now := source()
	if driftPerDay == 0 {
		return now.Add(offset)
	}
	elapsed := now.Sub(origin)
	if elapsed < 0 {
		elapsed = 0
	}
	accumulated := time.Duration(float64(driftPerDay) * (float64(elapsed) / nanosecondsPerDay))
	return now.Add(offset + accumulated)
}

func (c *DeviceClock) Configure(offset, driftPerDay time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.offset = offset
	if c.driftPerDay != driftPerDay {
		c.origin = c.source()
		c.driftPerDay = driftPerDay
	}
}

func (c *DeviceClock) Settings() (offset, driftPerDay time.Duration) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.offset, c.driftPerDay
}
