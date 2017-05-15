# Bridge Commands

Available commands:

- [bridge.create](#create)
- [bridge.list](#list)
- [bridge.delete](#delete)


<a id="create"></a>
## bridge.create

Create a bridge with the given name, MAC address and networking setup.


Arguments:
```javascript
{
  "name": {name},
  "hwaddr": {mac-address},
  "network": {
      "mode": {network-mode},
      "nat": {nat},
      "settings": {settings},
   }
}
```

Values:

- **name**: Name of the bridge (must be unique)
- **mac-address**: MAC address of the bridge, if none, one will be created for you
- **network-mode**: Networking mode, options are `none`, `static`, and `dnsmasq`
- **nat**: If true, SNAT will be enabled on this bridge
- **settings**: Networking settings, depending on the selected mode:
  - none: no settings, bridge won't get any IP settings
  - static: `settings={'cidr': 'ip/net'}`, bridge will get assigned the given IP address
  - dnsmasq: `settings={'cidr': 'ip/net', 'start': 'ip', 'end': 'ip'}`, bridge will get assigned the IP address in CIDR and each running container that is attached to this IP address will get IP address from the start/end range, Netmask of the range is the netmask part of the provided CIDR, if nat is true, SNAT rules will be automatically added in the firewall


<a id="list"></a>
## bridge.list

List all available bridges. Takes no arguments.


<a id="delete"></a>
## bridge.delete

Delete the given bridge name.

Arguments:

```javascript
{
    "name": "bridge-name",
}
```
