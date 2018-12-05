package events

import (
	"encoding/json"
	"fmt"
	"time"
)

type Snoopy struct {
	Syslog

	Cmd      string `json:"cmd"`
	Filename string `json:"filename"`
	Cwd      string `json:"cwd"`
	Tty      string `json:"tty"`
	Sid      string `json:"sid"`
	Gid      string `json:"gid"`
	Group    string `json:"group"`
	UID      string `json:"uid"`
	Username string `json:"username"`
	SSH      struct {
		DstPort string `json:"dst_port"`
		DstIP   string `json:"dst_ip"`
		SrcPort string `json:"src_port"`
		SrcIP   string `json:"src_ip"`
	} `json:"ssh"`
	Login string `json:"login"`

	GameMeta *GameMeta `json:"gamemeta,omitempty"`
}

func (s Snoopy) JSON() ([]byte, error) {
	return json.Marshal(s)
}

func (s Snoopy) Source() Source {
	return Source{
		Host: s.Host,
		IP:   s.IP.String(),
	}
}

func (s *Snoopy) Rename(pretty string) {
	s.Host = pretty
}

func (s Snoopy) Key() string {
	return s.Filename
}

func (s Snoopy) EventTime() time.Time {
	return s.Timestamp
}
func (s Snoopy) GetSyslogTime() time.Time {
	return s.Timestamp
}
func (s Snoopy) SaganString() string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s|%s",
		s.Host,
		s.Facility,
		s.Severity,
		s.Severity,
		s.Program,
		s.GetSyslogTime().Format(saganDateFormat),
		s.GetSyslogTime().Format(saganTimeFormat),
		s.Program,
		s.Cmd,
	)
}

func (s *Snoopy) Meta(topic, iter string) Event {
	s.GameMeta = &GameMeta{
		Iter:  iter,
		Topic: topic,
	}
	return s
}
