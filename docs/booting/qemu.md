# Booting Zero-OS on a VM using QEMU

Follow below steps in order to boot a virtual machine with Zero-OS on a physical machine running Ubuntu using QEMU:

- [Install QEMU on Ubuntu](#install-qemu-on-ubuntu)
- [Configure the bridge](#configure-the-bridge)
- [Get a Zero-OS kernel](#get-a-zero-os-kernel)
- [Create the boot disk](#create-the-boot-disk)
- [Start the virtual machine](#start-the-virtual-machine)
- [Ping Zero-OS](#ping-zero-os)


## Install QEMU On Ubuntu

Install all required packages:
```shell
sudo apt-get update
sudo apt-get install qemu-kvm qemu virt-viewer libvirt-bin
```

After installing the above packages, reboot your system.

Check if your computer supports hardware virtualization (VT) and whether it is enabled:
```shell
kvm-ok
```

If hardware virtualization is enabled, you should see something like:
```shell
INFO: /dev/kvm exists
KVM acceleration can be used
```

If hardware virtualization is not enabled you have to go first to the BIOS/UEFI settings to enable it.


## Configure the bridge

We need to have a bridge to connect your virtual machine to your Ethernet network, typically you can use the default `virbr0`.

Check the state of existing networks:
```shell
sudo virsh net-list --all
```

If the state of the default network is **inactive**, activate it:
```shell
sudo virsh net-start default
```

> **NOTE**: If your WiFi card doesn't support bridging network connections, you have to connect the Ethernet cable.

By default the `qemu` directory and `bridge.conf` file are not created. In order to create them execute:
```shell
sudo mkdir /etc/qemu
sudo vi /etc/qemu/bridge.conf
```

Type the following on this file:
```shell
allow virbr0
```


## Get a Zero-OS kernel

Either build the kernel yourself as documented in [Building your Zero-OS Kernel](../building/README.md) or download it from the [Zero-OS Bootstrap Service](https://bootstrap.gig.tech/) as documented in [Zero-OS Bootstrap Service](../bootstrap/README.md).

We only require the kernel (`zero-os-master.efi`) file when booting with QEMU.


## Create the boot disk

Create an empty boot disk for the virtual machine:
```shell
qemu-img create -f qcow2 vda.qcow2 10G
```

> Note: Run this at any time when you want to wipe your boot disk.


## Start the virtual machine

```shell
ZEROTIER="<your_ZeroTier_network>"
sudo qemu-system-x86_64 -kernel staging/vmlinuz.efi -m 2048 -enable-kvm -cpu host -nographic -append "console=ttyS0,115200n8 zerotier=$ZEROTIER" -netdev bridge,id=virbr0,br=virbr0 -device virtio-net-pci,netdev=virbr0
```


## Ping Zero-OS

Using the Python client:

```python
from zeroos.core0.client import Client
cl = Client("<your_ZeroTier_network>")
cl.ping()
```
