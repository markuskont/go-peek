package meta

import (
	"net"

	"github.com/ccdcoe/go-peek/pkg/models/fields"
)

// NetworkPandasExport is for loading network table as exported from pandas to_json() method
type NetworkPandasExport struct {
	Name         map[int]string           `json:"Name"`
	Abbreviation map[int]string           `json:"Abbreviation"`
	VLAN         map[int]string           `json:"VLAN/Portgroup"`
	IPv4         map[int]fields.StringNet `json:"IPv4"`
	IPv6         map[int]fields.StringNet `json:"IPv6"`
	Desc         map[int]string           `json:"Description"`
	Whois        map[int]string           `json:"WHOIS"`
	Team         map[int]string           `json:"Team"`
}

func (n NetworkPandasExport) Extract() []*Network {
	items := 0
	list := make([]*Network, items)
	return list
}

// Network is a long representation of a segment, as used in exercise prep
type Network struct {
	ID           int       `json:"id"`
	Name         string    `json:"Name"`
	Abbreviation string    `json:"Abbreviation"`
	VLAN         string    `json:"VLAN/Portgroup"`
	IPv4         net.IPNet `json:"IPv4"`
	IPv6         net.IPNet `json:"IPv6"`
	Desc         string    `json:"Description"`
	Whois        string    `json:"WHOIS"`
	Team         string    `json:"Team"`
}

// NetSegment is a shorthand representation of Network, more suitable to be used in de-normalized logging
type NetSegment struct {
	ID int `json:"id"`
	net.IPNet
	Name string `json:"name"`
}
