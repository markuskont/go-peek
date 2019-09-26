package statistics

import (
	"math"
	"time"
)

var (
	OutlierScore = 3.0
)

func init() {
	// theoretical rule of thumb - most measurements should fall within 3 standard deviations from mean
	// no reason to have a threshold lower than that
	// higher thresh should be used in practice
	if OutlierScore < 3.0 {
		OutlierScore = 3.0
	}
}

// Standardize scales a value around mean 0 and stddev 1, allows measurements of different scales to be compared
func Standardize(value, mean, stdDev float64) float64 { return (value - mean) / stdDev }
func AbsStandardize(v, m, s float64) float64          { return math.Abs(Standardize(v, m, s)) }

// Normalize which functions as or Min-Max Scalar will scale a value to a fixed range, usually 0 to 1
func Normalize(value, min, max float64) float64 { return (value - min) / (max - min) }

type Values struct {
	Last, Current float64
	Max, Min      float64
	K             float64
}

func (v *Values) Update(value float64) *Values {
	v.Last = v.Current
	v.Current = value
	if v.Current < v.Min || v.K == 0 {
		v.Min = value
	}
	if v.Current > v.Max || v.K == 0 {
		v.Max = value
	}
	v.K++
	return v
}

func (v Values) Delta() float64 {
	return v.Current - v.Last
}

func (v Values) Normalize() float64 {
	return Normalize(v.Current, v.Min, v.Max)
}

type Measurement struct {
	Values
	Tags      map[string]string
	Rate      float64
	Timestamp time.Time
	Gauge     bool
}

func (m *Measurement) Update(value float64) *Measurement {
	m.Timestamp = time.Now()
	m.Values.Update(value)
	return m
}

func (m *Measurement) CalculateRatePerSec(since time.Duration) *Measurement {
	m.Rate = m.Values.K / since.Seconds()
	return m
}
