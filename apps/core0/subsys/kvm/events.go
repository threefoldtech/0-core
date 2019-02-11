// +build amd64

package kvm

import (
	"encoding/json"
	"runtime/debug"
	"strings"

	"github.com/google/shlex"
	"github.com/libvirt/libvirt-go"
	"github.com/threefoldtech/0-core/apps/core0/helper/socat"
	"github.com/threefoldtech/0-core/base/pm"
	"github.com/threefoldtech/0-core/base/pm/stream"
)

func (m *kvmManager) deviceRemovedFailedHandler(c *libvirt.Connect, d *libvirt.Domain, event *libvirt.DomainEventDeviceRemovalFailed) {
	uuid, _ := d.GetUUIDString()
	log.Errorf("device deleted failed event for domain '%s' %s", uuid, event)

	m.devDeleteEvent.Release(uuid, event.DevAlias)
}

func (m *kvmManager) deviceRemovedHandler(c *libvirt.Connect, d *libvirt.Domain, event *libvirt.DomainEventDeviceRemoved) {
	uuid, _ := d.GetUUIDString()
	log.Debugf("device deleted event for domain '%s' %s", uuid, event)

	m.devDeleteEvent.Release(uuid, event.DevAlias)
}

func (m *kvmManager) handleStopped(uuid, name string, domain *libvirt.Domain) error {
	/*
		It's too late to get the xml definition, so we don't know if this machine is booted from
		an flist or not. One approach is to keep in memory description of the machine that needs
		clean up. Or simply try to unmount the expected target by default, and hide unmount errors
	*/
	defer m.updateView()
	m.domainsInfoRWMutex.Lock()
	info, ok := m.domainsInfo[uuid]
	delete(m.domainsInfo, uuid)
	m.domainsInfoRWMutex.Unlock()
	if ok {
		socat.RemoveAll(m.forwardId(info.Sequence))
	}

	return m.flistUnmount(uuid)
}

func (m *kvmManager) domaineLifeCycleHandler(conn *libvirt.Connect, domain *libvirt.Domain, event *libvirt.DomainEventLifecycle) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("error processing domain event: %v", err)
			log.Error(string(debug.Stack()))
		}
	}()

	uuid, _ := domain.GetUUIDString()
	name, _ := domain.GetName()
	parts, _ := shlex.Split(event.String())
	data := map[string]interface{}{
		"uuid": uuid,
		"name": name,
	}

	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			data[kv[0]] = kv[1]
		}
	}
	if event, ok := data["event"]; ok {
		var err error
		switch event {
		case "stopped":
			err = m.handleStopped(uuid, name, domain)
		}
		if err != nil {
			log.Errorf("failed to handle event (%s) for vm (%s): %s", event, uuid, err)
		}
	}
	m.evch <- data
}

func (m *kvmManager) events(ctx *pm.Context) (interface{}, error) {
	var sequence uint64 = 1
	for data := range m.evch {
		data["sequence"] = sequence
		bytes, _ := json.Marshal(data)
		ctx.Log(string(bytes), stream.LevelResultJSON)

		sequence++
	}

	return nil, nil
}
