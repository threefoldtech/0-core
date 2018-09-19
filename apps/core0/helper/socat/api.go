package socat

import (
	"encoding/json"
	"fmt"

	"github.com/threefoldtech/0-core/base/pm"
)

const (
	cmdSocatList   = "socat.list"
	cmdSocalReseve = "socat.reserve"
)

func init() {
	pm.RegisterBuiltIn(cmdSocatList, socat.list)
	pm.RegisterBuiltIn(cmdSocalReseve, socat.reserve)
}

func (s *socatApi) list(cmd *pm.Command) (interface{}, error) {
	s.rm.Lock()
	defer s.rm.Unlock()

	type simpleRule struct {
		From string `json:"from"`
		To   string `json:"to"`
	}

	m := make(map[int]simpleRule)
	for s, r := range s.rules {
		m[s] = simpleRule{
			From: fmt.Sprintf("%s:%d", r.source.ip, r.source.port),
			To:   fmt.Sprintf("%s:%d", r.ip, r.port),
		}
	}

	return m, nil
}

func (s *socatApi) reserve(cmd *pm.Command) (interface{}, error) {
	var query struct {
		Numer int `json:"number"`
	}

	if err := json.Unmarshal(*cmd.Arguments, &query); err != nil {
		return nil, pm.BadRequestError(err)
	}

	if query.Numer == 0 {
		return nil, pm.BadRequestError("reseve number cannot be zero")
	}

	return s.Reserve(query.Numer)
}
