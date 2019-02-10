package main

import (
	"encoding/json"

	"github.com/threefoldtech/0-core/apps/plugins/nft"
	"github.com/threefoldtech/0-core/base/pm"
)

func (s *socatManager) list(ctx pm.Context) (interface{}, error) {
	s.rm.Lock()
	defer s.rm.Unlock()

	matches, err := s.nft().Find(nft.And{
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
	// return nil
	// type simpleRule struct {
	// 	From string `json:"from"`
	// 	To   string `json:"to"`
	// }

	// m := make(map[int]simpleRule)
	// for s, r := range s.rules {
	// 	m[s] = simpleRule{
	// 		From: fmt.Sprintf("%s:%d", r.source.ip, r.source.port),
	// 		To:   fmt.Sprintf("%s:%d", r.ip, r.port),
	// 	}
	// }

	// return m, nil
}

func (s *socatManager) reserve(ctx pm.Context) (interface{}, error) {
	var query struct {
		Number int `json:"number"`
	}
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &query); err != nil {
		return nil, pm.BadRequestError(err)
	}

	if query.Number == 0 {
		return nil, pm.BadRequestError("reseve number cannot be zero")
	}

	return s.Reserve(query.Number)
}

func (s *socatManager) resolveAction(ctx pm.Context) (interface{}, error) {
	var query struct {
		URL string `json:"url"`
	}
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &query); err != nil {
		return nil, pm.BadRequestError(err)
	}

	return s.ResolveURL(query.URL)
}
