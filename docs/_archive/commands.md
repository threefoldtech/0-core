> This document is a continues WIP

#Core0/CoreX communication protocol

## Core0
First process to start on bare metal. It works as a simple process manager.
When starts it first configure the networking then. It also starts a local redis instance to dispatch commands to CoreX Cores.

## Command structure

```javascript
{
	"id": "command-id",
	"command": "command-name",
	"arguments": {}, //command arguments depends on the command itself
	"queue": "optional-queue",
	"stats_interval": 0, //optional stats gathering interval (falls to default if not set)
	"max_time": 0, //Max run time of the command, if exceeded command will be killed
	"max_restart": 0, //Max number of retries to start the command if failed before giving up
	"recurring_period": 0, //If provided command is considered recurring
	"log_levels": [int] //Log levels to store locally and not discard.
}
```

The `Core0` Core understands a very specific set of management commands:

- Basic Commands
    - core.ping
    - core.system
    - core.kill
    - core.killall
    - core.state
    - core.reboot
- Info Query
    - info.cpu
    - info.disk
    - info.mem
    - info.nic
    - info.os
- CoreX Management
    - corex.create
    - corex.list
    - corex.dispatch
    - corex.terminate
- Bridge
    - bridge.create
    - bridge.list
    - bridge.delete
- Disk Management
    - disk.list
    - disk.mktable
    - disk.mkpart
    - disk.rmpart
    - disk.mount
    - disk.umount
- Btrfs Management
    - btrfs.create
    - btrfs.list
    - btrfs.subvol_create
    - btrfs.subvol_list
    - btrfs.subvol_delete

## Commands arguments
### core.ping
Doesn't take any arguments. returns a "pong". Main use case is to check the the core is responding.

### core.system
Arguments:
```javascript
{
	"name": "executable",
	"dir": "pwd",
	"args": ["command", "arguements"]
	"env": {"ENV1": "VALUE1", "ENV2": "VALUE2"},
	"stdin": "data to pass to executable over stdin"
}
```
Executes an arbitrary command

### core.kill
Arguments:
```javascript
{
    "id": "process-id-to-kill"
}
```
Kills a certain process giving the process ID. The process/command id is the id of the command used to start this process
in the first place.

### core.killall
Takes no arguments
Kills ALL processes on the system. (only the ones that where started by core0 itself) and still running by the time of calling this command

### core.state
Takes no arguments.
Returns aggregated state of all processes plus the consumption of core0 itself (cpu, memory, etc...)

### core.reboot
Takes no arguments.
Immediately reboot the machine.

### info.cpu
Takes no arguments.
Returns information about the host CPU types, speed and capabilities

### info.disk
Takes no arguments.
Returns information about the host attached disks

### info.mem
Takes no arguments.
Returns information about the host memory.

### info.nic
Takes no arguments.
Returns information about the host attached nics, and IPs

### info.os
Takes no arguments.
Returns information about the host OS.

### corex.create
Arguments:
```javascript
{
    "plist": "http://url/to/plist",
    "mount": {
        "/host/directory": "/container/directory"
    },
    "network": {
        "zerotier": "zerotier network id", //options
        "bridge": [], //list of bridges names to connect to
    }
    //TODO:
    "port": {
        host_port: container_port,
    }
}
```

### corex.list
Takes no arguments.
List all available `live` Cores on a host.

### corex.terminate
Arguemnts:
```javascript
{
    "container": container_id,
}
```
Destroys a Core and stops the Core processes. It takes a mandatory core ID.

### corex.dispatch
Arguments:
```javascript
{
     "container": core_id,
     "command": {
         //the full command payload
     }
}
```

### bridge.create
Arguments:
```javascript
{
    "name": "bridge-name", //required
    "hwaddr": "MAC address" //optional
}
```
Creates a new bridge

### bridge.list
takes no argumetns
List all available bridges

### bridge.delete
Arguments:
```javascript
{
    "name": "bridge-name", //required
}
```
Delete the given bridge name

### disk.list
Takes no arguments.
List all block devices (similar to lsblk)

### disk.mktable
Arguments:
```javascript
{
    "disk": "/dev/disk", //device
    "table_type": "gpt", //partition table type
}
```
Creates a new partition table on device. `table_type` can be any value 
that is supported by `parted mktable`

### disk.mkpart
Arguments:
```javascript
{
    "disk": "/dev/disk", //device
    "part_type": "primary", //part_type
    "start": "1", //start sector
    "end": "100%", //end sector
}
```
Creates a partition on given device. `part_type`, `start`, and `end` values must
be supported by the `parted mkpart` command

### disk.rmpart
Arguments:
```javascript
{
    "disk": "/dev/disk", //device
    "number": 1, //parition number (1 based index)
}
```
Removes a partition on given device with given 1 based index.

### disk.mount
Arguments:
```javascript
{
    "options": "auto", //mount options (required) if no options are needed set to "auto"
    "source": "/dev/part", //patition to mount
    "target": "/mnt/data", //location to mount on
}
```

### disk.umount
Arguments:
```javascript
{
    "source": "/dev/part", //partition to umount
}
```

### btrfs.create

Create a btrfs filesystem

Arguments:
```javascript
{
    "label": "FS label/name", // required
    "devices": ["/dev/sdc1", "/dev/sdc2"], // the devices, required
    "data": "data profile",
    "metadata": "metadata profile"
}
```

### btrfs.list

List all btrfs filesystems.

Takes no argument. Return array of all filesystems.

### btrfs.subvol_create

Creates a new btrfs subvolume

arguments:
```javascript
{
    "path": "/path/of/subvolume" required
}
```

### btrfs.subvol_list

List subvolume under a path

arguments:
```javascript
{
    "path": "/path/of/filesystem" required
}
```

### btrfs.subvol_delete

Delete a btrfs subvolume

arguments:
```javascript
{
    "path": "/path/of/subvolume" required
}
```
