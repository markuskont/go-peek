package run

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ccdcoe/go-peek/pkg/models/consumer"
	"github.com/ccdcoe/go-peek/pkg/models/events"
	"github.com/ccdcoe/go-peek/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func spawnWorkers(
	rx <-chan *consumer.Message,
	workers int,
	spooldir string,
) (<-chan *consumer.Message, *utils.ErrChan) {
	tx := make(chan *consumer.Message, 0)
	errs := utils.NewErrChan(100, "Event parse worker runtime errors")
	var wg sync.WaitGroup
	noparse := func() bool {
		if !viper.GetBool("processor.enabled") {
			log.Debug("all procesor plugins disabled globally, only parsing for timestamps")
			return true
		}
		return false
	}()
	var (
		count uint64
		every = time.NewTicker(3 * time.Second)
	)
	go func() {
		for {
			select {
			case <-every.C:
				log.Infof("Collected %d events", atomic.LoadUint64(&count))
			}
		}
	}()
	go func() {
		defer close(tx)
		defer close(errs.Items)
		mapping := func() map[string]events.Atomic {
			out := make(map[string]events.Atomic)
			for _, event := range events.Atomics {
				if src := viper.GetStringSlice(
					fmt.Sprintf("stream.%s.kafka.topic", event.String()),
				); len(src) > 0 {
					for _, item := range src {
						out[item] = event
					}
				}
			}
			return out
		}()
		kafkaTopicToEvent := func(topic string) events.Atomic {
			if val, ok := mapping[topic]; ok {
				return val
			}
			return events.SimpleE
		}
		//defer globalAssetCache.Close()
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				defer log.Tracef("worker %d done", id)

				log.Tracef("Spawning worker %d", id)
				/*
					localAssetCache := intel.NewLocalCache(globalAssetCache, id)
					defer localAssetCache.Close()
				*/

			loop:
				for msg := range rx {
					atomic.AddUint64(&count, 1)
					evType := msg.Event
					switch msg.Type {
					case consumer.Kafka:
						evType = kafkaTopicToEvent(msg.Source)
					}
					if noparse {
						msg.Time = time.Now()
						tx <- msg
						continue loop
					}

					e, err := events.ParseSyslog(msg.Data, evType)
					if err != nil {
						errs.Send(err)
						continue loop
					}
					msg.Time = e.Time()
					msg.Key = evType.String()

					meta := e.GetAsset()
					if meta == nil {
						errs.Send(fmt.Errorf(
							"unable to get meta for event %s",
							string(msg.Data),
						))
						continue loop
					}

					/*
						// Asset checking stuff
						if ip := meta.Asset.IP; ip != nil {
							if val, ok := localAssetCache.GetIP(ip); ok && val.IsAsset {
								meta.Asset = *val.Data
							}
						}
						if meta.Source != nil {
							if ip := meta.Source.IP; ip != nil {
								if val, ok := localAssetCache.GetIP(ip); ok && val.IsAsset && val.Data != nil {
									meta.Source = val.Data
								}
							}
						}
						if meta.Destination != nil {
							if ip := meta.Destination.IP; ip != nil {
								if val, ok := localAssetCache.GetIP(ip); ok && val.IsAsset && val.Data != nil {
									meta.Destination = val.Data
								}
							}
						}
					*/

					/*
						meta.SetDirection()
						if anon {
							// TODO - also get rid of host name fields in message
							meta.Host = meta.Alias
							if meta.Source != nil {
								meta.Source.Host = meta.Source.Alias
							}
							if meta.Destination != nil {
								meta.Destination.Host = meta.Destination.Alias
							}
						}
					*/

					e.SetAsset(*meta)
					modified, err := e.JSONFormat()
					if err != nil {
						errs.Send(err)
						continue loop
					}
					msg.Data = modified
					tx <- msg
				}
			}(i)
		}
		wg.Wait()
	}()
	return tx, errs
}
