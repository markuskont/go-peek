package trackers

import (
	"container/list"
	"context"
	"sync"
	"time"

	"github.com/ccdcoe/go-peek/pkg/statistics"
	log "github.com/sirupsen/logrus"
)

var (
	InBufferLen  = 10000
	OutBufferLen = 1000
)

type TrackItem struct {
	Name      string
	Timestamp time.Time
}

// Sample is a FIFO-like data structure for keeping a linked list of K values
type Sample struct {
	*list.List
	K int
}

func NewSample(k int) *Sample {
	return &Sample{
		K:    k,
		List: list.New(),
	}
}

func (s *Sample) Push(item interface{}) {
	if s.Len() == s.K {
		s.Remove(s.Back())
	}
	s.PushFront(item)
}

type EventMeasurement struct {
	Name string

	statistics.Welford
	Counters *statistics.Measurement
	Rates    *statistics.Measurement

	Counter float64
	Zscore  float64

	started time.Time
	elapsed time.Duration
	buflen  int
}

func NewEventMeasurement(name string, buflen int) *EventMeasurement {
	if buflen < 100 {
		buflen = 100
	}
	return &EventMeasurement{
		Name:    name,
		started: time.Now(),
		Welford: statistics.Welford{},
		Counters: &statistics.Measurement{
			Timestamp: time.Now(),
		},
		Rates: &statistics.Measurement{
			Timestamp: time.Now(),
		},
		buflen: buflen,
	}
}

// Increment is called whenever an event from asset is seen to increment total event count
func (e *EventMeasurement) Increment() *EventMeasurement {
	if e.started.IsZero() {
		e.started = time.Now()
	}
	e.Counter += 1.0
	return e
}

// PeriodicUpdate maintains Measurements to calculate mean and standard deviation
func (e *EventMeasurement) PeriodicUpdate(gauge bool) *EventMeasurement {
	e.elapsed = time.Since(e.started)
	e.Counters.Update(e.Counter)
	//e.Counters.CalculateRatePerSec(time.Since(e.started))
	//e.Rates.Update(e.Counters.Rate)
	e.Welford.Update(e.Counter)
	if gauge {
		e.Counter = 0.0
	}
	e.Zscore = statistics.AbsStandardize(e.Counter, e.Welford.Mean(), e.Welford.SdtDev())
	/*
		if status, isOutlier := statistics.ValueIsOutlier(
			e.Counter,
			e.Welford.Mean(),
			e.Welford.SdtDev(),
		); isOutlier && status == statistics.ValueBelowNormal || status == statistics.ValueAbnormal {
			e.EventState = StateWarn
		} else {
			e.EventState = StateOk
		}
	*/
	return e
}

type DeadmanConfig struct {
	CheckInterval     time.Duration
	MeasurementBuffer int
}

func (d *DeadmanConfig) Validate() error {
	if d.CheckInterval == 0 {
		d.CheckInterval = 5 * time.Second
	}
	if d.MeasurementBuffer < 100 {
		d.MeasurementBuffer = 100
	}
	return nil
}

type Deadman struct {
	mu      *sync.Mutex
	wg      *sync.WaitGroup
	data    map[string]*EventMeasurement
	rx      chan TrackItem
	tx      chan interface{}
	active  bool
	check   time.Duration
	samples int
}

func NewDeadman(c *DeadmanConfig) (*Deadman, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	d := &Deadman{
		mu:      &sync.Mutex{},
		wg:      &sync.WaitGroup{},
		rx:      make(chan TrackItem, InBufferLen),
		tx:      make(chan interface{}, OutBufferLen),
		data:    make(map[string]*EventMeasurement),
		active:  true,
		check:   c.CheckInterval,
		samples: c.MeasurementBuffer,
	}
	go func(ctx context.Context) {
		d.wg.Add(1)
		defer close(d.tx)
		defer close(d.rx)
		defer func() { d.active = false }()
		defer d.wg.Done()
		check := time.NewTicker(d.check)
	loop:
		for {
			select {
			case item, ok := <-d.rx:
				if !ok {
					break loop
				}
				d.mu.Lock()
				obj, ok := d.data[item.Name]
				if !ok {
					obj = NewEventMeasurement(item.Name, d.samples)
					d.data[item.Name] = obj.Increment()
					d.txSend(AssetFirstSeenOnWire{
						Timestamp: obj.started,
						Name:      obj.Name,
					})
				}
				obj.Increment()
				d.mu.Unlock()
			case <-check.C:
				d.mu.Lock()
				for k, obj := range d.data {
					obj.PeriodicUpdate(true)
					log.Tracef(
						"zscore: %.4f, val: %.f, mean: %.2f, dev: %.4f, %s",
						obj.Zscore,
						obj.Counters.Values.Current,
						obj.Welford.Mean(),
						obj.SdtDev(),
						k,
					)
				}
				d.mu.Unlock()
			case <-ctx.Done():
				break loop
			}
		}

	}(context.TODO())
	return d, nil
}

func (d Deadman) txSend(event interface{}) bool {
	if len(d.tx) == OutBufferLen {
		<-d.tx
	}
	d.tx <- event
	return true
}

func (d *Deadman) Send(item TrackItem) bool {
	if !d.active {
		return false
	}
	if d.rx == nil {
		d.rx = make(chan TrackItem, InBufferLen)
	}
	if len(d.rx) == InBufferLen {
		<-d.rx
	}
	d.rx <- item
	return true
}

func (d *Deadman) Input() chan<- TrackItem {
	if !d.active {
		return nil
	}
	if d.rx == nil {
		d.rx = make(chan TrackItem, InBufferLen)
	}
	return d.rx
}

func (d Deadman) Events() <-chan interface{} { return d.tx }
