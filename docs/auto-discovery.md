# Auto discovery

In a G8OS grid architecture, the G8OS node need to register themself against the AYS server that managed the node. In order to do that, core0 need to send an event to the AYS server.

Here is the script used to enable that behavior:
```toml
[startup."ays.register"]
name = "bash"
after = ["net"]
max_restart = 10000 # it automatically sleeps 1 sec between each trial

[startup."ays.register".args]
script = """
#!/bin/ash
echo -en \"POST /webhooks/events HTTP/1.1\r\n\" > /tmp/discover
echo -en \"Host: {ays_ip}\r\n\" >> /tmp/discover
echo -en \"Content-Type: application/json\r\n\" >> /tmp/discover
echo -en \"Connection: close\r\n\" >> /tmp/discover
echo -en \"Content-Length: 24\r\n\r\n\" >> /tmp/discover
echo -en '{\"command\": \"discover\"}' \"\r\n\" >> /tmp/discover

nc {ays_ip} {ays_port} < /tmp/discover
"""
```

Notice the `{ays_ip}` and `{ays_port}`. These two string will be replace by the kernel parameters passed during boot.

Example with qemu:
```shell
qemu-system-x86_64 -kernel /tmp/vmlinuz.efi -m 2048 -enable-kvm -cpu host -net nic,model=e1000 -net bridge,br=lxc0  -nodefaults  -nographic -serial mon:stdio -append 'console=ttyS0 ays_ip=192.168.5.10 ays_port=5000'
```
