package statistics

import "time"

var (
	OutlierScore = 4.0
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

// Normalize, or Min-Max Scalar will scale a value to a fixed range, usually 0 to 1
func Normalize(value, min, max float64) float64 { return (value - min) / (max - min) }

func ValueIsOutlier(val, mean, stdDev float64) bool {
	rng := OutlierScore * stdDev
	return val > mean-rng || val < mean+rng
}

type Measurement struct {
	Tags             map[string]string
	Last, Current    float64
	Rate             float64
	Total            float64
	Timestamp, Since time.Time
	Period           time.Duration
}

func (m *Measurement) Calculate() *Measurement {
	return m
}
