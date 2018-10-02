# Port forwards
Port forwards are supported for both containers and virtual machines. Both the `create` calls for containers and kvms takes an optional `port` argument

> The port argument is only applied when (and only when) the contianer/kvm is configured with default network.

Both apis as well provide methods to add/remove port forward rules dynamically on the fly

## Specifying a port forward rule
the `port` argument is a dict where the key is the source match (host port) and dest is just the container/kvm internal port (number). The syntax for the source port is defined as follows:

```
source = port
source = address:port
source = source|protocol
protocol = tcp
protocol = udp
protocol = protocol(+protocol)?
address = ip
address = ip/mask
address = nic
address = nic*
```

### Examples

- `"8080": 80`
- `"10.20.0.1:4330": 433`
- `"192.168.1.0/24:2022": 22`
- `"530|udp": 53"`
- `"eth0:1000|tcp+udp": 7000`
- `"zt*:7999|tcp": 7999`