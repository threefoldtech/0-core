from argparse import ArgumentParser
from jumpscale import j
import os


def teardown(options):
    zos_vm_name = os.environ['vm_zos_name']
    zos_client = j.clients.zos.get('zos-kds-farm', data={'host': '{}'.format(options.zos_ip)})
    print('Removing namespace and zdb')
    # remove zdb
    zos_vm = zos_client.client.kvm.get(zos_vm_name)
    zdb_cont_ip = zos_vm['params']['media'][0]['url'].split('zdb://')[1].split(':')[0]
    zdbs = zos_client.zerodbs.list()
    for zdb in zdbs:
        zdb_ip = str(zdb.container.default_ip()).split('/')[0]
        if zdb_ip == zdb_cont_ip:
            zdb.destroy()
            break
    # remove namespace
    zos_client.client.bash('rm -rf /mnt/zdbs/sda/data/mydisk').get()
    # remove vms and bridge
    print('removing vms and bridge')
    vms = zos_client.client.kvm.list()
    vm_uuid = [vm['uuid'] for vm in vms if vm['name'] == os.environ['vm_ubuntu_name']]
    if vm_uuid:
        zos_client.client.kvm.destroy(vm_uuid[0])
    vm_uuid = [vm['uuid'] for vm in vms if vm['name'] == zos_vm_name]
    if vm_uuid:
        zos_client.client.kvm.destroy(vm_uuid[0])
    bridge = os.environ['bridge']
    if bridge in zos_client.client.bridge.list():
        zos_client.client.bridge.delete(bridge)
    # leave the zt-network
    os.system('zerotier-cli leave {}'.format(os.environ['ZT_NET_ID']))


if __name__ == "__main__":
    parser = ArgumentParser()
    parser.add_argument("-z", "--zos_ip", type=str, dest="zos_ip", required=True,
                        help="IP of the zeroos machine that will be used")
    options = parser.parse_args()
    teardown(options)
