package statistics

import "math"

// Welford online algorithm allows us to accurately estimate sample mean and variance on a single pass
// Normal method would require keeping full vector of data in memory which is not often feasible or practical in online scenario
// Naive incremental method can be inaccruate at extreme values and even produce negative variance, which is impossible by definition
type Welford struct {
	mean, variance, s, k float64
}

func (w *Welford) Update(item float64) *Welford {
	w.k++
	mNext := w.mean + (item-w.mean)/w.k
	w.s = w.s + (item-w.mean)*(item-mNext)
	w.mean = mNext
	w.variance = w.s / (w.k - 1)
	return w
}

func (w Welford) Mean() float64     { return w.mean }
func (w Welford) Variance() float64 { return w.variance }
func (w Welford) SdtDev() float64   { return math.Sqrt(w.variance) }
func (w Welford) K() float64        { return w.k }
