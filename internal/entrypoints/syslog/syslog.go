package syslog

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/ccdcoe/go-peek/pkg/models/atomic"
	"github.com/ccdcoe/go-peek/pkg/models/fields"
	"github.com/influxdata/go-syslog/rfc5424"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type UdpMsg struct {
	IP   net.IP
	Data []byte
}

func Entrypoint(cmd *cobra.Command, args []string) {
	addr := fmt.Sprintf("0.0.0.0:%d", viper.GetInt("syslog.port"))
	ServerAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Fatal(err)
	}

	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer ServerConn.Close()
	log.Infof("Spawned syslog server on %s", addr)

	buf := make([]byte, 64*1024)

	msgs := make(chan UdpMsg, 0)

	var wg sync.WaitGroup
	go func() {
		for i := 0; i < viper.GetInt("work.threads"); i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				p := rfc5424.NewParser(rfc5424.WithBestEffort())
				for raw := range msgs {
					msg, err := p.Parse(raw.Data)
					if err != nil {
						log.Error(err)
					}
					s := atomic.Syslog{
						Timestamp: *msg.Timestamp(),
						Host:      *msg.Hostname(),
						Program:   *msg.Appname(),
						Severity:  *msg.SeverityLevel(),
						Facility:  *msg.FacilityLevel(),
						Message:   *msg.Message(),
						IP:        &fields.StringIP{IP: raw.IP},
					}
					switch s.Program {
					case "suricata":
					case "snoopy":
					default:
					}
					j, _ := json.Marshal(s)
					fmt.Println(string(j))
				}
			}(i)
		}
		wg.Wait()
	}()

	for {
		n, ip, err := ServerConn.ReadFromUDP(buf)
		if err != nil {
			log.Error(err)
		}
		msgs <- UdpMsg{
			IP:   ip.IP,
			Data: buf[0:n],
		}
	}
}
