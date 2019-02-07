from argparse import ArgumentParser
from zeroos.core0 import client
import os
import requests


def teardown(options):
    zos_vm_name = os.environ['vm_zos_name']
    # get zeroos host client
    params = {'grant_type': 'client_credentials','client_id': options.client_id,
              'client_secret': options.client_secret,'response_type': 'id_token',
              'scope': 'user:memberof:threefold.sysadmin,offline_access'}
    jwt = requests.post('https://itsyou.online/v1/oauth/access_token', params=params).text
    zos_client = client.Client(options.zos_ip, password=jwt)

    # remove vms and bridge
    print('removing vms and bridge')
    vms = zos_client.kvm.list()
    vm_uuid = [vm['uuid'] for vm in vms if vm['name'] == os.environ['vm_ubuntu_name']]
    if vm_uuid:
        zos_client.kvm.destroy(vm_uuid[0])
    vm_uuid = [vm['uuid'] for vm in vms if vm['name'] == zos_vm_name]
    if vm_uuid:
        zos_client.kvm.destroy(vm_uuid[0])
    bridge = os.environ['bridge']
    if bridge in zos_client.bridge.list():
        zos_client.bridge.delete(bridge)

    # remove the zos_vm disk from the host 
    print('removing zos_vm disk')
    zos_client.bash('rm -rf /var/cache/{}.qcow2'.format(os.environ['ubuntu_port']))

    # leave the zt-network
    os.system('sudo zerotier-cli leave {}'.format(os.environ['ZT_NET_ID']))


if __name__ == "__main__":
    parser = ArgumentParser()
    parser.add_argument("-z", "--zos_ip", type=str, dest="zos_ip", required=True,
                        help="IP of the zeroos machine that will be used")
    parser.add_argument("-d", "--client_id", type=str, dest="client_id",
                        help="client id to generate token for zos client")
    parser.add_argument("-s", "--client_secret", type=str, dest="client_secret",
                        help="client secret to generate token for zos client")
    options = parser.parse_args()
    teardown(options)
