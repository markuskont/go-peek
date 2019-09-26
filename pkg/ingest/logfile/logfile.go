package logfile

import (
	"context"

	"github.com/ccdcoe/go-peek/pkg/models/consumer"
)

type Consumer struct {
	h      *Handle
	tx     chan *consumer.Message
	ctx    context.Context
	cancel context.CancelFunc
}

// Messages implements consumer.Messager
func (c Consumer) Messages() <-chan *consumer.Message { return c.tx }
