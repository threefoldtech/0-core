package main

import (
	"fmt"
	"io/ioutil"
	"os"

	logging "github.com/op/go-logging"
)

const (
	//NFTDebug if true, nft files will not be deleted for inspection
	NFTDebug = false
)

var (
	log = logging.MustGetLogger("nft")
)

//ApplyFromFile applies nft rules from a file
func (m *manager) ApplyFromFile(cfg string) error {
	_, err := m.api.System("nft", "-f", cfg)
	return err
}

//Apply (merge) nft rules
func (m *manager) Apply(nft Nft) error {
	data, err := nft.MarshalText()
	if err != nil {
		return err
	}
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
		if !NFTDebug {
			os.RemoveAll(f.Name())
		}
	}()

	if _, err := f.Write(data); err != nil {
		return err
	}
	f.Close()
	log.Debugf("nft applying: %s", f.Name())
	return m.ApplyFromFile(f.Name())
}

//findRules validate that the sub is part of the ruleset, and fill the
//rules handle with the values from the ruleset
func findRules(ruleset, sub Nft) (Nft, error) {
	for tn, t := range sub {
		currenttable, ok := ruleset[tn]
		if !ok {
			return nil, fmt.Errorf("table %s not found", tn)
		}

		for cn, c := range t.Chains {
			currentchain, ok := currenttable.Chains[cn]
			if !ok {
				return nil, fmt.Errorf("chain %s not found in table %s", cn, tn)
			}
			for r := range c.Rules {
				for _, rr := range currentchain.Rules {
					if rr.Body == c.Rules[r].Body {
						c.Rules[r].Handle = rr.Handle
						break
					}
				}
				if c.Rules[r].Handle == 0 {
					return nil, fmt.Errorf("rule '%s' not found", c.Rules[r].Body)
				}
			}
		}
	}

	return sub, nil
}

//DropRules removes nft rules from a file
func (m *manager) DropRules(sub Nft) error {
	ruleset, err := m.Get()

	if err != nil {
		return err
	}

	sub, err = findRules(ruleset, sub)
	if err != nil {
		return err
	}
	// Two loops to achieve all or nothing
	for tn, t := range sub {
		for cn, c := range t.Chains {
			for _, r := range c.Rules {
				if err := m.Drop(t.Family, tn, cn, r.Handle); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

//Drop drops a single rule given a handle
func (m *manager) Drop(family Family, table, chain string, handle int) error {
	_, err := m.api.System("nft", "delete", "rule", string(family), table, chain, "handle", fmt.Sprint(handle))
	return err
}

//Get gets current nft ruleset
func (m *manager) Get() (Nft, error) {
	//NOTE: YES --numeric MUST BE THERE 2 TIMES, PLEASE DO NOT REMOVE
	job, err := m.api.System("nft", "--json", "--handle", "--numeric", "--numeric", "list", "ruleset")
	if err != nil {
		return nil, err
	}

	return Parse(job.Streams.Stdout())
}
