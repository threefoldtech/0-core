# Btrfs Commands

Available commands:

- [btrfs.create](#create)
- [btrfs.device_add](#device_add)
- [btrfs.remove](#device_remove)
- [btrfs.list](#list)
- [btrfs.info](#info)
- [btrfs.subvol_create](#subvol_create)
- [btrfs.subvol_list](#subvol_list)
- [btrfs.subvol_delete](#subvol_delete)
- [btrfs.subvol_snapshot](#subvol_snapshot)


<a id="create"></a>
## btrfs.create

Creates a Btrfs filesystem with the given label, devices, and profiles.

Arguments:
```javascript
{
    "label": "{label}",
    "devices": {devices},
    "data": "{data-profile}",
    "metadata": "{metadata-profile}",
    "overwrite": "{overwrite}",
}
```

Values:
- **label**: Name/label
- **devices**: Array of devices, e.g. `["/dev/sdc1", "/dev/sdc2"]`
- **metadata-profile**: `raid0`, `raid1`, `raid5`, `raid6`, `raid10`, `dup` or `single`
- **data-profile**: Same as metadata-profile
- **overwrite**: If true forces creation of the filesystem, overwriting any existing filesystem


<a id="device_add"></a>
## btrfs.device_add

Adds one or more devices to the Btrfs filesystem mounted under a given `mountpoint`.

Arguments:
```javascript
{
  'mountpoint': {mount-point},
  'devices': {devices},
}
```

Values:
- **mount-point**: Mount point of the Btrfs system
- **devices**: One ore more devices to add


<a id="device_remove"></a>
## btrfs.device_remove

Removes one or more devices from the Btrfs filesystem mounted under `mountpoint`.

Arguments:
```javascript
{
  'mountpoint': {mount-point},
  'devices': {devices},
}
```

Values:
- **mount-point**: Mount point of the Btrfs system
- **devices**: One ore more devices to remove


<a id="list"></a>
## btrfs.list

Lists all Btrfs filesystems. It takes no arguments. Return array of all filesystems.


<a id="info"></a>
## btrfs.info

Gets Btrfs info. It takes no arguments.


<a id="subvol_create"></a>
## btrfs.subvol_create

Creates a new Btrfs subvolume in the specified path.

arguments:
```javascript
{
    "path": "{path}"
}
```

Values:
- **path**: Path where to create the subvolume, e.g. `/path/of/subvolume`

<a id="subvol_list"></a>
## btrfs.subvol_list

Lists all subvolumes under a given path.

arguments:
```javascript
{
    "path": "{path}"
}
```

Values:
- **path**: Path to list


<a id="subvol_delete"></a>
## btrfs.subvol_delete

Deletes a Btrfs subvolume in the specified path.

arguments:
```javascript
{
    "path": "{path}"
}
```

Values:
- **path**: Path where to deleted the subvolumes


<a id="subvol_snapshot"></a>
## btrfs.subvol_snapshot

Takes a snapshot.

arguments:
```javascript
{
  "source": {source},
  "destination": {destination},
  "read_only": {read_only},
}
```

Values:
- **source** Source path of subvolume
- **destination**: Destination path of the snapshot
- **readonly**: Set read-only on the snapshot
