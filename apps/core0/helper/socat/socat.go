package socat

import (
	"fmt"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/threefoldtech/0-core/base/nft"
	"github.com/threefoldtech/0-core/base/pm"

	logging "github.com/op/go-logging"
	cache "github.com/patrickmn/go-cache"
)

func main() {} //silence error

const (
	addressAll = "0.0.0.0"
)

var (
	log = logging.MustGetLogger("socat")

	mgr socatManager
)

var (
	portMatch        = regexp.MustCompile(`dport\s+(\d+)`)
	defaultProtocols = []string{"tcp"}
	validProtocols   = map[string]struct{}{
		"tcp": struct{}{},
		"udp": struct{}{},
	}
)

type socatManager struct {
	rm sync.Mutex

	sm       sync.Mutex
	reserved *cache.Cache
}

func init() {
	mgr.reserved = cache.New(2*time.Minute, 1*time.Minute)
}

//ValidHost checks if the host string is valid
//Valid hosts is (port, ip:port, or device:port)
func ValidHost(host string) bool {
	_, err := getSource(host)
	return err == nil
}

//SetPortForward create a single port forward from host(port), to ip(addr) and dest(port) in this namespace
//The namespace is used to group port forward rules so they all can get terminated
//with one call later.
func (s *socatManager) SetPortForward(ns NS, ip string, host string, dest int) error {
	s.rm.Lock()
	defer s.rm.Unlock()

	src, err := getSource(host)
	if err != nil {
		return err
	}

	matches, err := nft.Find(nft.And{
		&nft.TableFilter{
			Table: "nat",
		},
		&nft.ChainFilter{
			Chain: "pre",
		},
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
	})

	if err != nil {
		return err
	}

	//NOTE: this will only check if the port is used for port forwarding
	//if a port on the host is using this port it will get masked out
	if len(matches) > 0 {
		return fmt.Errorf("port %d already in use", src.port)
	}

	rule := rule{
		ns:     ns,
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

	s.sm.Lock()
	defer s.sm.Unlock()
	s.reserved.Delete(fmt.Sprint(src.port))

	return nil
}

func forwardID(namespace string, host int, dest int) string {
	return fmt.Sprintf("socat-%v-%v-%v", namespace, host, dest)
}

//RemovePortForward removes a single port forward
func (s *socatManager) RemovePortForward(ns NS, host string, dest int) error {
	s.rm.Lock()
	defer s.rm.Unlock()
	src, err := getSource(host)
	if err != nil {
		return err
	}

	matches, err := nft.Find(nft.And{
		&nft.TableFilter{
			Table: "nat",
		},
		&nft.ChainFilter{
			Chain: "pre",
		},
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
		&nft.MarkFilter{
			Mark: uint32(ns),
		},
	})

	if err != nil {
		return err
	}

	for _, rule := range matches {
		if err := nft.Drop(rule.Family, rule.Table, rule.Chain, rule.Handle); err != nil {
			return err
		}
	}

	return nil
}

//RemoveAll remove all port forwrards that were created in this namespace.
func (s *socatManager) RemoveAll(ns NS) error {
	s.rm.Lock()
	defer s.rm.Unlock()

	matches, err := nft.Find(nft.And{
		&nft.TableFilter{
			Table: "nat",
		},
		&nft.ChainFilter{
			Chain: "pre",
		},
		&nft.MarkFilter{
			Mark: uint32(ns),
		},
	})

	if err != nil {
		return err
	}

	for _, rule := range matches {
		if err := nft.Drop(rule.Family, rule.Table, rule.Chain, rule.Handle); err != nil {
			return err
		}
	}

	return nil
}

func (s *socatManager) List(ns NS) (map[string]int, error) {
	s.rm.Lock()
	defer s.rm.Unlock()

	matches, err := nft.Find(nft.And{
		&nft.TableFilter{
			Table: "nat",
		},
		&nft.ChainFilter{
			Chain: "pre",
		},
		&nft.MarkFilter{
			Mark: uint32(ns),
		},
	})

	if err != nil {
		return nil, err
	}

	rules := make(map[uint64]*rule)

	for _, ruleBody := range matches {
		parsed, err := getRuleFromNFTRule(ruleBody.Body)
		if err != nil {
			return nil, err
		}

		func(parsed rule) {
			if r, ok := rules[parsed.source.port]; ok {
				r.source.protocols = append(
					r.source.protocols,
					parsed.source.protocols...,
				)
			} else {
				rules[parsed.source.port] = &parsed
			}

		}(parsed)
	}

	results := make(map[string]int)
	for _, rule := range rules {
		results[rule.source.String()] = rule.port
	}

	return results, nil
}

func (s *socatManager) getForwardedPorts() (map[uint16]struct{}, error) {
	rules, err := nft.Find(nft.And{
		&nft.TableFilter{
			Table: "nat",
		},
		&nft.ChainFilter{
			Chain: "pre",
		},
	})

	if err != nil {
		return nil, err
	}
	ports := map[uint16]struct{}{}
	for _, rule := range rules {
		m := portMatch.FindStringSubmatch(rule.Body)
		if len(m) != 2 {
			continue
		}

		port, err := strconv.ParseUint(m[1], 10, 16)
		if err != nil {
			return nil, err
		}

		ports[uint16(port)] = struct{}{}
	}

	return ports, err
}

//Reserve reseves the first n number of ports, and return the reserved ports
//reseved ports are reserved only for around 2 min, after that a new reserve
//call can return the same ports.
func (s *socatManager) Reserve(n int) ([]uint16, error) {
	//get all listening tcp ports
	type portInfo struct {
		Network string `json:"network"`
		Port    uint16 `json:"port"`
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

	used := make(map[uint16]struct{})

	for _, port := range ports {
		if port.Network == "tcp" {
			used[port.Port] = struct{}{}
		}
	}

	s.rm.Lock()
	defer s.rm.Unlock()

	forwarded, err := s.getForwardedPorts()
	if err != nil {
		return nil, err
	}

	for port := range forwarded {
		used[port] = struct{}{}
	}

	s.sm.Lock()
	defer s.sm.Unlock()

	//used is now filled with all assigned system ports (except reserved)
	//we can safely find the first port that is not used, and not in reseved and add it to
	//the result list
	var result []uint16
	var p uint16 = 1024
	for i := 0; i < n; i++ {
		for ; p <= 65535; p++ { //i know last valid port is at 65535, but check code below
			if _, ok := used[p]; ok {
				continue
			}

			if _, ok := s.reserved.Get(fmt.Sprint(p)); ok {
				continue
			}

			break
		}

		if p == 65535 {
			return result, fmt.Errorf("pool is exhausted")
		}

		s.reserved.Set(fmt.Sprint(p), nil, cache.DefaultExpiration)
		result = append(result, p)
	}

	return result, nil
}
