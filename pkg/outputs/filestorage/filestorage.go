package filestorage

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ccdcoe/go-peek/pkg/models/consumer"
	"github.com/ccdcoe/go-peek/pkg/utils"
)

var (
	TimeFmt = "20060102150405"
)

type Config struct {
	Directory string
	Combined  string
	Gzip      bool
	Timestamp bool
	Stream    <-chan consumer.Message
}

func (c *Config) Validate() error {
	if c == nil {
		c = &Config{}
	}
	if c.Combined == "" && c.Directory == "" {
		return fmt.Errorf(
			"filestorage module requires either a root directory or explicit destination file for storing events",
		)
	}
	if c.Stream == nil {
		return fmt.Errorf("missing input stream for filestorage")
	}
	return nil
}

type Handle struct {
	errs *utils.ErrChan
	rx   <-chan consumer.Message

	combined  string
	timestamp bool
	gzip      bool
}

func NewHandle(c *Config) (*Handle, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	h := &Handle{
		errs:      utils.NewErrChan(100, "filestorage handle"),
		rx:        c.Stream,
		timestamp: c.Timestamp,
		gzip:      c.Gzip,
	}
	if c.Combined != "" {
		h.combined = c.Combined
	}
	return h, nil
}

func (h Handle) Do(ctx context.Context) error {
	go func() {
		writeSingleFile(h.combined, *h.errs, h.rx, h.timestamp, h.gzip, ctx)
	}()
	return nil
}

func writeSingleFile(
	path string,
	errs utils.ErrChan,
	rx <-chan consumer.Message,
	timestamp bool,
	gz bool,
	ctx context.Context,
) error {
	path, err := utils.ExpandHome(path)
	if err != nil {
		return err
	}
	if timestamp {
		path = fmt.Sprintf("%s.%s", path, time.Now().Format(TimeFmt))
	}
	if gz {
		path = fmt.Sprintf("%s.gz", path)
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
	defer w.Close()
	var written int
loop:
	for {
		select {
		case msg := <-rx:
			if _, err := fmt.Fprintf(w, "%s\n", string(msg.Data)); err != nil {
				errs.Send(err)
			}
			written++
		case <-ctx.Done():
			break loop
		}
	}
	return nil
}
