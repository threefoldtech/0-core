package builtin

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseMountCmd(t *testing.T) {
	mount := `/dev/sda2 on /var/lib/docker/plugins type ext4 (rw,relatime,errors=remount-ro,data=ordered)
/dev/sda2 on /var/lib/docker/aufs type ext4 (rw,relatime,errors=remount-ro,data=ordered)
	`
	mounts := parseMountCmd(mount)
	mountpoints, exists := mounts["/dev/sda2"]
	if ok := assert.Equal(t, true, exists); !ok {
		t.Fatal()
	}
	if ok := assert.Equal(t, 2, len(mountpoints)); !ok {
		t.Fatal()
	}

	pointOne := mountpoints[0]
	if ok := assert.Equal(t, "/var/lib/docker/plugins", pointOne.Mountpoint); !ok {
		t.Fatal()
	}
	if ok := assert.Equal(t, "ext4", pointOne.Filesystem); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, "1", pointOne.Options["rw"]); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, "ordered", pointOne.Options["data"]); !ok {
		t.Fatal()
	}

	pointTwo := mountpoints[1]
	if ok := assert.Equal(t, "/var/lib/docker/aufs", pointTwo.Mountpoint); !ok {
		t.Fatal()
	}

}
