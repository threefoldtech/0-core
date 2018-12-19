# Power Commands

Available commands:

- [core.reboot](#reboot)
- [core.poweroff](#poweroff)
- [core.update](#update)

<a id="reboot"></a>
## core.reboot
Full reboot of the node. Takes no argument, and perform an immediate full reboot of the node


<a id="poweroff"></a>
## core.poweroff
Full power off of the node. Takes no argument, and perform an immediate full power off of the node

<a id="update"></a>
## core.update
Update the node with given image, and fast reboot into this image. No hardware reboot will happen

Arguments:
```javascript
{
  'image': {image},
}
```
- **image** efi image name, the image will be downloaded from https://bootstrap.grid.tf/kernel . example: `zero-os-development.efi`