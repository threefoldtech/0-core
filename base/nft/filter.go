package nft

import (
	"encoding/json"
	"net"
	"strings"
)

//NftJsonBlock defines a nft json block
type NftJsonBlock map[string]json.RawMessage

type NftTableBlock struct {
	//{'family': 'ip', 'name': 'nat', 'handle': 0}
	Family Family `json:"family"`
	Name   string `json:"name"`
	Handle int    `json:"handle"`
}

type NftSetBlock struct {
	/*
		{'family': 'ip',
		'name': 'host',
		'table': 'nat',
		'elem': ['10.20.1.1', '172.18.0.1', '172.19.0.1'],
		'type': 'ipv4_addr',
		'handle': 0}
	*/

	Family   Family   `json:"family"`
	Name     string   `json:"name"`
	Table    string   `json:"table"`
	Elements []string `json:"elem"`
	Type     string   `json:"type"`
	Handle   int      `json:"handle"`
}

type NftChainBlock struct {
	/*
		{'hook': 'prerouting',
		'family': 'ip',
		'prio': 0,
		'table': 'nat',
		'name': 'pre',
		'handle': 1,
		'type': 'nat',
		'policy': 'accept'}
	*/
	Hook     string `json:"hook"`
	Family   Family `json:"family"`
	Priority int    `json:"prio"`
	Table    string `json:"table"`
	Name     string `json:"name"`
	Handle   int    `json:"handle"`
	Type     Type   `json:"type"`
	Policy   string `json:"policy"`
}

type NftRuleBlock struct {
	/*
		{'family': 'inet',
		'expr': [{'match': {'right': {'set': ['established', 'related']},
			'left': {'ct': {'key': 'state'}}}},
		{'accept': None}],
		'table': 'filter',
		'handle': 5,
		'chain': 'input'}
	*/
	Family    Family         `json:"family"`
	Expresion []NftJsonBlock `json:"expr"`
	Table     string         `json:"table"`
	Handle    int            `json:"handle"`
	Chain     string         `json:"chain"`
}

//Filter interface
type Filter interface {
	Match(rule *NftRuleBlock) bool
}

//MetaFilter find a rule by meta mark
type MarkFilter struct {
	Mark uint32
}

func (f *MarkFilter) matchMeta(exp json.RawMessage) bool {
	//{'left': {'meta': 'mark'}, 'right': 123}
	var data struct {
		Left struct {
			Meta string `json:"meta"`
		} `json:"left"`
		Right uint32 `json:"right"`
	}

	if err := json.Unmarshal(exp, &data); err != nil {
		return false
	}

	if data.Left.Meta != "mark" {
		return false
	}

	return data.Right == f.Mark
}

func (f *MarkFilter) Match(rule *NftRuleBlock) bool {
	for _, exp := range rule.Expresion {
		for expType, expMessage := range exp {
			if expType == "mangle" {
				if f.matchMeta(expMessage) {
					return true
				}
			}
		}
	}

	return false
}

type MetaMatchFilter struct {
	Name  string
	Value string
}

func (f *MetaMatchFilter) valueMatch(m string) bool {
	if !strings.HasSuffix(m, "*") {
		return m == f.Value
	}

	return strings.HasPrefix(f.Value, strings.TrimRight(m, "*"))
}

func (f *MetaMatchFilter) matchExp(exp json.RawMessage) bool {
	/*
		{"match": {
			"left": {
				"meta": "iifname"
			},
			"right": "vxbackend"
			}
		}
	*/

	var data struct {
		Left struct {
			Meta string `json:"meta"`
		} `json:"left"`
		Right string `json:"right"`
	}

	if err := json.Unmarshal(exp, &data); err != nil {
		return false
	}

	return data.Left.Meta == f.Name && f.valueMatch(data.Right)
}

func (f *MetaMatchFilter) Match(rule *NftRuleBlock) bool {
	for _, exp := range rule.Expresion {
		for expType, expMessage := range exp {
			if expType == "match" {
				if f.matchExp(expMessage) {
					return true
				}
			}
		}
	}

	return false
}

//MatchFilter is a simple match rule
type IntMatchFilter struct {
	Name  string
	Field string
	Value uint64
}

func (f *IntMatchFilter) matchExp(exp json.RawMessage) bool {
	//{'left': {'payload': {'name': 'tcp', 'field': 'dport'}},
	//   'right': 8001}
	//fmt.Println(string(exp))
	var data struct {
		Left struct {
			Payload struct {
				Name  string `json:"name"`
				Field string `json:"field"`
			} `json:"payload"`
		} `json:"left"`
		Right uint64 `json:"right"`
	}

	if err := json.Unmarshal(exp, &data); err != nil {
		return false
	}

	return data.Left.Payload.Name == f.Name &&
		data.Left.Payload.Field == f.Field &&
		data.Right == f.Value
}

