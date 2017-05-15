# Network Configuration

Core0 applies the network configuration as defined in the `/etc/g8os/network.toml`, structured as follows:

```toml
[network]
auto = true

[interface.<ifname>]
protocol = "<protocol>"

[<protocol>.<ifname>]
#protocol specific configuration
#for that interface
```

The configuration is applied as follows:

- It tries to reach at least one of the configured `controllers`
- If it can't reach any, it tries `dhcp` on all interfaces, one at a time and tries to reach any of the controllers again
- If it still can't reach a controller, it will fall back to a random static IP address in the `10.254.254.0/24` range and tries to reach a controller on `10.254.254.254`

The `[network]` section currently only has the `auto` flag, which tells core to auto configure all found interfaces using the DHCP method if there is no specific `interface` section for that interface. If `auto` is set to false, the core will just ignore any interface that is not specifically configured in the `network.toml` file.


## Supported network protocols

Currently only the following network protocols are supported:

- **none**: Only brings the interface up, it doesn't set an IP address on that interface, it also excludes it from the networking fallback plan. This is useful if the interface is added to an Open vSwitch switch
- **dhcp**: Doesn't require any further configurations, specifying `protocl = "dhcp"` is enough
- **static**: Requires a `[static.<if>]` section, see the below example

Below two examples:

- [Static](#static)
- [Open vSwitch](#ovs)


<a id="static"></a>
### Static Example

```toml
[interface.eth0]
protocol = "static"

[static.eth0]
ip = "10.20.30.40/24"
gateway = "10.20.30.1"
```


<a id="ovs"></a>
### Open vSwitch example

Open vSwitch services are started before processing the `network.toml` file. So by the time of processing the `network.tom` file all bridges should be in place. After that, all tap devices in the `network.toml` file are going to be created.

```bash
ovs-vsctrl add-br br0
ovs-vsctrl add-port br0 eth0
ovs-vsctrl add-port br0 port1
```

```toml
[interface.lo]
protocol = "static"

[static.lo]
ip = "127.0.0.1/8"

[interface.br0]
protocol = "dhcp"

[interface.eth0]
protocol = "none"
```

The final result of this setup should be something like the following:

```
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast master ovs-system state UP group default qlen 1000
    link/ether 08:00:27:b8:5a:7b brd ff:ff:ff:ff:ff:ff
3: ovs-system: <BROADCAST,MULTICAST> mtu 1500 qdisc noop state DOWN group default
    link/ether ea:ac:67:3f:18:42 brd ff:ff:ff:ff:ff:ff
4: br0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UNKNOWN group default
    link/ether 08:00:27:b8:5a:7b brd ff:ff:ff:ff:ff:ff
    inet 10.254.254.79/24 scope global br0
       valid_lft forever preferred_lft forever
5: port1: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 qdisc noqueue master ovs-system state DOWN group default qlen 500
    link/ether c2:36:82:75:04:51 brd ff:ff:ff:ff:ff:ff
```

```
Bridge "br0"
    Port "port1"
        Interface "port1"
    Port "br0"
        Interface "br0"
            type: internal
    Port "eth0"
        Interface "eth0"
```
