package socat

import (
	"encoding/json"

	"github.com/threefoldtech/0-core/base/nft"
	"github.com/threefoldtech/0-core/base/pm"
)

const (
	cmdSocatList   = "socat.list"
	cmdSocalReseve = "socat.reserve"
)

func init() {
	pm.RegisterBuiltIn(cmdSocatList, mgr.list)
	pm.RegisterBuiltIn(cmdSocalReseve, mgr.reserve)
}

func (s *socatManager) list(cmd *pm.Command) (interface{}, error) {
	s.rm.Lock()
	defer s.rm.Unlock()

	matches, err := nft.Find(nft.And{
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

	var rules []string
	for _, rule := range matches {
		rules = append(rules, rule.Body)
	}

	return rules, nil
}

func (s *socatManager) reserve(cmd *pm.Command) (interface{}, error) {
	var query struct {
		Number int `json:"number"`
	}

	if err := json.Unmarshal(*cmd.Arguments, &query); err != nil {
		return nil, pm.BadRequestError(err)
	}

	if query.Number == 0 {
		return nil, pm.BadRequestError("reseve number cannot be zero")
	}

	return s.Reserve(query.Number)
}

func (s *socatManager) resolveAction(cmd *pm.Command) (interface{}, error) {
	var query struct {
		URL string `json:"url"`
	}

	if err := json.Unmarshal(*cmd.Arguments, &query); err != nil {
		return nil, pm.BadRequestError(err)
	}

	return s.ResolveURL(query.URL)
}
