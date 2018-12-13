package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/threefoldtech/0-core/apps/plugins/nft"
)

//NftJsonBlock defines a nft json block
type NftJsonBlock map[string]json.RawMessage

type NftTableBlock struct {
	//{'family': 'ip', 'name': 'nat', 'handle': 0}
	Family nft.Family `json:"family"`
	Name   string     `json:"name"`
	Handle int        `json:"handle"`
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

	Family   nft.Family `json:"family"`
	Name     string     `json:"name"`
	Table    string     `json:"table"`
	Elements []string   `json:"elem"`
	Type     string     `json:"type"`
	Handle   int        `json:"handle"`
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
	Hook     string     `json:"hook"`
	Family   nft.Family `json:"family"`
	Priority int        `json:"prio"`
	Table    string     `json:"table"`
	Name     string     `json:"name"`
	Handle   int        `json:"handle"`
	Type     nft.Type   `json:"type"`
	Policy   string     `json:"policy"`
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
	Family    nft.Family     `json:"family"`
	Expresion []NftJsonBlock `json:"expr"`
	Table     string         `json:"table"`
	Handle    int            `json:"handle"`
	Chain     string         `json:"chain"`
}

func setTableBlock(set nft.Nft, msg json.RawMessage) error {
	var table NftTableBlock
	if err := json.Unmarshal(msg, &table); err != nil {
		return err
	}

	set[table.Name] = nft.Table{
		Chains: nft.Chains{},
		Sets:   nft.Sets{},
		Family: table.Family,
	}

	return nil
}

func setSetBlock(n nft.Nft, msg json.RawMessage) error {
	var set NftSetBlock
	if err := json.Unmarshal(msg, &set); err != nil {
		return err
	}
	table, ok := n[set.Table]
	if !ok {
		return fmt.Errorf("unknown table %s", set.Table)
	}

	table.Sets[set.Name] = nft.Set{
		Elements: set.Elements,
	}

	n[set.Table] = table
	return nil
}

func setChainBlock(set nft.Nft, msg json.RawMessage) error {
	var chain NftChainBlock
	if err := json.Unmarshal(msg, &chain); err != nil {
		return err
	}
	table, ok := set[chain.Table]
	if !ok {
		return fmt.Errorf("unknown table %s", chain.Table)
	}

	table.Chains[chain.Name] = nft.Chain{
		Type:     chain.Type,
		Hook:     chain.Hook,
		Priority: chain.Priority,
		Policy:   chain.Policy,
	}

	set[chain.Table] = table

	return nil
}

func renderDnat(buf *strings.Builder, msg json.RawMessage) error {
	//{'port': 7999, 'addr': '172.18.0.2'}
	var dnat struct {
		Port    int    `json:"port"`
		Address string `json:"addr"`
	}

	if err := json.Unmarshal(msg, &dnat); err != nil {
		return err
	}

	buf.WriteString(fmt.Sprintf("dnat to %s:%d", dnat.Address, dnat.Port))

	return nil
}

func renderLeft(buf *strings.Builder, msg json.RawMessage) error {
	var left struct {
		Meta string `json:"meta"`
		CT   struct {
			Key string `json:"key"`
		} `json:"ct"`
		Payload struct {
			Name  string `json:"name"`
			Field string `json:"field"`
		} `json:"payload"`
	}

	if err := json.Unmarshal(msg, &left); err != nil {
		return err
	}

	if len(left.Meta) > 0 {
		buf.WriteString(left.Meta)
	} else if len(left.CT.Key) > 0 {
		buf.WriteString(fmt.Sprintf("ct %s", left.CT.Key))
	} else if len(left.Payload.Name) > 0 {
		buf.WriteString(fmt.Sprintf("%s %s", left.Payload.Name, left.Payload.Field))
	}

	return nil
}

