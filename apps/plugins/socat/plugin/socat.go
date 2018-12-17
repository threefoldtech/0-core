package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"

	"github.com/threefoldtech/0-core/apps/plugins/nft"
	"github.com/threefoldtech/0-core/apps/plugins/socat"

	"github.com/op/go-logging"
	"github.com/patrickmn/go-cache"
)

func main() {} //silence error

const (
	addressAll = "0.0.0.0"
)

var (
	log = logging.MustGetLogger("socat")

	mgr socatManager
	_   socat.API = (*socatManager)(nil) //validation

	//Plugin plugin entry point
	Plugin = plugin.Plugin{
		Name:    "socat",
		Version: "1.0",
		Open: func(api plugin.API) (err error) {
			return newSocatManager(&mgr, api)
		},
		API: func() interface{} {
			return &mgr
		},
		Actions: map[string]pm.Action{
			"list":    mgr.list,
			"reserve": mgr.reserve,
		},
	}
)

var (
	defaultProtocols = []string{"tcp"}
	validProtocols   = map[string]struct{}{
		"tcp": struct{}{},
		"udp": struct{}{},
	}
)

type socatManager struct {
	api plugin.API
	nft nft.API

	rm    sync.Mutex
	rules map[int]rule

	sm       sync.Mutex
	reserved *cache.Cache
}

func newSocatManager(mgr *socatManager, api plugin.API) error {
	p, err := api.Plugin("nft")
	if err != nil {
		return err
	}
	nft, ok := p.(nft.API)
	if !ok {
		return fmt.Errorf("wrong nft api")
	}

	mgr.api = api
	mgr.nft = nft
	mgr.rules = make(map[int]rule)
	mgr.reserved = cache.New(2*time.Minute, 1*time.Minute)

	return mgr.init()
}

func (s *socatManager) init() error {
	return s.monitorIPChangesUpdateSocat()
}

//SetPortForward create a single port forward from host(port), to ip(addr) and dest(port) in this namespace
//The namespace is used to group port forward rules so they all can get terminated
//with one call later.
func (s *socatManager) SetPortForward(namespace string, ip string, host string, dest int) error {
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

	if err := s.nft.Apply(set); err != nil {
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
func (s *socatManager) RemovePortForward(namespace string, host string, dest int) error {
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

	if err := s.nft.DropRules(set); err != nil {
		return err
	}

	delete(s.rules, src.port)
	return nil
}

//RemoveAll remove all port forwrards that were created in this namespace.
func (s *socatManager) RemoveAll(namespace string) error {
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

	if err := s.nft.DropRules(set); err != nil {
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
func (s *socatManager) Reserve(n int) ([]int, error) {
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
	if err := s.api.Internal("info.port", nil, &ports); err != nil {
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
