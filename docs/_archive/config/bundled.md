## Quick description to bundled config

### disk.toml
Add extensions to manage disks using `parted` to create, delete partitions. Also mount and umount of partitions

### libvirt.toml
Startup needed services for libvirtd operations

### modprobe.toml
Load needed generic modules for ethernet devices

### monitor.toml
Start recurring monitoring actions to monitor cpu, memory, disks, etc...

### redis.toml
Start required redis services for core0 communication. 

### zerotier.toml
Add extension to manage zerotier networks (join, list and leave). Also add startup section
to join a zerotier network passed from the kernel