func (f *IntMatchFilter) Match(rule *NftRuleBlock) bool {
	for _, exp := range rule.Expresion {
		for expType, expMessage := range exp {
			if expType == "match" {
				if f.matchExp(expMessage) {
					return true
				}
			}
		}
	}

	return false
}

type NetworkMatchFilter struct {
	Name  string
	Field string
	Value *net.IPNet
}

func (f *NetworkMatchFilter) matchExp(exp json.RawMessage) bool {
	/*
		{
			"match": {
				"left": {
					"payload": {
					"name": "ip",
					"field": "saddr"
					}
				},
				"right": {
					"prefix": {
					"addr": "172.18.0.0",
					"len": 16
					}
				}
			}
		}
	*/

	var data struct {
		Left struct {
			Payload struct {
				Name  string `json:"name"`
				Field string `json:"field"`
			} `json:"payload"`
		} `json:"left"`
		Right struct {
			Prefix struct {
				Addr net.IP `json:"addr"`
				Len  int    `json:"len"`
			} `json:"prefix"`
		} `json:"right"`
	}

	if err := json.Unmarshal(exp, &data); err != nil {
		return false
	}

	mask, _ := f.Value.Mask.Size()
	return data.Left.Payload.Name == f.Name &&
		data.Left.Payload.Field == f.Field &&
		data.Right.Prefix.Addr.Equal(f.Value.IP) &&
		data.Right.Prefix.Len == mask
}

func (f *NetworkMatchFilter) Match(rule *NftRuleBlock) bool {
	for _, exp := range rule.Expresion {
		for expType, expMessage := range exp {
			if expType == "match" {
				if f.matchExp(expMessage) {
					return true
				}
			}
		}
	}

	return false
}

type IPMatchFilter struct {
	Name  string
	Field string
	Value net.IP
}

func (f *IPMatchFilter) matchExp(exp json.RawMessage) bool {
	/*
		{
			"match": {
				"left": {
					"payload": {
					"name": "ip",
					"field": "saddr"
					}
				},
				"right": {
					"prefix": {
					"addr": "172.18.0.0",
					"len": 16
					}
				}
			}
		}
	*/

	var data struct {
		Left struct {
			Payload struct {
				Name  string `json:"name"`
				Field string `json:"field"`
			} `json:"payload"`
		} `json:"left"`
		Right json.RawMessage `json:"right"`
	}

	if err := json.Unmarshal(exp, &data); err != nil {
		return false
	}

	if data.Left.Payload.Name != f.Name ||
		data.Left.Payload.Field != f.Field {
		return false
	}

	var right interface{}
	if err := json.Unmarshal(data.Right, &right); err != nil {
		return false
	}
	switch right := right.(type) {
	case string:
		return f.Value.String() == right
	}
	var prefix struct {
		Prefix struct {
			Addr net.IP `json:"addr"`
			Len  int    `json:"len"`
		} `json:"prefix"`
	}

	if err := json.Unmarshal(data.Right, &prefix); err != nil {
		return false
	}

	network := net.IPNet{
		IP:   prefix.Prefix.Addr,
		Mask: net.CIDRMask(prefix.Prefix.Len, 32), //32 for a IPv4
	}

	return network.Contains(f.Value)
}

func (f *IPMatchFilter) Match(rule *NftRuleBlock) bool {
	for _, exp := range rule.Expresion {
		for expType, expMessage := range exp {
			if expType == "match" {
				if f.matchExp(expMessage) {
					return true
				}
			}
		}
	}

	return false
}

//And allows grouping filters in an And op
type And []Filter

func (f And) Match(rule *NftRuleBlock) bool {
	if len(f) == 0 {
		return false
	}

	for _, filter := range f {
		if !filter.Match(rule) {
			return false
		}
	}

	return true
}

type Or []Filter

func (f Or) Match(rule *NftRuleBlock) bool {
	for _, filter := range f {
		if filter.Match(rule) {
			return true
		}
	}

	return false
}

type TableFilter struct {
	Table string
}

func (f *TableFilter) Match(rule *NftRuleBlock) bool {
	return rule.Table == f.Table
}

type ChainFilter struct {
	Chain string
}

func (f *ChainFilter) Match(rule *NftRuleBlock) bool {
	return rule.Chain == f.Chain
}

type FamilyFilter struct {
	Family Family
}

func (f *FamilyFilter) Match(rule *NftRuleBlock) bool {
	return rule.Family == f.Family
}
