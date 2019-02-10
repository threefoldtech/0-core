package main

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/threefoldtech/0-core/apps/plugins/socat"
)

var (
	ruleRegex = regexp.MustCompile(`(?:ip saddr ([^\s]+)|iifname "([^"]+)")? (tcp|udp) dport (\d+).+?dnat to ([^:]+):(\d+)`)
)

type source struct {
	ip        string
	port      uint64
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

func (s source) String() string {
	var buf strings.Builder
	if len(s.ip) > 0 && s.ip != "0.0.0.0" {
		buf.WriteString(s.ip)
		buf.WriteRune(':')
	}

	buf.WriteString(fmt.Sprint(s.port))
	if len(s.protocols) != 1 || s.protocols[0] != "tcp" {
		buf.WriteRune('|')
		buf.WriteString(strings.Join(s.protocols, "+"))
	}

	return buf.String()
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

func getRuleFromNFTRule(body string) (r rule, err error) {
	/*
		ip daddr @host iifname "zt*" tcp dport 1028 mark set 0x01000002 dnat to 172.18.0.3:6379
		ip daddr @host iifname "zt*" udp dport 1028 mark set 0x01000002 dnat to 172.18.0.3:6379
		ip daddr @host ip saddr 10.20.100.100 tcp dport 1029 mark set 0x01000002 dnat to 172.18.0.3:6379
		ip daddr @host ip saddr 192.168.0.0/16 udp dport 6378 mark set 0x01000002 dnat to 172.18.0.3:6379
	*/

	match := ruleRegex.FindStringSubmatch(body)
	if len(match) == 0 {
		return r, fmt.Errorf("invalid rule string")
	}

	r.ip = match[5]
	fmt.Sscanf(match[6], "%d", &r.port)
	if len(match[1]) != 0 {
		r.source.ip = match[1]
	} else {
		r.source.ip = match[2]
	}

	r.source.protocols = append(r.source.protocols, match[3])
	fmt.Sscanf(match[4], "%d", &r.source.port)

	return
}

type rule struct {
	ns     socat.NS
	source source
	port   int
	ip     string
}

func (r rule) rule(match string) string {
	return fmt.Sprintf("ip daddr @host %s meta mark set %d dnat to %s:%d", match, r.ns, r.ip, r.port)
}

func (r rule) Rules() []string {
	var rules []string
	for _, match := range r.source.Matches() {
		rules = append(rules, r.rule(match))
	}

	return rules
}
