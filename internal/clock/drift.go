package clock

import (
	"math"
	"time"
)

const (
	MinDriftSamples   = 3
	secondsPerDay     = 86400.0
	maxDriftPerDayAbs = 365 * 24 * float64(time.Hour)
)

type Drift struct {
	PerDay      time.Duration
	Determined  bool
	SampleCount int
	Fit         float64
	Span        time.Duration
}

func computeDrift(items []entry) Drift {
	count := len(items)
	if count < MinDriftSamples {
		return Drift{SampleCount: count}
	}

	origin := items[0].at
	xs := make([]float64, count)
	ys := make([]float64, count)
	var sumX, sumY float64
	for i, item := range items {
		xs[i] = item.at.Sub(origin).Seconds()
		ys[i] = item.skew.Seconds()
		sumX += xs[i]
		sumY += ys[i]
	}

	meanX := sumX / float64(count)
	meanY := sumY / float64(count)

	var covariance, varianceX float64
	for i := range items {
		dx := xs[i] - meanX
		covariance += dx * (ys[i] - meanY)
		varianceX += dx * dx
	}
	span := items[count-1].at.Sub(origin)
	if varianceX == 0 {
		return Drift{SampleCount: count, Span: span}
	}

	slope := covariance / varianceX
	perDaySeconds := slope * secondsPerDay
	if math.IsNaN(perDaySeconds) || math.IsInf(perDaySeconds, 0) {
		return Drift{SampleCount: count, Span: span}
	}
	perDayNanos := perDaySeconds * float64(time.Second)
	if math.Abs(perDayNanos) > maxDriftPerDayAbs {
		perDayNanos = math.Copysign(maxDriftPerDayAbs, perDayNanos)
	}

	return Drift{
		PerDay:      time.Duration(perDayNanos),
		Determined:  true,
		SampleCount: count,
		Fit:         coefficientOfDetermination(xs, ys, meanX, meanY, slope),
		Span:        span,
	}
}

func coefficientOfDetermination(xs, ys []float64, meanX, meanY, slope float64) float64 {
	intercept := meanY - slope*meanX
	var residual, total float64
	for i := range xs {
		predicted := intercept + slope*xs[i]
		residual += (ys[i] - predicted) * (ys[i] - predicted)
		total += (ys[i] - meanY) * (ys[i] - meanY)
	}
	if total == 0 {
		return 0
	}
	fit := 1 - residual/total
	if math.IsNaN(fit) || math.IsInf(fit, 0) {
		return 0
	}
	return fit
}
