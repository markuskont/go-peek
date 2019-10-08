package uxsock

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/ccdcoe/go-peek/pkg/models/consumer"
	"github.com/ccdcoe/go-peek/pkg/models/events"
	"github.com/ccdcoe/go-peek/pkg/utils"
)

const (
	bufsize = 32 * 1024 * 1024
)

type ErrSocketCreate struct {
	Path string
	Err  error
}

func (e ErrSocketCreate) Error() string {
	return fmt.Sprintf("Socket: %s Error: [%s]", e.Path, e.Err)
}

type Config struct {
	Sockets []string
	MapFunc func(string) events.Atomic
	Ctx     context.Context
}

func (c *Config) Validate() error {
	if c.Sockets == nil || len(c.Sockets) == 0 {
		return fmt.Errorf("Unix socket input has no paths configured")
	}
	if c.MapFunc == nil {
		c.MapFunc = func(string) events.Atomic {
			return events.SimpleE
		}
	}
	if c.Ctx == nil {
		c.Ctx = context.Background()
	}
	return nil
}

type handle struct {
	path     string
	listener *net.UnixListener
	atomic   events.Atomic
}

type Consumer struct {
	h        []*handle
	tx       chan *consumer.Message
	ctx      context.Context
	stoppers utils.WorkerStoppers
	wg       *sync.WaitGroup
	errs     *utils.ErrChan
	timeouts int
}

func NewConsumer(c *Config) (*Consumer, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	con := &Consumer{
		tx:   make(chan *consumer.Message, 0),
		ctx:  c.Ctx,
		h:    make([]*handle, 0),
		errs: utils.NewErrChan(100, "uxsock consume"),
	}
	for _, f := range c.Sockets {
		sock, err := net.Listen("unix", f)
		if err != nil {
			return nil, &ErrSocketCreate{
				Path: f,
				Err:  err,
			}
		}
		con.h = append(con.h, &handle{
			path:     f,
			listener: sock.(*net.UnixListener),
			atomic:   c.MapFunc(f),
		})
	}
	var wg sync.WaitGroup
	con.stoppers = utils.NewWorkerStoppers(len(con.h))
	con.wg = &wg
	go func() {
		<-con.ctx.Done()
		con.close()
	}()
	go func() {
		defer con.wg.Wait()
		defer close(con.tx)
		for i, h := range con.h {
			con.wg.Add(1)
			go func(id int, ctx context.Context, h handle) {
				defer socketCleanUp(h.path)
				defer con.wg.Done()
			loop:
				for {
					select {
					case <-ctx.Done():
						break loop
					default:
						h.listener.SetDeadline(time.Now().Add(1e9))
						c, err := h.listener.Accept()
						if err != nil {
							if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
								con.timeouts++
								continue loop
							}
							con.errs.Send(err)
						}
						go scanSocketToChan(consumer.Message{
							Offset:    -1,
							Partition: -1,
							Type:      consumer.UxSock,
							Event:     h.atomic,
							Source:    h.path,
							Key:       "",
							Time:      time.Now(),
						}, c, con.tx, context.TODO(), con.errs)
					}
				}
			}(i, con.stoppers[i].Ctx, *h)
		}
	}()
	return con, nil
}

func (c Consumer) Messages() <-chan *consumer.Message { return c.tx }
func (c Consumer) Timeouts() int                      { return c.timeouts }

func (c Consumer) close() error {
	if c.stoppers == nil || len(c.stoppers) == 0 || c.wg == nil {
		return fmt.Errorf("Cannot close unix socket consumer, not properly instanciated")
	}
	c.stoppers.Close()
	c.wg.Wait()
	return nil
}

func scanSocketToChan(
	tpl consumer.Message,
	raw io.Reader,
	tx chan<- *consumer.Message,
	ctx context.Context,
	errs *utils.ErrChan,
) {
	scanner := bufio.NewScanner(raw)
	buf := make([]byte, 0, bufsize)
	scanner.Buffer(buf, bufsize)
loop:
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			break loop
		default:
			tx <- &consumer.Message{
				Data:      utils.DeepCopyBytes(scanner.Bytes()),
				Offset:    tpl.Offset,
				Partition: tpl.Partition,
				Type:      tpl.Type,
				Event:     tpl.Event,
				Source:    tpl.Source,
				Key:       tpl.Key,
				Time:      tpl.Time,
			}
		}
	}

	if err := scanner.Err(); err != nil {
		errs.Send(err)
	}
	return
}

func socketCleanUp(p string) {
	_, err := os.Stat(p)
	if err == nil {
		os.Remove(p)
	}
}

/*
func tester() {
	events, _ := net.Listen("unix", "/tmp/test.sock")
}
*/
