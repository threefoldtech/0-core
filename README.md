
[![Build Status](https://api.travis-ci.org/zero-os/0-core.svg?branch=development)](https://travis-ci.org/zero-os/0-core/)
[![codecov](https://codecov.io/gh/zero-os/0-core/branch/master/graph/badge.svg)](https://codecov.io/gh/zero-os/0-core)

# 0-core

The core of Zero-OS is 0-core, which is the Zero-OS replacement for systemd.

## Branches

- [master](https://github.com/zero-os/0-core/tree/master) - production

## Releases

See the release schedule in the [Zero-OS home repository](https://github.com/zero-os/home).

## Development setup

Check the page on how to boot zos in a local setup [here](docs/booting/README.md). Choose the best one that suits your
setup. For development, we would recommend the [VM using QEMU](docs/booting/qemu.md).

## Enteracting with zos
ZOS does not provide interactive shell, or a UI all interactions id done through any of its interfaces. For more details about interaction with zos please check [the docs here](docs/interacting/README.md)

## Features
### v1.2
- Lots and lots of bug fixes
- Many new primitives
- Auto cache disk creation and mounting
- Crash report
- Different boot modes
- Starting services based on boot mode (service conditions)
- Support passthrough, and macvlan for container nic type
- portforward using dnat instead of socat

### v1.1.0-alpha2
- Lots and lots of bug fixes
- Containers plugins
- Unprivileged containers (still in beta)
- Libvirt bindings
- Processes API
- Support multiple ZeroTier in container networking
- Support Open vSwitch networking for both containers and KVM domains
- `corectl` command line tool to manage Zero-OS from within the node

### v1.0.0
- New Flist format, the flist used in the 0-fs is now a distributed as a RocksDB database.
- Creation of the [0-Hub](https://github.com/zero-os/core0/tree/1.0.0), see https://github.com/zero-os/0-hub
- Improvement of the builtin commands of 0-core and coreX


### v0.11
- include of the monitoring of all processes running on the Zero-OS.
  It produces aggregated statistics on the processes that can be dump into a time series database and displayed used something like Grafana.


### v0.10
- change datastore for fuse filesystem from ipfs to [Zero-OS Store](https://github.com/g8os/stor).

### v0.9
- Boot the 0-core as init process
- Manage disks
- Create containers
  - Full Namespace isolation
  - Host the root filesystem of the containers via ipfs
  - Network stack dedicated
  - ZeroTier Network integration
  - Use flist file format as root metadata
- Remotly administrate the process
  - via Python client
  - via redis

### Next

See the milestones in the [Zero-OS home repository](https://github.com/zero-os/home): [Zero-OS Milestones](https://github.com/zero-os/home/tree/master/milestones)

## Schema
![Schema Plan](specs/schema.png)

## Documentation

All documentation is in the [`/docs`](./docs) directory, including a [table of contents](/docs/SUMMARY.md).

In [Getting Started with Core0](/docs/gettingstarted/README.md) you find the recommended path to quickly get up and running.
