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
	InBufferLen  = 1000
	OutBufferLen = 1000
)

type TrackItem struct {
	Name      string
	Timestamp time.Time
}

type EventMeasurement struct {
	Name string

	// Linked list to keep last N observations
	Sample *list.List

	statistics.Welford
	statistics.Measurement

	start   time.Time
	elapsed time.Duration
	buflen  int
}

func NewEventMeasurement(name string, buflen int) *EventMeasurement {
	if buflen < 100 {
		buflen = 100
	}
	return &EventMeasurement{
		Name:    name,
		start:   time.Now(),
		Welford: statistics.Welford{},
		Measurement: statistics.Measurement{
			Since:     time.Now(),
			Timestamp: time.Now(),
		},
		buflen: buflen,
	}
}

// Increment is called whenever an event from asset is seen to increment total event count
func (e *EventMeasurement) Increment() *EventMeasurement {
	if e.start.IsZero() {
		e.start = time.Now()
	}
	//e.Updated = time.Now()
	e.Total++
	return e
}

// PeriodicUpdate maintains Measurements to calculate mean and standard deviation
func (e *EventMeasurement) PeriodicUpdate() *EventMeasurement {
	e.elapsed = time.Since(e.start)
	/*
		e.LatestMsgFreq = e.Total / e.elapsed.Seconds()
		e.Welford.Update(e.LatestMsgFreq)
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
						//Timestamp: obj.Updated,
						Name: obj.Name,
					})
				}
				obj.Increment()
				d.mu.Unlock()
			case <-check.C:
				d.mu.Lock()
				for _, obj := range d.data {
					obj.PeriodicUpdate()
					log.Tracef("%+v", obj)
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
