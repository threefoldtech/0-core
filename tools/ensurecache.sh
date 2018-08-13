set -ex

udevadm settle

LABEL="sp_zos-cache"
CACHE="/var/cache"
STORAGEPOOL="/mnt/storagepools"

MNT="${STORAGEPOOL}/${LABEL}"

function error {
    echo "[-]" $@ >&2
}

function log {
    echo "[+]" $@ >&2
}

function labelmount {
    disk=/dev/disk/by-label/$1
    target=$2

    btrfs check --repair $disk
    mount $disk $target
}

function preparedisk {
    DISK=""
    for disk in `lsblk -e 2 -e 11 -pdn -o NAME,ROTA,TYPE,TRAN|grep -v usb|sort -nk 2|cut -d " " -f 1`; do
        if ! lsblk -n -o TYPE ${disk} | grep part > /dev/null; then
            # disk does not have any parition
            DISK=$disk
            break
        fi
    done

    if [ "$DISK" == "" ]; then
        error "no free disks found"
        return 1
    fi

    parted -s ${DISK} mktable gpt
    parted -s ${DISK} mkpart primary btrfs 1 100% | true
    sync
    partprobe
    udevadm settle

    mkfs.btrfs ${DISK}1 -f -L ${LABEL}

    sync
    partprobe
    udevadm settle

    return 0
}

function cleanup {
    path=$1
    if [ ! -d $path ]; then
        return 0
    fi

    for vol in `ls $path`; do
        full="$path/$vol"
        btrfs subvol del $full | rm -rf $full | true
    done
    return 0
}

function hook {
    # create required subvols and mount them if not exits
    log "create and mount subvolume for ${LABEL}"
    # 1 - cache subvol
    btrfs subvol create $1/cache || true
    mount $1/cache /var/cache/

    # clean up old container, and vms working directories
    cleanup /var/cache/containers
    cleanup /var/cache/vms

    logs=$1/logs
    btrfs subvol create ${logs} || true
    current="log-$(date +%Y%m%d-%H%M)"
    btrfs subvol create ${logs}/${current}
    cp -a /var/log/* ${logs}/${current}/
    mount ${logs}/${current} /var/log
    kill -USR1 1 #log rotation
    return 0
}

function sharedcache {
    # try to mount the shared cache if possible
    CACHEPATH=/var/cache/zerofs
    if [ ! -d ${CACHEPATH} ]; then
        mkdir -p ${CACHEPATH}
    fi

    if ! mount -t 9p zoscache ${CACHEPATH}; then
        log "No shared cache exposed to the node"
    fi
}

function main {
    mkdir -p ${MNT}
    if mountpoint -q ${MNT}; then
        error "${MNT} is already mounted"
        exit 1
    fi

    if ! labelmount ${LABEL} ${MNT}; then
        # no parition found with that label
        # prepare the first availabel disk
	    log "${LABEL} not mounted, search for available disk"
        if preparedisk; then
            labelmount ${LABEL} ${MNT}
        fi
    fi

    if mountpoint -q ${MNT}; then
        hook ${MNT}
    fi

    sharedcache
}

main
