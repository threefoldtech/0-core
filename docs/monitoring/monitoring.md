# Monitoring

All running process can report logs about its operations using the same [logging](logging.md) mechanism.

- See [Logging](logging.md) for details about how log information is processed and reported
- See [Statistics](stats.md) for details about how statistics are reported using log messages

Monitoring is one of the processes that come with Core0. It collects metrics about the system. It reports about it using the [Logging](logging.md) mechanism. As documented in [Statistics](stats.md) the default monitoring process uses **log level 10**, which is the reserved log level for logging statistics.

Here's an example of a statistics log message:

```
10::{monitoring-metric}:23.12|A
```

This example log message is reporting a statistic (log level 10) and reports that the metric `{monitoring-metric}` is `23.12` and the reported values should be averaged over the defined aggregator period (usually 5 minutes), as indicated by the `A`; see [Statistics](stats.md) for more details about this statistics log message format.

See below for:

- [Monitoring metrics](#metrics)
- [Configuring monitoring](#config)


<a id="metrics"></a>
## Monitoring metrics

The built-in/default monitoring logs statistics for following metrics:

```
disk.iops.read@phys.sda
disk.iops.read@phys.sda1
disk.iops.write@phys.sda
disk.iops.write@phys.sda1
disk.throughput.read@phys.sda
disk.throughput.read@phys.sda1

machine.CPU.contextswitch@phys
machine.CPU.interrupts@phys

machine.CPU.percent@pyhs.0
machine.CPU.percent@pyhs.1
machine.CPU.percent@pyhs.2
machine.CPU.percent@pyhs.3
machine.CPU.percent@pyhs.4
machine.CPU.percent@pyhs.5
machine.CPU.percent@pyhs.6
machine.CPU.percent@pyhs.7

machine.CPU.utilisation@pyhs.0
machine.CPU.utilisation@pyhs.1
machine.CPU.utilisation@pyhs.2
machine.CPU.utilisation@pyhs.3
machine.CPU.utilisation@pyhs.4
machine.CPU.utilisation@pyhs.5
machine.CPU.utilisation@pyhs.6
machine.CPU.utilisation@pyhs.7

machine.memory.ram.available@phys
machine.memory.swap.left@phys
machine.memory.swap.used@phys

network.packets.rx@phys.core-0
network.packets.rx@phys.eth0
network.packets.rx@phys.lo
network.packets.rx@phys.zt0
network.packets.tx@phys.core-0
network.packets.tx@phys.eth0
network.packets.tx@phys.lo
network.packets.tx@phys.zt0

network.throughput.incoming@phys.core-0
network.throughput.incoming@phys.eth0
network.throughput.incoming@phys.lo
network.throughput.incoming@phys.zt0
network.throughput.outgoing@phys.core-0
network.throughput.outgoing@phys.eth0
network.throughput.outgoing@phys.lo
network.throughput.outgoing@phys.zt0
```


<a id="config></a>
## Configuring Monitoring

The built-in/default monitoring is scheduled to run automatically as defined in the `conf/monitor.toml` configuration file:

```
[startup."monitor.cpu"]
name = "monitor"
recurring_period = 30 #seconds

[startup."monitor.cpu".args]
domain = "cpu"

[startup."monitor.memory"]
name = "monitor"
recurring_period = 30 #seconds

[startup."monitor.memory".args]
domain = "memory"

[startup."monitor.disk"]
name = "monitor"
recurring_period = 30 #seconds

[startup."monitor.disk".args]
domain = "disk"

[startup."monitor.network"]
name = "monitor"
recurring_period = 30 #seconds

[startup."monitor.network".args]
domain = "network"
```
