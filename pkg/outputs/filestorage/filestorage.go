package filestorage

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/ccdcoe/go-peek/pkg/models/consumer"
	"github.com/ccdcoe/go-peek/pkg/utils"
)

var (
	TimeFmt = "20060102150405"
)

type Config struct {
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
	if c.Combined == "" {
		return fmt.Errorf(
			"filestorage module requires either a root directory or explicit destination file for storing events",
		)
	}
	if c.Stream == nil {
		return fmt.Errorf("missing input stream for filestorage")
	}
	/*
		if c.Directory != "" {
			if !utils.StringIsValidDir(c.Directory) {
				return fmt.Errorf("path %s is not valid directory", c.Directory)
			}
		}
	*/
	return nil
}

type Handle struct {
	errs *utils.ErrChan
	rx   <-chan consumer.Message

	combined string

	timestamp bool
	gzip      bool

	rotateEnabled  bool
	rotateInterval time.Duration

	wg *sync.WaitGroup
}

func NewHandle(c *Config) (*Handle, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &Handle{
		errs:           utils.NewErrChan(100, "filestorage handle"),
		rx:             c.Stream,
		timestamp:      c.Timestamp,
		gzip:           c.Gzip,
		combined:       c.Combined,
		wg:             &sync.WaitGroup{},
		rotateEnabled:  c.RotateEnabled,
		rotateInterval: c.RotateInterval,
	}, nil
}

func (h *Handle) Do(ctx context.Context) error {
	// function for formatting output file name
	filenameFunc := func(ts time.Time, path string) string {
		if h.timestamp {
			path = fmt.Sprintf("%s.%s", path, ts.Format(TimeFmt))
		}
		if h.gzip {
			path = fmt.Sprintf("%s.gz", path)
		}
		return path
	}

	channels := make(map[string]<-chan consumer.Message)

	channels["combined"] = func() <-chan consumer.Message {
		return h.rx
	}()

	if h.combined != "" {
		now := time.Now()
		if err := writeSingleFile(filenameFunc(
			now,
			h.combined,
		), *h.errs, h.rx, h.gzip, ctx, h.wg); err != nil {
			return err
		}
	}

	return nil
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
		defer f.Close()
		defer w.Close()
		defer wg.Done()
		var written int
	loop:
		for {
			select {
			case msg := <-rx:
				fmt.Fprintf(w, "%s\n", string(msg.Data))
				written++
			case <-ctx.Done():
				break loop
			}
		}
	}()
	return nil
}

func (h Handle) Wait() {
	h.wg.Wait()
	return
}
