# KVM Commands

Available commands:

- [experimental.kvm.create](#create)
- [experimental.kvm.destroy](#destroy)
- [experimental.kvm.list](#list)


<a id="create"></a>
## experimental.kvm.create

Arguments:
```javascript
{
  'name': {name},
  'image': {image},
  'cpu': {cpu},
  'memory': {memory},
  'bridge': {bridge},
}
```


Values:

- **name**: Name of the virtual machine
- **image**: Name of image to use
- **cpu**: Number of virtual CPU core, e.g. `2`
- **memory**: memory in MB, e.g. `512`
- **bridge**: bridge name, e.g. `none`


<a id="destroy"></a>
## experimental.kvm.destroy

Destroys a given virtual machine.


<a id="list"></a>
## experimental.kvm.list

Lists all virtual machines.
