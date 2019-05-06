from utils.utils import BaseTest
from git import Repo
import unittest
import time
from zeroos.core0 import client
from random import randint
import random


class SystemTests(BaseTest):

    def setUp(self):
        super(SystemTests, self).setUp()
        self.check_zos_connection(SystemTests)
        self.zos_flist = 'https://hub.grid.tf/tf-autobuilder/zero-os-development.flist'

    def ping_zos(self, client, timeout=60):
        now = time.time()
        while now + timeout > time.time():
            try:
                client.ping()
                return True
            except:
                continue
        return False

    def test001_update_zos_versoin(self):
        """ zos-052
        *Test case for updating zeroos version

        **Test Scenario:**
        #. Create a vm, should succeed.
        #. Update zos vm version.
        #. Wait till the node is back, then check if the version has been updated.
        """
        self.lg('Create a vm, should succeed')
        vm_name = self.rand_str()
        bridge = self.rand_str()
        vm_zos_mac = self.random_mac()
        rand_num = random.randint(100,200)
        cidr = '20.201.{}.1/24'.format(rand_num)
        start = '20.201.{}.2'.format(rand_num)
        end = '20.201.{}.254'.format(rand_num)
        vm_zos_ip = '20.201.{}.21'.format(rand_num)
        self.client.bridge.create(bridge, network='dnsmasq', nat=True,
                                  settings={'cidr': cidr, 'start': start, 'end': end})
        self.client.json('bridge.host-add', {'bridge': bridge, 'ip': vm_zos_ip, 'mac': vm_zos_mac})
        nics = [{'type': 'bridge', 'id': bridge, 'hwaddr': vm_zos_mac}]
        pub_port = randint(4000,5000)

        self.client.bash('qemu-img create -f qcow2 /var/cache/{}.qcow2 5G'.format(pub_port)).get()

        vm_uuid = self.create_vm(name=vm_name, flist=self.zos_flist, media=[{'url': '/var/cache/{}.qcow2'.format(pub_port)}],
                                 memory=2048, nics=nics, cmdline='development')
        res = self.client.system('nft add rule ip nat pre ip daddr @host tcp dport {} mark set 0xff000000 dnat to {}:6379'.format(pub_port, vm_zos_ip)).get() 

        self.assertEqual(res.state, 'SUCCESS')
        time.sleep(60)
        vm_cl = client.Client(self.target_ip, port=pub_port)

        self.lg('Update zos node')
        rs = vm_cl.ping()
        self.assertEqual(rs[:4], 'PONG')
        cur_version = rs.split()[2]
        r = Repo('../.')
        for branch in r.remotes.origin.refs:
            if cur_version != branch.remote_head:
                new_version = branch.remote_head
                break
        vm_cl.power.update('zero-os-{}.efi'.format(new_version))
        time.sleep(10)

        self.lg('Wait till the node is back, then check if the version has been updated.')
        res = self.ping_zos(vm_cl, timeout=300)
        self.assertTrue(res, "Can't ping zos node")
        rs = vm_cl.ping()
        self.assertEqual(rs.split()[2], new_version)

        self.lg('Destroy the vm')
        self.client.kvm.destroy(vm_uuid)
