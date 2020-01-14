package filestorage

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"sync"
	"time"

	"github.com/ccdcoe/go-peek/pkg/models/consumer"
	"github.com/ccdcoe/go-peek/pkg/utils"
	log "github.com/sirupsen/logrus"
)

var (
	TimeFmt = "20060102150405"
)

type Config struct {
	Dir       string
	Combined  string
	Gzip      bool
	Timestamp bool
	Stream    <-chan consumer.Message

	RotateEnabled  bool
	RotateInterval time.Duration
}

func (c *Config) Validate() error {
	if c == nil {
		c = &Config{}
	}
	if c.Stream == nil {
		return fmt.Errorf("missing input stream for filestorage")
	}
	if c.Combined == "" && c.Dir == "" {
		return fmt.Errorf(
			"filestorage module requires either a root directory or explicit destination file for storing events",
		)
	}
	if c.Dir != "" {
		if !utils.StringIsValidDir(c.Dir) {
			return fmt.Errorf("path %s is not valid directory", c.Dir)
		}
	}
	return nil
}

type Handle struct {
	errs *utils.ErrChan
	rx   <-chan consumer.Message

	filterChannels map[string]chan consumer.Message
	filterEnabled  bool

	combinedEnabled bool

	combined string
	dir      string

	timestamp bool
	gzip      bool

	rotate *time.Ticker

	mu *sync.Mutex
	wg *sync.WaitGroup
}

func NewHandle(c *Config) (*Handle, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &Handle{
		errs:      utils.NewErrChan(100, "filestorage handle"),
		rx:        c.Stream,
		timestamp: c.Timestamp,
		gzip:      c.Gzip,
		combined:  c.Combined,
		dir:       c.Dir,
		wg:        &sync.WaitGroup{},
		mu:        &sync.Mutex{},
		combinedEnabled: func() bool {
			if c.Combined != "" {
				return true
			}
			return false
		}(),
		filterEnabled: func() bool {
			if c.Dir != "" {
				return true
			}
			return false
		}(),
		rotate: func() *time.Ticker {
			if c.RotateEnabled {
				return time.NewTicker(c.RotateInterval)
			}
			return nil
		}(),
		filterChannels: make(map[string]chan consumer.Message),
	}, nil
}

func (h *Handle) Do(ctx context.Context) error {
	// function for formatting output file name
	filenameFunc := func(ts time.Time, path string) string {
		if h.timestamp || h.rotate != nil {
			path = fmt.Sprintf("%s.%s", path, ts.Format(TimeFmt))
		}
		if h.gzip {
			path = fmt.Sprintf("%s.gz", path)
		}
		return path
	}

	h.wg.Add(1)
	combineCh := make(chan consumer.Message, 0)
	if h.filterChannels == nil {
		h.filterChannels = make(map[string]chan consumer.Message)
	}
	go func() {
		defer func() {
			if h.filterChannels != nil {
				for _, ch := range h.filterChannels {
					close(ch)
				}
			}
		}()
		defer close(combineCh)
		defer h.wg.Done()

	loop:
		for {
			select {
			case msg, ok := <-h.rx:
				if !ok {
					break loop
				}
				if h.combinedEnabled {
					combineCh <- msg
				}

				if h.filterEnabled {
					key := msg.Event.String()
					if ch, ok := h.filterChannels[key]; ok {
						ch <- msg
					} else {
						h.mu.Lock()
						ch := make(chan consumer.Message, 0)
						h.filterChannels[key] = ch
						h.mu.Unlock()

						now := time.Now()
						if err := writeSingleFile(
							func() string {
								path := fmt.Sprintf("%s", path.Join(h.dir, msg.Event.String()))
								if h.timestamp || h.rotate != nil {
									path = fmt.Sprintf("%s.%s", path, now.Format(TimeFmt))
								}
								if h.gzip {
									path = fmt.Sprintf("%s.gz", path)
								}
								return path
							}(),
							*h.errs, ch, h.gzip, context.TODO(), h.wg); err != nil {
							h.errs.Send(err)
						}
					}
				}

			case <-ctx.Done():
				break loop
				//case _, ok := <-h.rotate.C:
			}
		}
		log.Trace("filestorage filter loop good exit")
	}()

	if h.combinedEnabled {
		if h.rotate != nil {
			panic("NOT IMPLEMENTED")
		} else {
			now := time.Now()
			if err := writeSingleFile(
				filenameFunc(now, h.combined),
				*h.errs,
				combineCh,
				h.gzip,
				context.TODO(),
				h.wg,
			); err != nil {
				return err
			}
		}
	}

	return nil
}

func (h Handle) Errors() <-chan error {
	return h.errs.Items
}

func writeSingleFile(
	path string,
	errs utils.ErrChan,
	rx <-chan consumer.Message,
	gz bool,
	ctx context.Context,
	wg *sync.WaitGroup,
) error {
	path, err := utils.ExpandHome(path)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	w := func() io.WriteCloser {
		if gz {
			return gzip.NewWriter(f)
		}
		return f
	}()
	wg.Add(1)
	go func() {
		defer w.Close()
		defer f.Close()
		defer wg.Done()
		var written int
	loop:
		for {
			select {
			case msg, ok := <-rx:
				if !ok {
					break loop
				}
				fmt.Fprintf(w, "%s\n", string(msg.Data))
				written++
			case <-ctx.Done():
				break loop
			}
		}
		log.Tracef("%s proper exit", path)
	}()
	return nil
}

func (h Handle) Wait() {
	h.wg.Wait()
	return
}
