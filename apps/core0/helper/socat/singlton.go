package socat

//SetPortForward create a single port forward from host(port), to ip(addr) and dest(port) in this namespace
//The namespace is used to group port forward rules so they all can get terminated
//with one call later.
func SetPortForward(namespace string, ip string, host string, dest int) error {
	return socat.SetPortForward(namespace, ip, host, dest)
}

//RemovePortForward removes a single port forward
func RemovePortForward(namespace string, host string, dest int) error {
	return socat.RemovePortForward(namespace, host, dest)
}

//RemoveAll remove all port forwrards that were created in this namespace.
func RemoveAll(namespace string) error {
	return socat.RemoveAll(namespace)
}

//Resolve resolves an address of the form <ip>:<port> to a direct address to the endpoint
//IF
// - the ip address is a local address of this machine
// - port has a forwarding rule
//ELSE
// - return address unchanged
func Resolve(address string) string {
	return socat.Resolve(address)
}

//ResolveURL rewrites a url to a direct address to the end point. Return original url
//if no forwarding rule configured that matches the given address
//note, the url host part must be an ip, can't use host names
func ResolveURL(raw string) (string, error) {
	return socat.ResolveURL(raw)
}
