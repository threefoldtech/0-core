package socat

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

//getInterfaceMatch returns the first interface that has the given ip
//as <name>, <address>, error
//error return if no match is found
func getInterfaceMatch(ip string) (name string, address net.IP, err error) {
	nics, err := net.Interfaces()
	if err != nil {
		return
	}
	for _, nic := range nics {
		var addrs []net.Addr
		addrs, err = nic.Addrs()
		if err != nil {
			return
		}

		for _, addr := range addrs {
			if addr, ok := addr.(*net.IPNet); ok {
				if addr.IP.String() == ip {
					return nic.Name, addr.IP, nil
				}
			}
		}
	}

	err = fmt.Errorf("no match found")
	return
}

//Resolve resolves an address of the form <ip>:<port> to a direct address to the endpoint
//IF
// - the ip address is a local address of this machine
// - port has a forwarding rule
//ELSE
// - return address unchanged
func Resolve(address string) string {
	src, err := getSource(address)
	if err != nil {
		return address
	}

	if len(src.ip) == 0 {
		//we have this check here because getSource allows the <port> <ip>:<port> syntax as well
		return address
	}

	nic, ip, err := getInterfaceMatch(src.ip)
	if err != nil {
		return address
	}

	lock.Lock()
	dst, ok := rules[src.port]
	lock.Unlock()

	if !ok {
		return address
	}

	/*
		the actual source ip can be as follows:
		  - empty/0.0.0.0
		  - IP
		  - IP/MASK (CIDR)
		  - interface name
		  - prefix* (partial match of interface name)
	*/
	rewritten := fmt.Sprintf("%s:%d", dst.ip, dst.port)
	if len(dst.source.ip) == 0 || // empty
		dst.source.ip == "0.0.0.0" || // 0.0.0.0 ip
		dst.source.ip == nic || // exact nic match
		dst.source.ip == ip.String() { // exac ip match

		return rewritten
	} else if _, network, err := net.ParseCIDR(dst.source.ip); err == nil {
		if network.Contains(ip) {
			//source ip is in network
			return rewritten
		}

	} else if i := strings.Index(dst.source.ip, "*"); i > 0 {
		if strings.HasPrefix(nic, dst.source.ip[:i]) {
			return rewritten
		}
	}

	return address
}

//ResolveURL rewrites a url to a direct address to the end point. Return original url
//if no forwarding rule configured that matches the given address
//note, the url host part must be an ip, can't use host names
func ResolveURL(raw string) (string, error) {
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

	u.Host = Resolve(host)

	return u.String(), nil
}
