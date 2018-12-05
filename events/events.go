package events

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"time"
)

const saganDateFormat = "2006-01-02"
const saganTimeFormat = "15:04:05"

type stringIP struct{ net.IP }

func (t *stringIP) UnmarshalJSON(b []byte) error {
	raw, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}
	t.IP = net.ParseIP(raw)
	return err
}

type Source struct {
	Host, IP string
}

type Event interface {
	JSON() ([]byte, error)
	Source() Source
	Rename(string)
	Key() string
	GetEventTime() time.Time
	GetSyslogTime() time.Time
	SaganString() string
	Meta(string, string) Event
}

type EventRenamer interface {
	Rename(string)
}
type EventJsonDumper interface {
	JSON() ([]byte, error)
}
type EventSourcer interface {
	Source() Source
}
type EventIdentifier interface {
	Key() string
}
type EventTimeReporter interface {
	GetEventTime() time.Time
	GetSyslogTime() time.Time
}

func NewEvent(topic string, payload []byte) (Event, error) {

	switch topic {
	case "syslog":
		var m Syslog
		if err := json.Unmarshal(payload, &m); err != nil {
			return nil, err
		}
		return &m, nil

	case "snoopy":
		var m Snoopy
		if err := json.Unmarshal(payload, &m); err != nil {
			return nil, err
		}
		return &m, nil

	case "suricata":
		var m Eve
		if err := json.Unmarshal(payload, &m); err != nil {
			return nil, err
		}
		return &m, nil

	case "eventlog":
		return NewDynaEventLog(payload)

	default:
		return nil, fmt.Errorf("Unsupported topic %s",
			topic)
	}
}
