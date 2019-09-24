package trackers

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
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
	Name         string
	Total        float64
	MsgFrequency float64
	Measurements [256]float64
	Updated      time.Time
	start        time.Time
	elapsed      time.Duration
}

func (e *EventMeasurement) Update() *EventMeasurement {
	if e.start.IsZero() {
		e.start = time.Now()
	}
	e.Updated = time.Now()
	e.elapsed = time.Since(e.start)
	e.Total++
	e.MsgFrequency = e.Total / e.elapsed.Seconds()
	return e
}

type DeadmanConfig struct {
	CheckInterval time.Duration
}

func (d *DeadmanConfig) Validate() error {
	if d.CheckInterval == 0 {
		d.CheckInterval = 5 * time.Second
	}
	return nil
}

type Deadman struct {
	mu     *sync.Mutex
	wg     *sync.WaitGroup
	data   map[string]*EventMeasurement
	rx     chan TrackItem
	tx     chan interface{}
	active bool
	check  time.Duration
}

func NewDeadman(c *DeadmanConfig) (*Deadman, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	d := &Deadman{
		mu:     &sync.Mutex{},
		wg:     &sync.WaitGroup{},
		rx:     make(chan TrackItem, InBufferLen),
		tx:     make(chan interface{}, OutBufferLen),
		data:   make(map[string]*EventMeasurement),
		active: true,
		check:  c.CheckInterval,
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
					obj = &EventMeasurement{
						Name:  item.Name,
						start: time.Now(),
					}
					d.data[item.Name] = obj
					d.txSend(AssetFirstSeenOnWire{
						Timestamp: obj.Updated,
						Name:      obj.Name,
					})
				}
				d.mu.Unlock()
			case <-check.C:
				d.mu.Lock()
				for _, v := range d.data {
					fmt.Fprintf(os.Stdout, "%+v\n", v)
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
