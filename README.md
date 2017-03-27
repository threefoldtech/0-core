
[![Build Status](https://travis-ci.org/g8os/core0.svg?branch=master)](https://travis-ci.org/g8os/core0)
[![codecov](https://codecov.io/gh/g8os/core0/branch/master/graph/badge.svg)](https://codecov.io/gh/g8os/core0)

# Core

Systemd replacement for G8OS

## Releases:
- [0.9.0](https://github.com/g8os/core0/tree/0.9.0)
- [0.10.0](https://github.com/g8os/core0/tree/0.10.0)
- [0.11.0](https://github.com/g8os/core0/tree/0.11.0) - last release

## Sample setup
The following steps will create a docker container that have core0 as the init process. When running,
u can send commands to core0 using the pyclient

First we need to prepare the base docker image to host core0
Copy the following content to a `DockerFile` some where on your system

```dockerfile
FROM ubuntu:16.04
RUN apt-get update && \
    apt-get install -y wget && \
    apt-get install -y fuse && \
    apt-get install -y iproute2 && \
    apt-get install -y nftables && \
    apt-get install -y dnsmasq && \
    apt-get install -y libvirt-bin && \
    apt-get install -y redis-server
```

```bash
mkdir ~/build
cd ~/build
# write the content above to a Dockerfile in the build director
docker build -t core .
```

When the build is complete you should have the `core` image ready to be used in the following steps.

### Starting a container with core0
Make sure this repo is cloned under your correct GOPATH (should be under $GOPATH/src/github.com/g8os/core0). Then move to that location the do a `make`

```
cd $GOPATH/src/github.com/g8os/core0
go get github.com/g8os/core0/core0
go get github.com/g8os/core0/coreX
make
```

The do
```
docker run --privileged -d \
    --name core-jo \
    -v `pwd`/bin/core0:/usr/sbin/core0 \
    -v `pwd`/bin/coreX:/usr/sbin/coreX \
    -v `pwd`/core0/g8os.dev.toml:/root/core.toml \
    -v `pwd`/core0/conf:/root/conf \
    -p 6379:6379 \
    core \
    core0 -c /root/core.toml
```

> Note: You might ask why we do this instead of copying those files directly to the image
> the point is, now it's very easy for development, each time u rebuild the binary or change the config
> u can just do `docker restart core0` without rebuilding the whole image.


> NOTE: if u are intending to use `containers` feature of core0, make sure u either copy the `g8ufs` binary to the container, or bind it (with `-v`) like core0 and coreX binaries


To follow the container logs do
```bash
docker logs -f core0
```

## Using the client
Before using the client make sure the `./pyclient` is in your *PYTHONPATH*

```python
import client

cl = client.Client(host='ip of docker container running core0')

#validate that core0 is reachable
print(cl.ping())

#then u can do stuff like
print(
    cl.system('ps -eF').get()
)

print(
    cl.system('ip a').get()
)

#client exposes more tools for disk, bridges, and container mgmt
print(
    cl.disk.list()
)
```

# Features
v0.9:
- Boot the core0 as init process
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

v0.10:
- change datastore for fuse filesystem from ipfs to [G8OS Store](https://github.com/g8os/stor).

v0.11:
- include of the monitoring of all processes running on the g8os.
  It produces aggregated statistics on the processes that can be dump into a time series database and displayed used something like grafana.


# Available Commands
[Commands Documentation](docs/commands.md)

# Schema
![Schema Plan](specs/schema.png)

# Examples
You can find example usage [here](docs/examples/index.md)
