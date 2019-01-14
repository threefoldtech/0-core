from utils.utils import BaseTest
from random import randint, choice
from zeroos.core0.client import Client
from nose_parameterized import parameterized
import unittest
import time

class AdvancedMachines(BaseTest):

    def setUp(self):
        super(AdvancedMachines, self).setUp()
        self.check_zos_connection(AdvancedMachines)
        self.zos_flist = 'https://hub.grid.tf/tf-autobuilder/zero-os-development.flist'
        self.vm_uuid = ''

    def tearDown(self):
        if self.vm_uuid:
            self.client.kvm.destroy(self.vm_uuid)
        super().tearDown()
        
    def ping_zos(self, client, timeout=60):
        now = time.time()
        while now + timeout > time.time():
            try:
                client.ping()
                return True
            except:
                continue
        return False

    def test001_create_kvm_with_params_memory_cpu(self):
        """ zos-054

        *Test case for creating kvm with parameters(memory, cpu)*

        **Test Scenario:**

        #. Create VM (VM1) with one of parameters(memory, cpu)
        #. Check that VM1 has been created with the specified parameters
        """
        self.lg('{} STARTED'.format(self._testID))

        self.lg('Create VM (VM1) with one of parameters(memory, cpu)')
        vm_name = self.rand_str()
        cpu = randint(1, 3)
        memory = randint(1, 3) * 1024
        self.vm_uuid = self.create_vm(name=vm_name, flist=self.ubuntu_flist, memory=memory, cpu=cpu)
        self.assertTrue(self.vm_uuid)

        self.lg('Check that VM1 has been created with the specified parameters')
        info = self.client.kvm.get(self.vm_uuid)['params']
        self.assertEqual(info['cpu'], cpu)
        self.assertEqual(info['memory'], memory)
        
        self.lg('{} ENDED'.format(self._testID))
    
    def test002_create_kvm_with_params_tags(self):
        """ zos-055

        *Test case for creating kvm with parameter tags*

        **Test Scenario:**

        #. Create VM (VM1) with parameter tags
        #. Search for (VM1) by its tags, should be found
        """
        self.lg('{} STARTED'.format(self._testID))

        self.lg('Create VM (VM1) with one of parameter tags')
        vm_name = self.rand_str()
        tags = self.rand_str()
        self.vm_uuid = self.create_vm(name=vm_name, flist=self.ubuntu_flist, tags=[tags])
        self.assertTrue(self.vm_uuid)

        self.lg('Search for (VM1) by its tags, should be found')
        vms = self.client.kvm.list()
        for vm in vms:
            vm_info = [vm for tag in vm['tags'] if tags[:5] in tag]
        self.assertTrue(vm_info)
        
        self.lg('{} ENDED'.format(self._testID))

    def test003_create_kvm_with_params_media(self):
        """ zos-056

        *Test case for creating kvm with parameter media*

        **Test Scenario:**
        
        #. Download ubuntu image 
        #. Create VM (VM1) with the ubuntu image has been downloaded
        #. Check that (VM1) has been created
        #. Delete the ubuntu image
        """
        self.lg('{} STARTED'.format(self._testID))

        self.lg('Download ubuntu image')
        image = 'ubuntu-18.04-minimal-cloudimg-amd64.img'
        img_loc = '/var/cache/{}'.format(self.rand_str())
        img_ftp = 'https://cloud-images.ubuntu.com/minimal/releases/bionic/release-20181122.1/ubuntu-18.04-minimal-cloudimg-amd64.img'
        media_path =  '{}/{}'.format(img_loc, image)
        
        state = self.client.bash('mkdir -p {}'.format(img_loc)).get().state
        self.assertEqual(state, 'SUCCESS')
        state = self.client.bash('wget {} -P {}'.format(img_ftp, img_loc)).get(150).state
        self.assertEqual(state, 'SUCCESS')

        self.lg('Create VM (VM1) with ubuntu image')
        vm_name = self.rand_str()
        media = [{'url': media_path}]
        self.vm_uuid = self.create_vm(name=vm_name, flist=None, media=media)
        self.assertTrue(self.vm_uuid)

        self.lg('Create VM (VM1) with the ubuntu image has been downloaded')
        self.assertTrue(self.vm_uuid)
        info = self.client.kvm.get(self.vm_uuid)['params']
        self.assertIn(info['media'][0]['url'], media_path)

        self.lg('Delete the ubuntu image')
        response = self.client.bash('rm -rf {}'.format(img_loc)).get()
        self.assertEqual(response.state, 'SUCCESS')

        self.lg('{} ENDED'.format(self._testID))

    @unittest.skip('https://github.com/threefoldtech/jumpscale_prefab/issues/31')
    def test004_create_kvm_with_params_mount(self):
        """ zos-057

        *Test case for creating kvm with parameter mount*

        **Test Scenario:**

        #. Create a directory (D1) on host and file (F1) inside it
        #. Create a vm (VM1) with mount directory (D1), should succeed
        #. Try to access VM1 and get F1, should be found
        #. Check that content of F1, should be the same
        #. Remove D1
        """
        self.lg('{} STARTED'.format(self._testID))

        self.lg('Create a directory (D1) on host and file (F1) inside it.')
        dir_path = '/{}'.format(self.rand_str())
        file_name = self.rand_str()
        data = self.rand_str()
        self.client.filesystem.mkdir(dir_path)
        self.client.bash('echo {} > {}/{}'.format(data, dir_path, file_name))

        self.lg('Create a vm (VM1) with mount directory (D1), should succeed.')
        vm_name = self.rand_str()
        pub_key = self.create_ssh_key()
        nics = [{'type': 'default'}]
        pub_port = randint(4000, 5000)
        config = {'/root/.ssh/authorized_keys': pub_key}
        vm_uuid = self.create_vm(name=vm_name, flist=self.ubuntu_flist, config=config, nics=nics,
                                 port={pub_port: 22}, mount=[{'source': dir_path, 'target': '/mnt'}])
        self.assertTrue(self.vm_uuid)

        self.lg('Try to access VM1 and get F1, should be found')
        flag = self.vm_reachable(self.target_ip, pub_port)
        self.assertTrue(flag, "vm is not reachable")
        cmd = 'cat /mnt/{}'.format(file_name)           
        resposne = self.execute_command(cmd=cmd, ip=self.target_ip, port=pub_port)

        self.lg('Check that content of F1, should be the same')
        content = resposne.stdout.strip()
        self.assertEqual(content, data)

        self.lg('Remove D1')
        response = self.client.bash('rm -rf {}'.format(dir_path)).get()
        self.assertEqual(response.state, 'SUCCESS')

        self.lg('{} ENDED'.format(self._testID))
    
    @unittest.skip('https://github.com/threefoldtech/jumpscale_prefab/issues/32')
    def test004_create_kvm_with_params_share_cache(self):
        """ zos-056

        *Test case for creating kvm with parameter share_cache*

        **Test Scenario:**

        #. Create a file (F1) inside /var/cache/zerofs
        #. Create a vm (VM1) with share_cache param, should succeed
        #. Try to access VM1 and get F1, should be found
        #. Check that content of F1, should be the same
        #. Remove F1
        """
        self.lg('{} STARTED'.format(self._testID))

        self.lg('Create a file (F1) inside /var/cache/zerofs')
        dir_path = '/var/cache/zerofs'
        file_name = self.rand_str()
        data = self.rand_str()
        self.client.bash('echo {} > {}/{}'.format(data, dir_path, file_name))
        
        self.lg('Create a vm (VM1) with share_cache param, should succeed')
        vm_name = self.rand_str()
        nics = [{'type': 'default'}]
        pub_port = randint(4000, 5000)
        vm_uuid = self.create_vm(name=vm_name, flist=self.zos_flist, nics=nics,
                                 port={pub_port: 6379}, share_cache=True)
        self.assertTrue(self.vm_uuid)

        self.lg('Try to access VM1 and get F1, should be found')
        client = Client(self.target_ip, port=pub_port)
        res = self.ping_zos(client, timeout=300)
        self.assertTrue(res, "Can't ping zos node")
        cmd = 'cat zoscahe/{}'.format(file_name)
        response = client.bash(cmd).get()

        self.lg('Check that content of F1, should be the same')
        content = response.stdout.strip()
        self.assertEqual(content, data)

        self.log('Remove F1')
        response = self.client.bash('rm -rf {}/{}'.format(dir_path, file_name)).get()
        self.assertEqual(response.state, 'SUCCESS')
        self.lg('{} ENDED'.format(self._testID))
