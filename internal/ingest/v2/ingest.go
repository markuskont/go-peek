package ingest

import "time"

type Source int

func (s Source) String() string {
	switch s {
	case Logfile:
		return "logfile"
	default:
		return "NA"
	}
}

const (
	Logfile Source = iota
)

// Message is an atomic log entry that is closely modeled after kafka event
type Message struct {
	// Raw message in byte array format
	// Can be original message or processed
	Data []byte

	// Message offset from input
	// e.g. kafka offset or file line number
	Offset int64

	// Enum that maps to supported source module
	// Logfile, unix socket, kafka, redis
	Type Source

	// Textual representation of input source
	// e.g. source file, kafka topic, redis key, etc
	// can also be a hash if source path is too long
	Source string

	// Optional message key, separate from source topic
	// Internal from message, as opposed to external from topic
	// e.g. Kafka key, syslog program, suricata event type, eventlog channel, etc
	Key string

	// Optional timestamp from source message
	// Can default to time.Now() if timestamp is missing or not parsed
	// Should default to time.Now() if message is consumed online
	Time time.Time
}

type Offsets struct {
	Beginning, End int64
}

func (o Offsets) Len() int64 {
	return (o.End - o.Beginning) + 1
}
