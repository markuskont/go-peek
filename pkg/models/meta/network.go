package meta

import "net"

// Network is a long representation of a segment, as used in exercise prep
type Network struct {
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
	net.IPNet
	Name string `json:"name"`
	Desc string `json:"desc"`
}
