package client

type PowerManager interface {
	Reboot() error
	PowerOff() error
}

type powerManager struct {
	cl Client
}

func Power(cl Client) PowerManager {
	return &powerManager{cl}
}

func (m *powerManager) Reboot() error {
	_, err := m.cl.Raw("power.reboot", A{})
	return err
}

func (m *powerManager) PowerOff() error {
	_, err := m.cl.Raw("power.poweroff", A{})
	return err
}
