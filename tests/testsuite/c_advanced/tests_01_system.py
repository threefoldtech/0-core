from utils.utils import BaseTest
from git import Repo
import unittest
import time
from zeroos.core0 import client
from random import randint


class SystemTests(BaseTest):

    def setUp(self):
        super(SystemTests, self).setUp()
        self.check_zos_connection(SystemTests)
        # Assuming development has client.power.update()
        self.zos_flist = 'https://hub.grid.tf/tf-autobuilder/zero-os-development.flist'

    def ping_zos(self, timeout=60):
        now = time.time()
        while now + timeout > time.time():
            try:
                self.client.ping()
                return True
            except:
                continue
        return False

    @unittest.skip('https://github.com/threefoldtech/0-core/issues/96')
    def test001_update_zos_versoin(self):

        """ zos-052
        *Test case for updating zeroos version

        **Test Scenario:**
        #. Create a vm, should succeed.
        #. Update zos vm version.
        #. Wait till the node is back, then check if the version has been updated.
        #. Update the node back to its old version and check again.
        """
        self.lg('Create a vm, should succeed')
        vm_name = self.rand_str()
        nics = [{'id': 'None', 'type': 'default'}]
        pub_port = 4444
        ports = {pub_port:6379}
        vm_uuid = self.create_vm(name=vm_name, flist=self.zos_flist
                                 ports=ports, nics=nics)
        time.sleep(10)
        vm_cl = client.Client(self.target_ip, port=pub_port)

        self.lg('Update zos node')
        rs = vm_cl.ping()
        self.assertEqual(rs[:4], 'PONG')
        cur_version = rs.split()[2]
        r = Repo('../.')
        r.remotes.origin.refs
        for branch in r.remotes.origin.refs:
            if cur_version != branch.remote_head:
                new_version = branch.remote_head
                break
        vm_cl.power.update('zero-os-{}.efi'.format(new_version))
        time.sleep(10)

        self.lg('Wait till the node is back, then check if the version has been updated.')
        res = self.ping_zos(timeout=300)
        self.assertTrue(res, "Can't ping zos node")
        rs = vm_cl.ping()
        self.assertEqual(rs.split()[2], new_version)
