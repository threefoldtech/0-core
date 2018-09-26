from argparse import ArgumentParser
from jumpscale import j
from env_setup import Utils
import os


def main(options):
    if not options.teardown:
        utils = Utils(options)
        cmd = "cd cd 0-core/tests; nosetests -v -s testsuite --tc-file config.ini"
        utils.run_cmd_on_remote_machine(cmd, options.zos_ip, int(os.environ['ubuntu_port']))
    else:
        zos_client = j.clients.zos.get('zos-kds-farm', data={'host': '{}'.format(options.zos_ip)})
        vms = zos_client.client.kvm.list()
        vm_uuid = [vm['uuid'] for vm in vms if vm['name'] == os.environ['vm_ubuntu_name']]
        if vm_uuid:
            zos_client.client.kvm.destroy(vm_uuid[0])
        vm_uuid = [vm['uuid'] for vm in vms if vm['name'] == os.environ['vm_zos_name']]
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
    parser.add_argument('--teardown', dest='teardown', default=False, action='store_true',
                        help='if "--teardown" flag is passed, the 2 vms and the bridge will be removed')
    options = parser.parse_args()
    main(options)
