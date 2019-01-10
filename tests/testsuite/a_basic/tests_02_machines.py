from utils.utils import BaseTest
import time
import unittest
from random import randint


class Machinetests(BaseTest):

    def setUp(self):
        super(Machinetests, self).setUp()
        self.check_zos_connection(Machinetests)

    def test001_create_destroy_list_kvm(self):
        """ zos-009

        *Test case for testing creating, listing and destroying VMs*

        **Test Scenario:**

        #. Check that system support hardware virtualization
        #. Create virtual machine (VM1), should succeed
        #. Create another vm with the same name, should fail
        #. List all virtual machines and check that VM1 is there
        #. Create another virtual machine with the same kvm domain, should fail
        #. Destroy VM1, should succeed
        #. List the virtual machines, VM1 should be gone
        #. Destroy VM1 again, should fail
        """

        self.lg('{} STARTED'.format(self._testID))
        vm_name = self.rand_str()

        self.lg('- Check that it support hardware virtualization ')
        responce = self.client.info.cpu()
        vmx = ['vmx'or'svm' in dec['flags'] for dec in responce]
        self.assertGreater(len(vmx), 0)

        self.lg('- Create virtual machine {} , should succeed'.format(vm_name))
        time.sleep(4)
        vm_uuid = self.create_vm(name=vm_name, flist=self.ubuntu_flist)

        self.lg('Create another vm with the same name, should fail')
        with self.assertRaises(RuntimeError):
            self.create_vm(name=vm_name, flist=self.ubuntu_flist)

        self.lg('- List all virtual machines and check that VM {} is there '.format(vm_name))
        Vms_list = self.client.kvm.list()
        self.assertTrue(any(vm['name'] == vm_name for vm in Vms_list))

        self.lg('- create another virtual machine with the same kvm domain ,should fail')
        with self.assertRaises(RuntimeError):
            self.create_vm(name=vm_name, flist=self.ubuntu_flist)

        self.lg('- Destroy VM {}'.format(vm_name))
        self.client.kvm.destroy(vm_uuid)

        self.lg('- List the virtual machines , VM {} should be gone'.format(vm_name))
        Vms_list = self.client.kvm.list()
        self.assertFalse(any(vm['name'] == vm_name for vm in Vms_list))

        self.lg('- Destroy VM {} again should fail'.format(vm_name))
        with self.assertRaises(RuntimeError):
            self.client.kvm.destroy(vm_uuid)

        self.lg('{} ENDED'.format(self._testID))

    def test002_create_list_delete_containers(self):
        """ zos-010
        *Test case for testing creating, listing and deleting containers*

        **Test Scenario:**
        #. Create a new container (C1), should succeed
        #. List all containers and check that C1 is there
        #. Destroy C1, should succeed
        #. List the containers, C1 should be gone
        #. Destroy C1 again, should fail

        """
        self.lg('{} STARTED'.format(self._testID))
        self.lg('Create a new container (C1)')
        C1 = self.create_container(root_url=self.root_url, storage=self.storage)

        self.lg('List all containers and check that C1 {}is there'.format(C1))
        containers = self.client.container.list()
        self.assertTrue(str(C1) in containers)

        self.lg('Destroy C1 {}, should succeed'.format(C1))
        time.sleep(2)
        res = self.client.container.terminate(C1)
        self.assertEqual(res, None)

        self.lg('List the containers, C1 {} should be gone'.format(C1))
        time.sleep(0.5)
        containers = self.client.container.list()
        self.assertFalse(str(C1) in containers)

        self.lg('Destroy C1 again, should fail')
        with self.assertRaises(RuntimeError):
            self.client.container.terminate(C1)

        self.lg('{} ENDED'.format(self._testID))

    def test003_deal_with_container_client(self):
        """ zos-011

        *Test case for testing dealing with container client*

        **Test Scenario:**

        #. Create a new container (C1), should succeed
        #. Get container(C1) client
        #. Use container client  to create  folder using system, should succeed
        #. Use container client to check folder is exist using bash
        #. Use zos client to check the folder is created only in container
        #. Use container client to delete created folder
        #. Destroy C1, should succeed

        """
        self.lg('{} STARTED'.format(self._testID))
        self.lg('Create a new container (C1), and make sure its exist')
        C1 = self.create_container(root_url=self.root_url, storage=self.storage)
        containers = self.client.container.list()
        self.assertTrue(str(C1) in containers)

        self.lg('Get container client(C1)')
        C1_client = self.client.container.client(C1)

        self.lg('Use container client to create  folder using system, should succeed')
        folder = self.rand_str()
        C1_client.system('mkdir {}'.format(folder))
        time.sleep(1.5)

        self.lg('Use container client to check folder is exist using bash')
        output = C1_client.bash('ls | grep {}'.format(folder))
        self.assertEqual(self.stdout(output), folder)

        self.lg('Check that the folder is created only in container')
        output2 = self.client.bash('ls | grep {}'.format(folder))
        self.assertEqual(self.stdout(output2), '')

        self.lg('Remove the created folder using bash,check that it removed ')
        C1_client.bash('rm -rf {}'.format(folder))
        time.sleep(1)
        output = self.client.bash('ls | grep {}'.format(folder))
        self.assertEqual(self.stdout(output), '')

        self.lg('Destroy C1, should succeed')
        self.client.container.terminate(C1)

        self.lg('{} ENDED'.format(self._testID))

    def vm_reachable(self, cmd, pub_port, timeout=60):
        while(timeout > 0):
            res = self.execute_command(cmd=cmd, ip=self.target_ip, port=pub_port)
            if res.stderr == '':
                return True
            time.sleep(1)
            timeout -= 1
        return False

    def test004_pause_resume_get_kvm(self):
        """ zos-050

        *Test case for testing pausing resuming VMs*

        **Test Scenario:**
        #. Create virtual machine (VM1), should succeed
        #. Pause VM1 and check state from get method, should be paused
        #. Make sure you can't reach VM1.
        #. Resume VM1 and check state from get method, should be resumed
        #. Make sure you can reach the VM1.
        #. Destroy VM1, should succeed
        """
        self.lg('{} STARTED'.format(self._testID))
        vm_name = self.rand_str()
        self.lg('- Create virtual machine {} , should succeed'.format(vm_name))
        pub_key = self.create_ssh_key()
        nics = [{'type': 'default'}]
        pub_port = randint(4000, 5000)
        config = {'/root/.ssh/authorized_keys': pub_key}
        vm_uuid = self.create_vm(name=vm_name, flist=self.ubuntu_flist,
                                 config=config, nics=nics, port={pub_port: 22})
        self.lg('Make sure VM1 is reachable')
        time.sleep(15)
        cmd = 'pwd'
        response = '/root\n'
        flag = self.vm_reachable(cmd, pub_port)
        self.assertTrue(flag, "vm is not reachable")

        self.lg('Pause VM1 and check state from get method ,should be paused')
        self.client.kvm.pause(vm_uuid)
        state_1 = self.client.kvm.get(vm_uuid)['state']
        self.assertEqual(state_1, 'paused')

        self.lg('Make sure you can\'t reach VM1')
        res = self.execute_command(cmd=cmd, ip=self.target_ip, port=pub_port)
        self.assertIn('No route to host', res.stderr)

        self.lg('Resume VM1 and check state from get method, should be resumed')
        self.client.kvm.resume(vm_uuid)
        state_2 = self.client.kvm.get(vm_uuid)['state']
        self.assertEqual(state_2, 'running')

        self.lg('Make sure you can reach the VM1')
        res = self.execute_command(cmd=cmd, ip=self.target_ip, port=pub_port)
        self.assertEqual(res.stdout, response)

        self.lg('- Destroy VM {}'.format(vm_name))
        self.client.kvm.destroy(vm_uuid)

    @unittest.skip('https://github.com/threefoldtech/0-core/issues/35')
    def test005_reset_reboot_shutdown_kvm(self):
        """ zos-053

        *Test case for testing reseting, rebooting and shutdown VMs*

        **Test Scenario:**
        #. Create virtual machine (VM1), should succeed
        #. Reset VM1 and make sure it is working after reseting.
        #. Reboot VM1 and make sure it is working after rebooting.
        #. shutdown VM1, and make sure you can't list it.
        """
        self.lg('{} STARTED'.format(self._testID))
        vm_name = self.rand_str()
        self.lg('- Create virtual machine {} , should succeed'.format(vm_name))
        pub_key = self.create_ssh_key()
        nics = [{'type': 'default'}]
        pub_port = randint(4000, 5000)
        config = {'/root/.ssh/authorized_keys': pub_key}
        vm_uuid = self.create_vm(name=vm_name, flist=self.ubuntu_flist,
                                 config=config, nics=nics, port={pub_port: 22})
        self.lg('Make sure vm is reachable')
        time.sleep(15)
        cmd = 'uptime'
        flag = self.vm_reachable(cmd, pub_port)
        self.assertTrue(flag, "vm is not reachable")

        self.lg('Reset VM1 and make sure it is working after reseting')
        self.client.kvm.reset(vm_uuid)
        # check that reset has been done
        res = self.execute_command(cmd=cmd, ip=self.target_ip, port=pub_port)
        self.assertEqual(int(res.stdout.split()[2]), 0)
        time.sleep(50)

        self.lg('Reboot VM1 make sure it is working after rebooting')
        self.client.kvm.reboot(vm_uuid)
        # check that reboot has been done
        res = self.execute_command(cmd=cmd, ip=self.target_ip, port=pub_port)
        self.assertEqual(int(res.stdout.split()[2]), 0)

        self.lg('shutdown VM1, and make sure you can\'t list it')
        self.client.kvm.shutdown(vm_uuid)
        time.sleep(10)
        vm = [vm for vm in self.client.kvm.list() if vm['uuid'] == vm_uuid]
        self.assertFalse(vm)
