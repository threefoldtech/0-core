package main

import (
	"fmt"
	"net"
	"strings"
)

type source struct {
	ip        string
	port      int
	protocols []string
}

func (s source) match(proto string) string {
	if addr := net.ParseIP(s.ip); addr != nil {
		if s.ip == "0.0.0.0" {
			return fmt.Sprintf("%s dport %d", proto, s.port)
		}
		return fmt.Sprintf("ip saddr %s %s dport %d", s.ip, proto, s.port)
	} else if _, _, err := net.ParseCIDR(s.ip); err == nil {
		//NETWORK
		return fmt.Sprintf("ip saddr %s %s dport %d", s.ip, proto, s.port)

	}
	//assume interface name
	return fmt.Sprintf("iifname \"%s\" %s dport %d", s.ip, proto, s.port)
}

func (s source) Matches() []string {
	var matches []string
	for _, proto := range s.protocols {
		matches = append(matches, s.match(proto))
	}

	return matches
}

/*
getSource parse source port

source = port
source = address:port
source = source|protocol
protocol = tcp
protocol = udp
protocol = protocol(+protocol)?
address = ip
address = ip/mask
address = nic
address = nic*
*/
func getSource(src string) (source, error) {
	parts := strings.SplitN(src, "|", 2)
	if len(parts) > 2 {
		return source{}, fmt.Errorf("invalid syntax")
	}
	var r = source{
		ip:        addressAll,
		protocols: defaultProtocols,
	}

	src = parts[0]
	if len(parts) == 2 {
		r.protocols = strings.Split(parts[1], "+")
	}

	for _, p := range r.protocols {
		if _, ok := validProtocols[p]; !ok {
			return source{}, fmt.Errorf("invalid protocol '%s'", p)
		}
	}

	parts = strings.SplitN(src, ":", 2)

	if _, err := fmt.Sscanf(parts[len(parts)-1], "%d", &r.port); err != nil {
		return r, err
	}

	if r.port <= 0 || r.port >= 65536 {
		return r, fmt.Errorf("invalid port number")
	}

	if len(parts) == 2 {
		r.ip = parts[0]
	}

	return r, nil
}

//ValidHost checks if the host string is valid
//Valid hosts is (port, ip:port, or device:port)
func ValidHost(host string) bool {
	_, err := getSource(host)
	return err == nil
}

type rule struct {
	ns     string
	source source
	port   int
	ip     string
}

func (r rule) rule(match string) string {
	return fmt.Sprintf("ip daddr @host %s dnat to %s:%d", match, r.ip, r.port)
}

func (r rule) Rules() []string {
	var rules []string
	for _, match := range r.source.Matches() {
		rules = append(rules, r.rule(match))
	}

	return rules
}
