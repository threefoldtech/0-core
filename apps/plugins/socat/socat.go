package socat

type API interface {
	SetPortForward(namespace string, ip string, host string, dest int) error
	RemovePortForward(namespace string, host string, dest int) error
	RemoveAll(namespace string) error
	Resolve(address string) string
	ResolveURL(raw string) (string, error)
}
