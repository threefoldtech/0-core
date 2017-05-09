# Python Client

**g8core** is the Python client used to talk to [G8OS core](https://github.com/g8os/core0)

## Install

```bash
pip install g8core
```

## How to use
```python
import g8core

cl = g8core.Client(host='ip of docker container running core0')

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

## JumpScale Integration

**g8core** is integrated in JumpScale under `j.clients.g8os`, see [JumpScale Client](jumpscale.md) for example code.
