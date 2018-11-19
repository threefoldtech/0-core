package socat

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/threefoldtech/0-core/base/nft"
	"github.com/threefoldtech/0-core/base/pm"

	"github.com/op/go-logging"
	"github.com/patrickmn/go-cache"
)

const (
	addressAll = "0.0.0.0"
)

var (
	log   = logging.MustGetLogger("socat")
	socat socatAPI

	defaultProtocols = []string{"tcp"}
	validProtocols   = map[string]struct{}{
		"tcp": struct{}{},
		"udp": struct{}{},
	}
)

type socatAPI struct {
	rm    sync.Mutex
	rules map[int]rule

	sm       sync.Mutex
	reserved *cache.Cache
}

func init() {
	socat.rules = make(map[int]rule)
	socat.reserved = cache.New(2*time.Minute, 1*time.Minute)
}

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

//SetPortForward create a single port forward from host(port), to ip(addr) and dest(port) in this namespace
//The namespace is used to group port forward rules so they all can get terminated
//with one call later.
func (s *socatAPI) SetPortForward(namespace string, ip string, host string, dest int) error {
	s.rm.Lock()
	defer s.rm.Unlock()

	src, err := getSource(host)
	if err != nil {
		return err
	}

	//NOTE: this will only check if the port is used for port forwarding
	//if a port on the host is using this port it will get masked out
	if _, exists := s.rules[src.port]; exists {
		return fmt.Errorf("port already in use")
	}

	rule := rule{
		ns:     forwardID(namespace, src.port, dest),
		source: src,
		port:   dest,
		ip:     ip,
	}

	var rs []nft.Rule
	for _, r := range rule.Rules() {
		rs = append(rs, nft.Rule{Body: r})
	}

	set := nft.Nft{
		"nat": nft.Table{
			Family: nft.FamilyIP,
			Sets: nft.Sets{
				"host": nft.Set{},
			},
			Chains: nft.Chains{
				"pre": nft.Chain{
					Rules: rs,
				},
			},
		},
	}

	if err := nft.Apply(set); err != nil {
		return err
	}

	s.rules[src.port] = rule

	s.sm.Lock()
	defer s.sm.Unlock()
	s.reserved.Delete(fmt.Sprint(src.port))

	return nil
}

func forwardID(namespace string, host int, dest int) string {
	return fmt.Sprintf("socat-%v-%v-%v", namespace, host, dest)
}

//RemovePortForward removes a single port forward
func (s *socatAPI) RemovePortForward(namespace string, host string, dest int) error {
	s.rm.Lock()
	defer s.rm.Unlock()
	src, err := getSource(host)
	if err != nil {
		return err
	}

	rule, ok := s.rules[src.port]
	if !ok {
		return fmt.Errorf("no port forward from host port: %d", src.port)
	}

	if rule.ns != forwardID(namespace, src.port, dest) {
		return fmt.Errorf("permission denied")
	}

	var rs []nft.Rule
	for _, r := range rule.Rules() {
		rs = append(rs, nft.Rule{Body: r})
	}

	set := nft.Nft{
		"nat": nft.Table{
			Family: nft.FamilyIP,
			Chains: nft.Chains{
				"pre": nft.Chain{
					Rules: rs,
				},
			},
		},
	}

	if err := nft.DropRules(set); err != nil {
		return err
	}

	delete(s.rules, src.port)
	return nil
}

//RemoveAll remove all port forwrards that were created in this namespace.
func (s *socatAPI) RemoveAll(namespace string) error {
	s.rm.Lock()
	defer s.rm.Unlock()

	var toDelete []nft.Rule
	var hostPorts []int

	for host, r := range s.rules {
		if !strings.HasPrefix(r.ns, fmt.Sprintf("socat-%s", namespace)) {
			continue
		}

		for _, rs := range r.Rules() {
			toDelete = append(toDelete, nft.Rule{Body: rs})
		}

		hostPorts = append(hostPorts, host)
	}

	if len(toDelete) == 0 {
		return nil
	}

	set := nft.Nft{
		"nat": nft.Table{
			Family: nft.FamilyIP,
			Chains: nft.Chains{
				"pre": nft.Chain{
					Rules: toDelete,
				},
			},
		},
	}

	if err := nft.DropRules(set); err != nil {
		log.Errorf("failed to delete ruleset: %s", err)
		return err
	}

	for _, host := range hostPorts {
		delete(s.rules, host)
	}

	return nil
}

//Reserve reseves the first n number of ports, and return the reserved ports
//reseved ports are reserved only for around 2 min, after that a new reserve
//call can return the same ports.
func (s *socatAPI) Reserve(n int) ([]int, error) {
	//get all listening tcp ports
	type portInfo struct {
		Network string `json:"network"`
		Port    int    `json:"port"`
	}
	var ports []portInfo

	/*
		list ports from local services, we of course can't grantee
		that a service will start listening after listing the ports
		but zos doesn't start any more services (it shouldn't) after
		the initial bootstrap, so we almost safe by using this returned
		list
	*/
	if err := pm.Internal("info.port", nil, &ports); err != nil {
		return nil, err
	}

	used := make(map[int]struct{})

	for _, port := range ports {
		if port.Network == "tcp" {
			used[port.Port] = struct{}{}
		}
	}

	s.rm.Lock()
	defer s.rm.Unlock()

	for port := range s.rules {
		used[port] = struct{}{}
	}

	s.sm.Lock()
	defer s.sm.Unlock()

	//used is now filled with all assigned system ports (except reserved)
	//we can safely find the first port that is not used, and not in reseved and add it to
	//the result list
	var result []int
	p := 1024
	for i := 0; i < n; i++ {
		for ; p <= 65536; p++ { //i know last valid port is at 65535, but check code below
			if _, ok := used[p]; ok {
				continue
			}

			if _, ok := s.reserved.Get(fmt.Sprint(p)); ok {
				continue
			}

			break
		}

		if p == 65536 {
			return result, fmt.Errorf("pool is exhausted")
		}

		s.reserved.Set(fmt.Sprint(p), nil, cache.DefaultExpiration)
		result = append(result, p)
	}

	return result, nil
}
