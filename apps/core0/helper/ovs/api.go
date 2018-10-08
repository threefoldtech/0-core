package ovs

import (
	"encoding/json"
	"fmt"
)

/*
"bridge-add":   ovs.BridgeAdd,
		"bridge-del":   ovs.BridgeDelete,
		"port-add":     ovs.PortAdd,
		"port-del":     ovs.PortDel,
		"bond-add":     ovs.BondAdd,
		"vtep-ensure":  ovs.VtepEnsure,
		"vtep-del":     ovs.VtepDelete,
		"vlan-ensure":  ovs.VLanEnsure,
		"vxlan-ensure": ovs.VXLanEnsure,
		"set":          ovs.Set,
*/
var (
	api ovsAPI
)

type ovsAPI struct{}

func ini() {

}

type Bridge struct {
	Bridge  string            `json:"bridge"`
	Options map[string]string `json:"options"`
}

func (b *Bridge) Validate() error {
	if b.Bridge == "" {
		return fmt.Errorf("bridge name is not set")
	}
	return nil
}

type PortAddArguments struct {
	Bridge
	Port    string            `json:"port"`
	VLan    uint16            `json:"vlan"`
	Options map[string]string `json:"options"`
}

func (p *PortAddArguments) Validate() error {
	if err := p.Bridge.Validate(); err != nil {
		return err
	}
	if p.Port == "" {
		return fmt.Errorf("missing port name")
	}
	return nil
}

type PortDelArguments struct {
	Bridge
	Port string `json:"port"`
}

func (p *PortDelArguments) Validate() error {
	if p.Port == "" {
		return fmt.Errorf("missing port name")
	}
	return nil
}

type SetArguments struct {
	Table  string            `json:"table"`
	Record string            `json:"record"`
	Values map[string]string `json:"values"`
}

func (s *SetArguments) Validate() error {
	if s.Table == "" {
		return fmt.Errorf("missing table name")
	}
	if s.Record == "" {
		return fmt.Errorf("missing record")
	}
	if len(s.Values) == 0 {
		return fmt.Errorf("no values to set")
	}

	return nil
}

type VLanEnsureArguments struct {
	Master string `json:"master"`
	VLan   uint16 `json:"vlan"`
	Name   string `json:"name"`
}

func (v *VLanEnsureArguments) Validate() error {
	if v.Master == "" {
		return fmt.Errorf("master bridge not specified")
	}
	if v.VLan < 0 || v.VLan >= 4095 { //0 for untagged
		return fmt.Errorf("invalid vlan tag")
	}
	return nil
}

type VTepEnsureArguments struct {
	Bridge
	VNID uint `json:"vnid"`
}

func (t *VTepEnsureArguments) Validate() error {
	if err := t.Bridge.Validate(); err != nil {
		return err
	}
	if t.VNID == 0 {
		return fmt.Errorf("invalid nid")
	}
	return nil
}

type VTepDeleteArguments struct {
	VNID uint `json:"vnid"`
}

func (t *VTepDeleteArguments) Validate() error {
	if t.VNID == 0 {
		return fmt.Errorf("invalid nid")
	}
	return nil
}

type VXLanEnsureArguments struct {
	Master string `json:"master"`
	VXLan  uint   `json:"vxlan"`
	Name   string `json:"name"`
}

func (v *VXLanEnsureArguments) Validate() error {
	if v.Master == "" {
		return fmt.Errorf("master bridge not specified")
	}

	return nil
}

func (a *ovsAPI) BridgeAdd(args json.RawMessage) (interface{}, error) {
	var bridge Bridge
	if err := json.Unmarshal(args, &bridge); err != nil {
		return nil, err
	}

	if err := bridge.Validate(); err != nil {
		return nil, err
	}

	return nil, BridgeAdd(bridge.Bridge, MakeOptions(bridge.Options)...)
}

func (a *ovsAPI) BridgeDelete(args json.RawMessage) (interface{}, error) {
	var bridge Bridge
	if err := json.Unmarshal(args, &bridge); err != nil {
		return nil, err
	}

	if err := bridge.Validate(); err != nil {
		return nil, err
	}

	return vsctl("del-br", bridge.Bridge)
}

func (a *ovsAPI) PortAdd(args json.RawMessage) (interface{}, error) {
	var port PortAddArguments
	if err := json.Unmarshal(args, &port); err != nil {
		return nil, err
	}

	if err := port.Validate(); err != nil {
		return nil, err
	}

	return nil, PortAdd(
		port.Port,
		port.Bridge.Bridge,
		port.VLan,
		MakeOptions(port.Options)...,
	)
}

func (a *ovsAPI) PortDel(args json.RawMessage) (interface{}, error) {
	var port PortDelArguments
	if err := json.Unmarshal(args, &port); err != nil {
		return nil, err
	}

	if err := port.Validate(); err != nil {
		return nil, err
	}

	return nil, PortDel(port.Port, port.Bridge.Bridge)
}

func (a *ovsAPI) Set(args json.RawMessage) (interface{}, error) {
	var s SetArguments
	if err := json.Unmarshal(args, &s); err != nil {
		return nil, err
	}
	if err := s.Validate(); err != nil {
		return nil, err
	}

	return nil, set(s.Table, s.Record, MakeOptions(s.Values)...)
}

func (a *ovsAPI) BondAdd(args json.RawMessage) (interface{}, error) {
	var bond BondAddArguments
	if err := json.Unmarshal(args, &bond); err != nil {
		return nil, err
	}

	if err := bond.Validate(); err != nil {
		return nil, err
	}
	mode := bond.Mode
	if mode == BondMode("") {
		mode = BondModeBalanceSLB
	}

	return nil, BondAdd(bond.Port, bond.Bridge.Bridge, mode, bond.LACP, bond.Links...)
}

func (a *ovsAPI) VLanEnsure(args json.RawMessage) (interface{}, error) {
	//abstract method to ensure a bridge exists that has this vlan tag.
	var vlan VLanEnsureArguments
	if err := json.Unmarshal(args, &vlan); err != nil {
		return nil, err
	}

	if err := vlan.Validate(); err != nil {
		return nil, err
	}

	return VLanEnsure(vlan.Name, vlan.Master, vlan.VLan)
}

func (a *ovsAPI) VtepEnsure(args json.RawMessage) (interface{}, error) {
	var vtep VTepEnsureArguments
	if err := json.Unmarshal(args, &vtep); err != nil {
		return nil, err
	}

	if err := vtep.Validate(); err != nil {
		return nil, err
	}

	return VtepEnsure(vtep.VNID, vtep.Bridge.Bridge)
}

func (a *ovsAPI) VtepDelete(args json.RawMessage) (interface{}, error) {
	var vtep VTepDeleteArguments
	if err := json.Unmarshal(args, &vtep); err != nil {
		return nil, err
	}

	if err := vtep.Validate(); err != nil {
		return nil, err
	}

	return nil, VtepDelete(vtep.VNID)
}

func (a *ovsAPI) VXLanEnsure(args json.RawMessage) (interface{}, error) {
	//abstract method to ensure a bridge exists that has this vlan tag.
	var vxlan VXLanEnsureArguments
	if err := json.Unmarshal(args, &vxlan); err != nil {
		return nil, err
	}

	if err := vxlan.Validate(); err != nil {
		return nil, err
	}

	return VXLanEnsure(vxlan.Name, vxlan.Master, vxlan.VXLan)
}
