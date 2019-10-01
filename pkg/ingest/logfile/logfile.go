package logfile

import (
	"fmt"

	"github.com/ccdcoe/go-peek/pkg/models/consumer"
	"github.com/ccdcoe/go-peek/pkg/utils"
)

type Config struct {
	Path string
}

func (c *Config) Validate() error {
	if c.Path != "" && !utils.StringIsValidDir(c.Path) {
		return fmt.Errorf("%s is not a valid directory", c.Path)
	}
	return nil
}

type Consumer struct {
	h    []*Handle
	tx   chan *consumer.Message
	Stat StatFileIntervalFunc
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
