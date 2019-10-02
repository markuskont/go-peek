package logfile

import (
	"fmt"
	"time"

	"github.com/ccdcoe/go-peek/pkg/models/consumer"
	"github.com/ccdcoe/go-peek/pkg/utils"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	Paths       []string
	StatFunc    StatFileIntervalFunc
	StatWorkers int

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
	if c.StatFunc == nil {
		log.Tracef(
			"File interval stat function missing for %+v, initializing empty interval",
			c.Paths,
		)
		c.StatFunc = func(first, last []byte) (utils.Interval, error) {
			return utils.Interval{
				Beginning: time.Time{},
				End:       time.Time{},
			}, nil
		}
	}
	if c.StatWorkers < 1 {
		c.StatWorkers = 1
	}
	return nil
}

type Consumer struct {
	h    []*Handle
	tx   chan *consumer.Message
	stat StatFileIntervalFunc
}

func NewConsumer(c *Config) (*Consumer, error) {
	if c == nil {
		return nil, fmt.Errorf("logfile consumer is missing config")
	}
	l := &Consumer{
		tx: make(chan *consumer.Message, 0),
	}
	return l, nil
}

// Messages implements consumer.Messager
func (c Consumer) Messages() <-chan *consumer.Message { return c.tx }
