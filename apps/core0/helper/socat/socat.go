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
	socat socatApi
)

type socatApi struct {
	rm    sync.Mutex
	rules map[int]rule

	sm      sync.Mutex
	reseved *cache.Cache
}

func init() {
	socat.rules = make(map[int]rule)
	socat.reseved = cache.New(2*time.Minute, 1*time.Minute)
}

type source struct {
	ip   string
	port int
}

func (s source) String() string {
	if addr := net.ParseIP(s.ip); addr != nil {
		if s.ip == "0.0.0.0" {
			return fmt.Sprintf("tcp dport %d", s.port)
		}
		return fmt.Sprintf("ip saddr %s tcp dport %d", s.ip, s.port)
	} else if _, _, err := net.ParseCIDR(s.ip); err == nil {
		//NETWORK
		return fmt.Sprintf("ip saddr %s tcp dport %d", s.ip, s.port)

	}
	//assume interface name
	return fmt.Sprintf("iifname \"%s\" tcp dport %d", s.ip, s.port)
}

func getSource(src string) (source, error) {
	parts := strings.SplitN(src, ":", 2)
	var r = source{
		ip: addressAll,
	}

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

func (r rule) Rule() string {
	return fmt.Sprintf("ip daddr @host %s dnat to %s:%d", r.source, r.ip, r.port)
}

//SetPortForward create a single port forward from host(port), to ip(addr) and dest(port) in this namespace
//The namespace is used to group port forward rules so they all can get terminated
//with one call later.
func (s *socatApi) SetPortForward(namespace string, ip string, host string, dest int) error {
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

	r := rule{
		ns:     forwardID(namespace, src.port, dest),
		source: src,
		port:   dest,
		ip:     ip,
	}

	set := nft.Nft{
		"nat": nft.Table{
			Family:   nft.FamilyIP,
			IPv4Sets: []string{"host"},
			Chains: nft.Chains{
				"pre": nft.Chain{
					Rules: []nft.Rule{
						{Body: r.Rule()},
					},
				},
			},
		},
	}

	if err := nft.Apply(set); err != nil {
		return err
	}

	s.rules[src.port] = r

	s.sm.Lock()
	defer s.sm.Unlock()
	s.reseved.Delete(fmt.Sprint(src.port))

	return nil
}

func forwardID(namespace string, host int, dest int) string {
	return fmt.Sprintf("socat-%v-%v-%v", namespace, host, dest)
}

//RemovePortForward removes a single port forward
func (s *socatApi) RemovePortForward(namespace string, host string, dest int) error {
	s.rm.Lock()
	defer s.rm.Unlock()
	src, err := getSource(host)
	if err != nil {
		return err
	}

	rule, ok := s.rules[src.port]
	if !ok {
		return fmt.Errorf("no port forwrard from host port: %d", src.port)
	}

	if rule.ns != forwardID(namespace, src.port, dest) {
		return fmt.Errorf("permission denied")
	}

	set := nft.Nft{
		"nat": nft.Table{
			Family: nft.FamilyIP,
			Chains: nft.Chains{
				"pre": nft.Chain{
					Rules: []nft.Rule{
						{Body: rule.Rule()},
					},
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
func (s *socatApi) RemoveAll(namespace string) error {
	s.rm.Lock()
	defer s.rm.Unlock()

	var todelete []nft.Rule
	var hostPorts []int

	for host, r := range s.rules {
		if !strings.HasPrefix(r.ns, fmt.Sprintf("socat-%s", namespace)) {
			continue
		}

		todelete = append(todelete, nft.Rule{
			Body: r.Rule(),
		})

		hostPorts = append(hostPorts, host)
	}

	if len(todelete) == 0 {
		return nil
	}

	set := nft.Nft{
		"nat": nft.Table{
			Family: nft.FamilyIP,
			Chains: nft.Chains{
				"pre": nft.Chain{
					Rules: todelete,
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
func (s *socatApi) Reserve(n int) ([]int, error) {
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

			if _, ok := s.reseved.Get(fmt.Sprint(p)); ok {
				continue
			}

			break
		}

		if p == 65536 {
			return result, fmt.Errorf("pool is exhausted")
		}

		s.reseved.Set(fmt.Sprint(p), nil, cache.DefaultExpiration)
		result = append(result, p)
	}

	return result, nil
}