func renderRight(buf *strings.Builder, msg json.RawMessage) error {
	//right block can be simple as a string or int, or complex as a prefix or set
	var right interface{}
	if err := json.Unmarshal(msg, &right); err != nil {
		return err
	}
	switch r := right.(type) {
	case json.Number:
		buf.WriteString(r.String())
		return nil
	case int64:
		buf.WriteString(fmt.Sprint(r))
		return nil
	case float64:
		buf.WriteString(fmt.Sprint(r))
		return nil
	case string:
		if r[0] == '@' || (r[0] >= '0' && r[0] <= '9') { //check if starts with number probably ip or cird
			//for sets
			buf.WriteString(r)
			return nil
		}
		//for interface names and such
		buf.WriteByte('"')
		buf.WriteString(r)
		buf.WriteByte('"')
		return nil
	}

	//if we reached here, then right is not a primitive
	//we need to load to a struct
	var rightStruct struct {
		Set    []interface{} `json:"set"`
		Prefix struct {
			Addr string `json:"addr"`
			Len  int    `json:"len"`
		} `json:"prefix"`
	}
	if err := json.Unmarshal(msg, &rightStruct); err != nil {
		return err
	}
	if len(rightStruct.Set) > 0 {
		buf.WriteString("{ ")
		for i, o := range rightStruct.Set {
			buf.WriteString(fmt.Sprint(o))
			if i != len(rightStruct.Set)-1 {
				buf.WriteString(", ")
			}
		}
		buf.WriteString(" }")
	}

	if len(rightStruct.Prefix.Addr) > 0 {
		buf.WriteString(fmt.Sprintf("%s/%d", rightStruct.Prefix.Addr, rightStruct.Prefix.Len))
	}

	return nil
}

func renderMatch(buf *strings.Builder, msg json.RawMessage) error {
	var match struct {
		Left  json.RawMessage `json:"left"`
		Right json.RawMessage `json:"right"`
	}
	if err := json.Unmarshal(msg, &match); err != nil {
		return err
	}
	if err := renderLeft(buf, match.Left); err != nil {
		return err
	}

	buf.WriteString(" ")

	if err := renderRight(buf, match.Right); err != nil {
		return err
	}
	return nil
}

func renderRule(expr []NftJsonBlock) (string, error) {
	var buf strings.Builder

	for _, exp := range expr {
		if len(exp) != 1 {
			return "", fmt.Errorf("invalid expression")
		}
		for expType, message := range exp {
			if buf.Len() != 0 {
				buf.WriteString(" ")
			}

			switch expType {
			case "match":
				if err := renderMatch(&buf, message); err != nil {
					return "", err
				}
			case "dnat":
				if err := renderDnat(&buf, message); err != nil {
					return "", err
				}
			case "masquerade":
				fallthrough
			case "accept":
				fallthrough
			case "drop":
				buf.WriteString(expType)
			default:
				return "", fmt.Errorf("unknown expr type '%s'", expType)
			}
		}
	}

	return buf.String(), nil
}

func setRuleBlock(set nft.Nft, msg json.RawMessage) error {
	var rule NftRuleBlock
	if err := json.Unmarshal(msg, &rule); err != nil {
		return err
	}
	table, ok := set[rule.Table]
	if !ok {
		return fmt.Errorf("unknown table %s", rule.Table)
	}

	chain, ok := table.Chains[rule.Chain]
	if !ok {
		return fmt.Errorf("unknown chain %s", rule.Chain)
	}

	body, err := renderRule(rule.Expresion)
	if err != nil {
		return err
	}
	chain.Rules = append(chain.Rules, nft.Rule{
		Handle: rule.Handle,
		Body:   body,
	})

	table.Chains[rule.Chain] = chain
	set[rule.Table] = table
	return nil
}

//Parse nft json output
func Parse(config string) (nft.Nft, error) {
	nft := nft.Nft{}

	var loaded struct {
		Blocks []NftJsonBlock `json:"nftables"`
	}

	if err := json.Unmarshal([]byte(config), &loaded); err != nil {
		return nft, err
	}

	for _, block := range loaded.Blocks {
		if len(block) != 1 {
			//this should never happen
			return nft, fmt.Errorf("invalid nft block")
		}
		var err error
		for blockType, message := range block {
			switch blockType {
			case "table":
				err = setTableBlock(nft, message)
			case "set":
				err = setSetBlock(nft, message)
			case "chain":
				err = setChainBlock(nft, message)
			case "rule":
				err = setRuleBlock(nft, message)
			}
		}
		if err != nil {
			return nft, err
		}
	}

	return nft, nil
}
