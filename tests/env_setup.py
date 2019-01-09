#python script for 0-core testcases
from jumpscale import j
import os
from argparse import ArgumentParser
from subprocess import Popen, PIPE
import random
import uuid
import time
import shlex


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

    def stream_run_cmd(self, cmd):
        sub = Popen(shlex.split(cmd), stdout=PIPE)
        while True:
            out = sub.stdout.readline()
            if out == b'' and sub.poll() is not None:
                break
            if out:
                print(out.strip())
        rc = sub.poll()
        return rc

    def send_script_to_remote_machine(self, script, ip, port):
        templ = 'scp -o StrictHostKeyChecking=no -r -o UserKnownHostsFile=/dev/null -P {} {} root@{}:'
        cmd = templ.format(port, script, ip)
        self.run_cmd(cmd)

    def run_cmd_on_remote_machine(self, cmd, ip, port):
        templ = 'ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p {} root@{} {}'
        cmd = templ.format(port, ip, cmd)
        return self.stream_run_cmd(cmd)

    def create_disk(self, zos_client):
        zdb_name = str(uuid.uuid4())[0:8]
        mount_path = zos_client.zerodbs.prepare()
        zdb = zos_client.primitives.create_zerodb(name=zdb_name, path=mount_path[0],
                                                  node_port=random.randint(4000, 5000),
                                                  mode='user', sync=False, admin='mypassword')
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

    # get zeroos host client
    zos_client = j.clients.zos.get('zos-kds-farm', data={'host': '{}'.format(options.zos_ip)})
    vm_zos_name = os.environ['vm_zos_name']
    vm_ubuntu_name = os.environ['vm_ubuntu_name']
    rand_num = random.randint(3, 125)
    vm_zos_ip = '10.100.{}.{}'.format(rand_num, random.randint(3, 125))
    vm_ubuntu_ip = '10.100.{}.{}'.format(rand_num, random.randint(126, 253))

    script = """
apt-get install git python3-pip -y
git clone https://github.com/threefoldtech/0-core.git
cd 0-core; git checkout %s; pip3 install client/py-client/.
cd tests
pip3 install -r requirements.txt
sed -i -e"s/^target_ip=.*/target_ip=%s/" config.ini
sed -i -e"s/^zt_access_token=.*/zt_access_token=%s/" config.ini
interface=$(ip a | grep 3: | awk '{print $2}')
dhclient $interface
    """ % (options.branch, vm_zos_ip, options.zt_token)

    # create a bridge and assign specific ips for the vms
    bridge = os.environ['bridge']
    vm_zos_mac = utils.random_mac()
    vm_ubuntu_mac = utils.random_mac()
    cidr = '10.100.{}.1/24'.format(rand_num)
    start = '10.100.{}.2'.format(rand_num)
    end = '10.100.{}.254'.format(rand_num)
    zos_client.client.bridge.create(bridge, network='dnsmasq', nat=True,
                                    settings={'cidr': cidr, 'start': start, 'end': end})
    zos_client.client.json('bridge.host-add', {'bridge': bridge, 'ip': vm_zos_ip, 'mac': vm_zos_mac})
    zos_client.client.json('bridge.host-add', {'bridge': bridge, 'ip': vm_ubuntu_ip, 'mac': vm_ubuntu_mac})

    # create a zeroos vm
    print('* Creating zero-os vm')
    print('zos_vm ip: ' + vm_zos_ip)
    vm_zos = zos_client.primitives.create_virtual_machine(name=vm_zos_name, type_='zero-os:{}'.format(options.branch))
    vm_zos.nics.add(name='nic1', type_='bridge', networkid=bridge, hwaddr=vm_zos_mac)
    vm_zos.vcpus = 4
    vm_zos.memory = 8192
    vm_zos.kvm = True
    disk = utils.create_disk(zos_client)
    vm_zos.disks.add(disk)
    vm_zos.kernel_args.add(name='development', key='development')
    vm_zos.deploy()

    # create sshkey and provide the public key
    keypath = '/root/.ssh/id_rsa.pub'
    if not os.path.isfile(keypath):
        os.system("echo  | ssh-keygen -P ''")
    with open(keypath, "r") as key:
        pub_key = key.read()
    pub_key.replace('\n', '')

    # create an ubuntu vm to run testcases from
    print('* Creating ubuntu vm to fire the testsuite from')
    print('ubuntu_vm ip: ' + vm_ubuntu_ip)
    ubuntu_port = int(os.environ['ubuntu_port'])
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
    script_path = '/tmp/setup_env.sh'
    with open(script_path, "w") as f:
        f.write(script)
    utils.send_script_to_remote_machine(script_path, options.zos_ip, ubuntu_port)

    print('* Setup the environment')
    cmd = 'bash setup_env.sh'
    utils.run_cmd_on_remote_machine(cmd, options.zos_ip, ubuntu_port)
    time.sleep(30)

    # Make sure zrobot server is not running
    zrobot_kill = """
from zeroos.core0 import client
cl = client.Client('%s')
cont = cl.container.find('zrobot')
if cont:
    cont_cl = cl.container.client(1)
    cl.bash("echo $'import time; time.sleep(1000000000)' > /mnt/containers/1/.startup.py").get()
    out = cont_cl.bash('ps aux | grep server').get().stdout
    cont_cl.bash('kill -9 {}'.format(out.split()[1])).get()
    """ % (vm_zos_ip)

    time.sleep(5)
    script_path = '/tmp/zrobot_kill.py'
    with open(script_path, "w") as f:
        f.write(zrobot_kill)
    utils.send_script_to_remote_machine(script_path, options.zos_ip, ubuntu_port)

    print('* killing zrobot server')
    cmd = 'python3 zrobot_kill.py'
    utils.run_cmd_on_remote_machine(cmd, options.zos_ip, ubuntu_port)


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
