#python script for 0-core testcases on kds farm
from jumpscale import j
import os
from argparse import ArgumentParser
from subprocess import Popen, PIPE
import random
import uuid
import time


class Utils(object):
    def __init__(self, options):
        self.options = options

    def run_cmd(self, cmd, timeout=20):
        now = time.time()
        while time.time() < now + timeout:
            sub = Popen([cmd], stdout=PIPE, stderr=PIPE, shell=True)
            out, err = sub.communicate()
            if sub.returncode == 0:
                return out.decode('utf-8')
            elif any(x in err.decode('utf-8') for x in ['Connection refused', 'No route to host']):
                time.sleep(1)
                continue
            else:
                break
        raise RuntimeError("Failed to execute command.\n\ncommand:\n{}\n\n{}".format(cmd, err.decode('utf-8')))

    def send_script_to_remote_machine(self, script, ip, port):
        templ = 'scp -o StrictHostKeyChecking=no -r -o UserKnownHostsFile=/dev/null -P {} {} root@{}:'
        cmd = templ.format(port, script, ip)
        self.run_cmd(cmd)

    def run_cmd_on_remote_machine(self, cmd, ip, port):
        templ = 'ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p {} root@{} {}'
        cmd = templ.format(port, ip, cmd)
        return self.run_cmd(cmd)

    def create_disk(self, zos_client):
        zdb = zos_client.primitives.create_zerodb(name='myzdb', path='/mnt/zdbs/sda',
                                                  mode='user', sync=False, admin='mypassword')
        zdb.namespaces.add(name='mynamespace', size=50, password='namespacepassword', public=True)
        zdb.deploy()
        disk = zos_client.primitives.create_disk('mydisk', zdb, size=50)
        disk.deploy()
        return disk

    def random_mac(self):
        return "52:54:00:%02x:%02x:%02x" % (random.randint(0, 255),
                                            random.randint(0, 255),
                                            random.randint(0, 255))


def main(options):
    utils = Utils(options)

    # make sure zerotier is installed
    if not os.path.exists('/usr/sbin/zerotier-one'):
        os.system('curl -s https://install.zerotier.com/ | sudo bash')
    os.system('zerotier-cli join {}'.format(options.zerotier_id))

    # get zeroos host client
    zos_client = j.clients.zos.get('zos-kds-farm', data={'host': '{}'.format(options.zos_ip)})
    vm_zos_name = 'vm-zeroos'
    vm_ubuntu_name = 'vm-ubuntu2'
    vm_zos_ip = '10.100.20.{}'.format(random.randint(3, 125))
    vm_ubuntu_ip = '10.100.20.{}'.format(random.randint(126, 253))
    script = """
apt-get install git python3-pip -y
git clone https://github.com/threefoldtech/0-core.git
cd 0-core; git checkout {}; pip3 install client/py-client/.
cd tests
pip3 install -r requirements.txt
sed -i -e"s/^target_ip=.*/target_ip={}/" config.ini
sed -i -e"s/^zt_access_token=.*/zt_access_token={}/" config.ini
nosetests -v -s testsuite/a_basic/tests_01_system.py --tc-file config.ini
    """.format(options.branch, vm_zos_ip, options.zt_token)

    try:
        # create a bridge and assign specific ips for the vms
        bridge = str(uuid.uuid4())[0:8]
        vm_zos_mac = utils.random_mac()
        vm_ubuntu_mac = utils.random_mac()
        zos_client.client.bridge.create(bridge, network='dnsmasq', nat=True, settings={'cidr': '10.100.20.1/24',
                                        'start': '10.100.20.2', 'end': '10.100.20.254'})
        zos_client.client.json('bridge.host-add', {'bridge': bridge, 'ip': vm_zos_ip, 'mac': vm_zos_mac})
        zos_client.client.json('bridge.host-add', {'bridge': bridge, 'ip': vm_ubuntu_ip, 'mac': vm_ubuntu_mac})

        # create a zeroos vm
        vm_zos = zos_client.primitives.create_virtual_machine(name=vm_zos_name, type_='zero-os:{}'.format(options.branch))
        vm_zos.nics.add(name='nic1', type_='bridge', networkid=bridge, hwaddr=vm_zos_mac)
        vm_zos.vcpus = 4
        vm_zos.memory = 8192
        disk = utils.create_disk(zos_client)
        vm_zos.disks.add(disk)
        vm_zos.deploy()

        # create sshkey and provide the public key
        keypath = '/root/.ssh/id_rsa.pub'
        if not os.path.isfile(keypath):
            os.system("echo  | ssh-keygen -P ''")
        with open(keypath, "r") as key:
            pub_key = key.read()
        pub_key.replace('\n', '')

        # create an ubuntu vm to run testcases from
        ubuntu_port = random.randint(2222, 3333)
        vm_ubuntu = zos_client.primitives.create_virtual_machine(name=vm_ubuntu_name, type_='ubuntu:lts')
        vm_ubuntu.nics.add(name='nic2', type_='default')
        vm_ubuntu.nics.add(name='nic3', type_='bridge', networkid=bridge, hwaddr=vm_ubuntu_mac)
        vm_ubuntu.configs.add('sshkey', '/root/.ssh/authorized_keys', pub_key)
        vm_ubuntu.ports.add('port2', ubuntu_port, 22)
        vm_ubuntu.vcpus = 4
        vm_ubuntu.memory = 8192
        vm_ubuntu.deploy()

        # access the ubuntu vm and start ur testsuite
        time.sleep(10)
        script_path = '/tmp/runtests.sh'
        with open(script_path, "w") as f:
            f.write(script)
        utils.send_script_to_remote_machine(script_path, options.zos_ip, ubuntu_port)

        cmd = 'bash runtests.sh'
        utils.run_cmd_on_remote_machine(cmd, options.zos_ip, ubuntu_port)

    except:
        raise

    finally:
        # tear down
        # delete the vms and bridge
        vms = zos_client.client.kvm.list()
        vm_uuid = [vm['uuid'] for vm in vms if vm['name'] == vm_ubuntu_name]
        if vm_uuid:
            zos_client.client.kvm.destroy(vm_uuid[0])
        vm_uuid = [vm['uuid'] for vm in vms if vm['name'] == vm_zos_name]
        if vm_uuid:
            zos_client.client.kvm.destroy(vm_uuid[0])
        if bridge in zos_client.client.bridge.list():
            zos_client.client.bridge.delete(bridge)
        # leave the zt-network
        os.system('zerotier-cli leave {}'.format(options.zerotier_id))


if __name__ == "__main__":
    parser = ArgumentParser()
    parser.add_argument("-z", "--zos_ip", type=str, dest="zos_ip", required=True,
                        help="IP of the zeroos machine that will be used")
    parser.add_argument("-i", "--zerotierid", type=str, dest="zerotier_id", required=True,
                        help="zerotier netowrkid that the zos node is joining")
    parser.add_argument("-b", "--branch", type=str, dest="branch", required=True,
                        help="0-core branch that the tests will run from")
    parser.add_argument("-t", "--zt_token", type=str, dest="zt_token", default='sgtQtwEMbRcDgKgtHEMzYfd2T7dxtbed', required=True,
                        help="zerotier token that will be used for the core0 tests")
    options = parser.parse_args()
    main(options)
