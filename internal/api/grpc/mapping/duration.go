package mapping

import "time"

// Millis renders a duration the way the wire carries every interval in this
// API: whole milliseconds. Sub-millisecond precision is below what a serial
// line can resolve, and a single unit across the surface removes the "is this
// field seconds or nanoseconds?" question at every call site.
func Millis(d time.Duration) int64 {
	return int64(d / time.Millisecond)
}
