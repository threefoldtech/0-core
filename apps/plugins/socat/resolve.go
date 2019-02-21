package socat

import (
	"fmt"
	"net"
	"net/url"
	"regexp"

	"github.com/threefoldtech/0-core/apps/plugins/nft"
)

var (
	dnatMatch = regexp.MustCompile(`dnat to ([^\s]+)`)
)

//getInterfaceMatch returns the first interface that has the given ip
//as <name>, <address>, error
//error return if no match is found
func getInterfaceMatch(ip string) (string, error) {
	nics, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, nic := range nics {
		var addrs []net.Addr
		addrs, err = nic.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			if addr, ok := addr.(*net.IPNet); ok {
				if addr.IP.String() == ip {
					return nic.Name, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no match found")
}

//Resolve resolves an address of the form <ip>:<port> to a direct address to the endpoint
//IF
// - the ip address is a local address of this machine
// - port has a forwarding rule
//ELSE
// - return address unchanged
func (s *socatManager) Resolve(address string) string {
	return s.resolve(address, s.nft())
}

func (s *socatManager) resolve(address string, api nft.API) string {
	log.Debugf("resolving: %s", address)
	src, err := getSource(address)
	if err != nil {
		return address
	}

	log.Debugf("source is: %v", src)

	if len(src.ip) == 0 {
		//we have this check here because getSource allows the <port> <ip>:<port> syntax as well
		return address
	}

	nic, err := getInterfaceMatch(src.ip)
	if err != nil {
		return address
	}
	log.Debugf("nic is: %v", nic)
	//address points to a local address, so it can be forwarded.
	//we need to find the rule that matches this address.
	//this can be done by matching a rule in the nat table that
	//uses those ports

	filter := nft.And{
		nft.Or{
			&nft.IntMatchFilter{
				Name:  "tcp",
				Field: "dport",
				Value: src.port,
			},
			&nft.IntMatchFilter{
				Name:  "udp",
				Field: "dport",
				Value: src.port,
			},
		},
		&nft.TableFilter{
			Table: "nat",
		},
		&nft.ChainFilter{
			Chain: "pre",
		},
	}

	rules, _ := api.Find(filter)
	if len(rules) == 0 {
		return address
	}
	m := dnatMatch.FindStringSubmatch(rules[0].Body)
	if len(m) != 2 {
		//that should never happen
		return address
	}

	return m[1]
}

//ResolveURL rewrites a url to a direct address to the end point. Return original url
//if no forwarding rule configured that matches the given address
//note, the url host part must be an ip, can't use host names
func (s *socatManager) ResolveURL(raw string) (string, error) {
	return s.resolveURL(raw, s.nft())
}

func (s *socatManager) resolveURL(raw string, api nft.API) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return raw, err
	}
	host := u.Host
	if u.Port() == "" {
		port, err := net.LookupPort("tcp", u.Scheme)
		if err != nil {
			return raw, err
		}
		host = fmt.Sprintf("%s:%d", host, port)
	}

	u.Host = s.resolve(host, api)

	return u.String(), nil
}
