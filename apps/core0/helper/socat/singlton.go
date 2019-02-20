package socat

//SetPortForward create a single port forward from host(port), to ip(addr) and dest(port) in this namespace
//The namespace is used to group port forward rules so they all can get terminated
//with one call later.
func SetPortForward(ns NS, ip string, host string, dest int) error {
	return mgr.SetPortForward(ns, ip, host, dest)
}

//RemovePortForward removes a single port forward
func RemovePortForward(ns NS, host string, dest int) error {
	return mgr.RemovePortForward(ns, host, dest)
}

//RemoveAll remove all port forwrards that were created in this namespace.
func RemoveAll(ns NS) error {
	return mgr.RemoveAll(ns)
}

//Resolve resolves an address of the form <ip>:<port> to a direct address to the endpoint
//IF
// - the ip address is a local address of this machine
// - port has a forwarding rule
//ELSE
// - return address unchanged
func Resolve(address string) string {
	return mgr.Resolve(address)
}

//ResolveURL rewrites a url to a direct address to the end point. Return original url
//if no forwarding rule configured that matches the given address
//note, the url host part must be an ip, can't use host names
func ResolveURL(raw string) (string, error) {
	return mgr.ResolveURL(raw)
}

func List(ns NS) (PortMap, error) {
	return mgr.List(ns)
}

func ListAll(system uint8) (map[NS]PortMap, error) {
	return mgr.ListAll(system)
}
