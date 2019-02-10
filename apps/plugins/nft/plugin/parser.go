package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/threefoldtech/0-core/apps/plugins/nft"
)

func renderMangle(buf *strings.Builder, msg json.RawMessage) error {
	var mangle struct {
		Left struct {
			Meta string `json:"meta"`
		} `json:"left"`
		Right uint32 `json:"right"`
	}

	if err := json.Unmarshal(msg, &mangle); err != nil {
		return err
	}

	buf.WriteString(fmt.Sprintf("%s set 0x%08x", mangle.Left.Meta, mangle.Right))
	return nil
}

func renderDnat(buf *strings.Builder, msg json.RawMessage) error {
	//{'port': 7999, 'addr': '172.18.0.2'}
	var dnat struct {
		Port uint16 `json:"port"`
		Addr string `json:"addr"`
	}

	if err := json.Unmarshal(msg, &dnat); err != nil {
		return err
	}

	buf.WriteString(fmt.Sprintf("dnat to %s:%d", dnat.Addr, dnat.Port))

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

func renderRule(expr []nft.NftJsonBlock) (string, error) {
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
			case "mangle":
				if err := renderMangle(&buf, message); err != nil {
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

//Filter nft json output
func Filter(config string, filters ...nft.Filter) ([]nft.FilterRule, error) {
	var loaded struct {
		Blocks []nft.NftJsonBlock `json:"nftables"`
	}

	if err := json.Unmarshal([]byte(config), &loaded); err != nil {
		return nil, err
	}

	var rules []nft.FilterRule
	for _, block := range loaded.Blocks {
		for blockType, message := range block {
			if blockType != "rule" {
				//filter only works on rules
				continue
			}

			var rule nft.NftRuleBlock
			if err := json.Unmarshal(message, &rule); err != nil {
				return nil, err
			}

			for _, filter := range filters {
				if !filter.Match(&rule) {
					continue
				}

				body, err := renderRule(rule.Expresion)
				if err != nil {
					return nil, err
				}
				rules = append(rules,
					nft.FilterRule{
						Handle: rule.Handle,
						Rule:   nft.Rule{Body: body},
						Table:  rule.Table,
						Chain:  rule.Chain,
						Family: rule.Family,
					},
				)

				break
			}
		}
	}

	return rules, nil
}
