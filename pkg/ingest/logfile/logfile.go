package logfile

import (
	"fmt"
	"time"

	"github.com/ccdcoe/go-peek/pkg/models/consumer"
	"github.com/ccdcoe/go-peek/pkg/utils"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	Paths []string

	Stat struct {
		Enabled bool
		Func    StatFileIntervalFunc
		// Workers for statting timestamps if AsyncConsume is true
		// Optimal for throughput is number of CPU threads - 2/3
		Workers int
	}

	Consume struct {
		Async   bool
		Workers int
	}

	//TODO
	files   []string
	pattern string
}

func (c *Config) Validate() error {
	if c.Paths == nil {
		return fmt.Errorf("File input module is missing root paths")
	}
	for _, pth := range c.Paths {
		if !utils.StringIsValidDir(pth) {
			return fmt.Errorf("%s is not a valid directory", pth)
		}
	}
	if c.Stat.Enabled && c.Stat.Func == nil {
		log.Tracef(
			"File interval stat function missing for %+v, initializing empty interval",
			c.Paths,
		)
		c.Stat.Func = func(first, last []byte) (utils.Interval, error) {
			return utils.Interval{
				Beginning: time.Time{},
				End:       time.Time{},
			}, nil
		}
	}
	if c.Stat.Workers < 1 {
		c.Stat.Workers = 1
	}
	if c.Consume.Workers < 1 {
		c.Consume.Workers = 1
	}
	return nil
}

type Consumer struct {
	h    []*Handle
	tx   chan *consumer.Message
	conf Config
}

func NewConsumer(c *Config) (*Consumer, error) {
	if c == nil {
		return nil, fmt.Errorf("logfile consumer is missing config")
	}
	l := &Consumer{
		h:    make([]*Handle, 0),
		tx:   make(chan *consumer.Message, 0),
		conf: *c,
	}
	for i, dir := range c.Paths {
		log.WithFields(log.Fields{
			"workers": l.conf.Stat.Workers,
			"dir":     dir,
		}).Tracef("%d - invoking async stat", i)
		files, err := AsyncStatAll(dir, l.conf.Stat.Func, l.conf.Stat.Workers)
		if err != nil {
			return nil, err
		}
		l.h = append(l.h, files...)
	}
	return l, nil
}

// Messages implements consumer.Messager
func (c Consumer) Messages() <-chan *consumer.Message { return c.tx }
