from utils.utils import BaseTest
import time
import unittest
from nose_parameterized import parameterized
import os
import io
from random import randint


class SystemTests(BaseTest):

    def setUp(self):
        super(SystemTests, self).setUp()
        self.check_zos_connection(SystemTests)

    def get_permission(self, client, path):
        return int(self.stdout(client.bash('stat -c %a {}'.format(path))))

    def container_create(self):
        self.cid = self.create_container(root_url=self.root_url, storage=self.storage)
        self.client_container = self.client.container.client(self.cid)

    def remove_container(self):
        self.client.container.terminate(self.cid)

    def getNicInfo(self, client):
        r = client.bash("ip -br a | awk '{print $1}'").get().stdout
        nics = [x.split()[0] for x in r.splitlines() if x.strip() != '']
        nicInfo = []
        for nic in nics:
            if '@' in nic:
                nic = nic[:nic.index('@')]
            addrs = client.bash('ip -br a show "{}"'.format(nic)).get().stdout.splitlines()[0].split()[2:]
            mtu = int(self.stdout(client.bash('cat /sys/class/net/{}/mtu'.format(nic))))
            hardwareaddr = self.stdout(client.bash('cat /sys/class/net/{}/address'.format(nic)))
            if hardwareaddr == '00:00:00:00:00:00':
                    hardwareaddr = ''
            tmp = {"name": nic, "hardwareaddr": hardwareaddr, "mtu": mtu, "addrs": [{"addr": x} for x in addrs]}
            nicInfo.append(tmp)

        return nicInfo

    def getCpuInfo(self, client):
        cpuInfo = {'vendorId': [], 'family': [], 'stepping': [], 'cpu': [], 'coreId': [], 'model': [],
                    'cacheSize': [], 'cores': [], 'flags': [], 'modelName': [], 'physicalId':[]}

        mapping = { "vendor_id": "vendorId", "cpu family": "family", "processor": "cpu", "core id": "coreId",
                    "cache size": "cacheSize", "cpu cores": "cores", "model name": "modelName", "physical id": "physicalId",
                    "stepping": "stepping", "flags": "flags", "model": "model"}

        keys = mapping.keys()
        for key in keys:
            lines = client.bash("cat /proc/cpuinfo | grep '{}' ".format(key)).get().stdout.splitlines()
            for line in lines:
                line = line.replace('\t', '')
                if key == line[:line.find(':')]:
                    item = line[line.index(':') + 1:].strip()
                    if key in ['processor', 'stepping', 'cpu cores']:
                        item = int(item)
                    if key == 'cache size':
                        item = int(item[:item.index(' KB')])
                    if key == 'flags':
                        item = item.split(' ')
                    cpuInfo[mapping[key]].append(item)
        return cpuInfo

    def getDiskInfo(self, client):
        response = client.bash('mount').get().stdout
        lines = response.splitlines()
        disks = []
        for line in lines:
            line = line.strip()
            if line == '':
                continue

            line = line.split()
            diskInfo = {'mountpoint': [], 'fstype': [], 'device': [], 'opts': []}
            diskInfo['mountpoint'] = line[2]
            diskInfo['fstype'] = line[4]
            diskInfo['device'] = line[0]
            diskInfo['opts'] = line[5][1:-1]
            disks.append(diskInfo)
        return disks

    def getMemInfo(self, client):

        lines = client.bash('cat /proc/meminfo').get().stdout.splitlines()
        memInfo = { 'active': 0, 'available': 0, 'buffers': 0, 'cached': 0,
                    'free': 0,'inactive': 0, 'total': 0}

        mapping = { 'Active': 'active', 'MemAvailable': 'available', 'Buffers':'buffers',
                    'Cached': 'cached', 'MemFree': 'free', 'Inactive':'inactive', 'MemTotal':'total'}

        keys = mapping.keys()
        for line in lines:
            line = line.replace('\t', '')
            for key in keys:
                if key == line[:line.find(':')]:
                    item = int(line[line.index(':') + 1:line.index(' kB')].strip())
                    item = item * 1024
                    memInfo[mapping[key]] = item

        return memInfo

    def get_dmi_bios_info(self, client):

        lines = client.bash('dmidecode -t bios').get().stdout.splitlines()
        dmi_info = {'Vendor':'vendor', 'Version':'version','Release Date':'release date','Address':'address',
                    'Runtime Size':'runtime size', 'ROM Size':'rom size','BIOS Revision':'bios revision' }

        for line in lines:
            line = line.replace('\t', '')
            for key in dmi_info.keys():
                if key == line[:line.find(':')]:
                    item = (line[line.index(':') + 2 :])
                    dmi_info[key] = item

        return dmi_info

    def get_port_info(self, client, protocol):

        if protocol == 'tcp':
            lines_1 = client.bash('netstat -tlpn | tail -n+3 | awk "{print \$4}"').get().stdout.splitlines() #ip:port
            lines_2 = client.bash('netstat -tlpn | tail -n+3 | awk "{print \$7}"').get().stdout.splitlines() #pid

        elif protocol == 'udp':
            lines_1 = client.bash('netstat -ulpn | tail -n+3 | awk "{print \$4}"').get().stdout.splitlines() #ip:port
            lines_2 = client.bash('netstat -ulpn | tail -n+3 | awk "{print \$6}"').get().stdout.splitlines() #pid

        else:
            return 0

        ip = []
        port = []
        pid = []
        # Separate ip, port from ip:port line
        for line in lines_1:
            if line == '' or line == '\n':
                continue
            else:
                line = line.replace('\t', '')
                ip.append(line[: line.rfind(':')])
                port.append(line[line.rfind(':') + 1 :])

        # Get pid
        for line in lines_2:
            if line == '' or line == '\n':
                continue
            else:
                line = line.replace('\t', '')
                if line[: line.rfind('/')] == '':
                    pid.append('0')
                else:
                    pid.append(line[: line.rfind('/')])

        port_info = {'ip' : ip, 'port': port, 'pid': pid}
        return port_info


    def test001_execute_commands(self):

        """ zos-001
        *Test case for testing basic commands using  bash and system*

        **Test Scenario:**
        #. Check if you can ping the remote host, should succeed
        #. Create folder using system
        #. Check that the folder is created using bash (C1)
        #. Check that you can get same responce for (C1)
        #. Remove the created folder
        """

        self.lg('{} STARTED'.format(self._testID))

        self.lg('Check if you can ping the remote host, should succeed')
        rs = self.client.ping()
        self.assertEqual(rs[:4], 'PONG')

        self.lg('Create folder using system')
        folder = self.rand_str()
        self.client.system('mkdir {}'.format(folder))

        self.lg('Check that the folder is created')
        rs1 = self.client.bash('ls | grep {}'.format(folder))
        rs_ob = rs1.get()
        self.assertEqual(rs_ob.stdout.replace('\n', ''), '{}'.format(folder))
        self.assertEqual(rs_ob.state, 'SUCCESS')

        self.lg('Check that you can get same responce for (C1)')
        rs11 = self.client.response_for(rs1.id)
        self.assertEqual(self.stdout(rs11), self.stdout(rs1))
        self.assertEqual(rs1.id, rs11.id)

        self.lg('Remove the created folder')
        self.client.bash('rm -rf {}'.format(folder))
        time.sleep(0.5)
        rs2 = self.client.bash('ls | grep {}'.format(folder))
        self.assertEqual(self.stdout(rs2), '')

        self.lg('{} ENDED'.format(self._testID))

    def test002_kill_list_processes(self):

        """ zos-002
        *Test case for testing killing and listing processes*

        **Test Scenario:**
        #. Create process that runs for long time using both system and bash
        #. List the process, should be found
        #. Kill the process
        #. List the process, shouldn't be found
        """

        self.lg('{} STARTED'.format(self._testID))

        cmd = 'sleep 40'
        self.client.bash(cmd)
        self.lg('Created process that runs for long time using {}'.format(cmd))

        self.lg('List the process, should be found')
        id = self.get_process_id(cmd)
        self.assertIsNotNone(id)

        self.lg('Kill the process')
        self.client.process.kill(id)

        self.lg('List the process, shouldn\'t be found')
        id = self.get_process_id(cmd)
        self.assertIsNone(id)

        self.lg('{} ENDED'.format(self._testID))

    @parameterized.expand(['client', 'container'])
    def test003_os_info(self, client_type):

        """ zos-003
        *Test case for checking on the system os information*

        **Test Scenario:**
        #. Get the os information using zos/container client
        #. Get the hostname and compare it with the zos/container os insformation
        #. Get the kernal's name and compare it with the zos/container os insformation
        """

        self.lg('{} STARTED'.format(self._testID))

        if client_type == 'client':
            client = self.client
        else:
            self.container_create()
            client = self.client_container

        self.lg('Get the os information using zos/container client')
        os_info = client.info.os()

        self.lg('Get the hostname and compare it with the zos/container os insformation')
        hostname = client.system('uname -n').get().stdout.strip()
        self.assertEqual(os_info['hostname'], hostname)

        self.lg('Get the kernal\'s name and compare it with the zos/container os insformation')
        krn_name = client.system('uname -s').get().stdout.strip()
        self.assertEqual(os_info['os'], krn_name.lower())

        self.lg('{} ENDED'.format(self._testID))

    @parameterized.expand(['client', 'container'])
    def test004_mem_info(self, client_type):

        """ zos-004
        *Test case for checking on the system memory information*

        **Test Scenario:**
        #. Get the memory information using zos/container client
        #. Get the memory information using bash
        #. Compare memory zos/container  results to that of the bash results, should be the same
        """
        self.lg('{} STARTED'.format(self._testID))

        if client_type == 'client':
            client = self.client
        else:
            self.container_create()
            client = self.client_container

        self.lg('get memory info using bash')
        expected_mem_info = self.getMemInfo(client)

        self.lg('get memory info using zos/container ')
        zos_mem_info = client.info.mem()

        self.lg('compare zos/container  results to bash results')
        self.assertEqual(expected_mem_info['total'], zos_mem_info['total'])
        params_to_check = ['active', 'available', 'buffers', 'cached', 'free', 'inactive']
        for key in params_to_check:
            threshold = 1024 * 10000  # acceptable threshold (10 MB)
            zos_value = zos_mem_info[key]
            expected_value = expected_mem_info[key]
            self.assertTrue(expected_value - threshold <= zos_value <= expected_value + threshold, key)

        self.lg('{} ENDED'.format(self._testID))

    @parameterized.expand(['client', 'container'])
    def test005_cpu_info(self, client_type):

        """ zos-005
        *Test case for checking on the system CPU information*

        **Test Scenario:**
        #. Get the CPU information using zos/container client
        #. Get the CPU information using bash
        #. Compare CPU zos/container  results to that of the bash results, should be the same
        """
        self.lg('{} STARTED'.format(self._testID))

        if client_type == 'client':
            client = self.client
        else:
            self.container_create()
            client = self.client_container

        self.lg('get cpu info using bash')
        expected_cpu_info = self.getCpuInfo(client)

        self.lg('get cpu info using zos')
        zos_cpu_info = client.info.cpu()

        self.lg('compare zos/container results to bash results')
        for key in expected_cpu_info.keys():
            if key == 'cores':
                continue
            zos_param_list = [x[key] for x in zos_cpu_info]
            self.assertEqual(expected_cpu_info[key], zos_param_list)

        self.lg('{} ENDED'.format(self._testID))

    @parameterized.expand(['client', 'container'])
    def test006_disk_info(self, client_type):

        """ zos-006
        *Test case for checking on the disks information*

        **Test Scenario:**
        #. Get the disks information using zos client
        #. Get the disks information using bas
        #. Compare disks zos results to that of the bash results, should be the same
        """
        self.lg('{} STARTED'.format(self._testID))

        if client_type == 'client':
            client = self.client
        else:
            self.container_create()
            client = self.client_container

        self.lg('get disks info using linux bash command (mount)')
        expected_disk_info = self.getDiskInfo(client)

        self.lg('get cpu info using zos')
        zos_disk_info = client.info.disk()

        self.lg('compare zos results to bash results')
        for disk in zos_disk_info:
            self.assertIn(disk, expected_disk_info)

        self.lg('{} ENDED'.format(self._testID))

    def test007_nic_info(self):

        """ zos-007
        *Test case for checking on the system nic information*

        **Test Scenario:**
        #. Get the nic information using zos client
        #. Get the information using bash
        #. Compare nic zos results to that of the bash results, should be the same
        """
        self.lg('{} STARTED'.format(self._testID))

        self.lg('get nic info using linux bash command (ip a)')
        expected_nic_info = self.getNicInfo(self.client)

        self.lg('get nic info using zos client')
        zos_nic_info = self.client.info.nic()

        self.lg('compare zos/container results to bash results')
        params_to_check = ['name', 'addrs', 'mtu', 'hardwareaddr']
        for i in range(len(expected_nic_info) - 1):
            for param in params_to_check:
                self.assertEqual(expected_nic_info[i][param], zos_nic_info[i][param])

        self.lg('{} ENDED'.format(self._testID))

    @parameterized.expand(['client', 'container'])
    def test008_mkdir_exists_list_chmod_move_remove_directory(self, client_type):
        """ zos-015
        *Test case for test filesystem mkdir, exists, list, chmod, move, remove methods*

        **Test Scenario:**
        #. Make new directory (D1), should succeed
        #. Check directory (D1) exists, should succeed
        #. Move directory (D1), should succeed
        #. Remove directory (D1), should succeed
        #. Make new parent directory (D2), should succeed
        #. Make new directory (D3) inside (D2), should succeed
        #. Change directory (D2) mode, should succeed
        """
        if client_type == 'client':
            client = self.client
        else:
            self.container_create()
            client = self.client_container

        self.lg('{} STARTED'.format(self._testID))

        self.lg('Make new directory (D1), should succeed')
        dir_name = self.rand_str()
        client.filesystem.mkdir(dir_name)

        self.lg('Check directory (D1) is exists, should succeed')
        ## using bash
        ls = client.bash('ls').get().stdout.splitlines()
        self.assertIn(dir_name, ls)
        ## using bash filesystem.list
        ls = [x['name'] for x in client.filesystem.list('.')]
        self.assertIn(dir_name, ls)
        ## using bash filesystem.exists
        self.assertTrue(client.filesystem.exists(dir_name))

        self.lg('Move directory (D1), should succeed')
        new_destination = '/root/{}'.format(dir_name)
        client.filesystem.move(dir_name, new_destination)
        self.assertTrue(client.filesystem.exists(new_destination))
        self.assertFalse(client.filesystem.exists(dir_name))

        self.lg('Remove directory (D1), should succeed')
        client.filesystem.remove(new_destination)
        ls = client.bash('ls').get().stdout.splitlines()
        self.assertNotIn(dir_name, ls)

        self.lg('Make new parent directory (D2), should succeed')
        parent_dir = self.rand_str()
        client.filesystem.mkdir(parent_dir)

        self.lg('Make new directory (D3) inside (D2), should succeed')
        child_dir = '{}/{}'.format(parent_dir, self.rand_str())
        client.filesystem.mkdir(child_dir)

        self.lg('Change directory (D2) mode, should succeed')
        client.filesystem.chmod(parent_dir, 0o777)
        parent_dir_perimission = self.get_permission(client, parent_dir)
        child_dir_perimission = self.get_permission(client, child_dir)
        self.assertEqual(parent_dir_perimission, 777)
        self.assertNotEqual(child_dir_perimission, 777)

        self.lg('Change directory (D2) mode recursively, should succeed')
        client.filesystem.chmod(parent_dir, 0o321, recursive=True)
        parent_dir_perimission = self.get_permission(client, parent_dir)
        child_dir_perimission = self.get_permission(client, child_dir)
        self.assertEqual(parent_dir_perimission, 321)
        self.assertEqual(child_dir_perimission, 321)

        if client_type == 'container':
            self.remove_container()

        self.lg('{} ENDED'.format(self._testID))

    @parameterized.expand(['client', 'container'])
    def test009_open_close_read_write_file(self, client_type):

        """ zos-019
        *Test case for test filesystem upload, download, upload_file, download_file methods*

        **Test Scenario:**
        #. Open file (F1) in read (r) mode, should succeed
        #. Read file (F1) and check its content, should succeed
        #. Try to write to the file (F1), should fail
        #. Open file (F1) in (w+) mode
        #. Write text to file (F1), should succeed
        #. Check file (F1) is truncated and contains only (T2) text
        #. Try to read the file (F1), should fail
        #. Open file (F1) in (w+) mode
        #. Write text to file (F1), should succeed
        #. Read text from file (F1), should succeed
        #. Open file (F1) in (r+) mode
        #. Write text to file (F1), should success
        #. Check file (F1) content, should success
        #. Open file (F1) in (a) mode
        #. Write text to file (F1), should succeed
        #. Check file (F1) text , should succeed
        #. Create file (F2) using open in (x) mode, should succeed
        #. Check if  file (F2) exists, should succeed
        """

        if client_type == 'client':
            client = self.client
        else:
            self.container_create()
            client = self.client_container

        self.lg('{} STARTED'.format(self._testID))

        file_name = '{}.txt'.format(self.rand_str())
        client.bash('touch /{}'.format(file_name))
        time.sleep(1)

        open_modes = ['r', 'w', 'a', 'w+', 'r+', 'a+', 'x']
        for mode in open_modes:
            if mode != 'x':
                txt = 'line1\nline2\nline3'
                client.bash('echo "{}" > {}'.format(txt, file_name))
                time.sleep(1)
                f = client.filesystem.open(file_name, mode=mode)

            if mode == 'r':

                self.lg('Open file (F1) in read (r) mode, should succeed')

                self.lg('Read file (F1) and check its content, should succeed')
                file_text = client.filesystem.read(f).decode('utf-8')
                self.assertEqual(file_text, '{}\n'.format(txt))

                self.lg('Try to write to the file (F1), should fail')
                txt = new_txt = str.encode(self.rand_str())
                with self.assertRaises(RuntimeError):
                    client.filesystem.write(f, txt)
                with self.assertRaises(RuntimeError):
                    client.filesystem.open(self.rand_str(), mode=mode)

            if mode == 'w':

                self.lg('Open file (F1) in write only (w) mode')

                self.lg('Write text (T2) to file')
                new_txt = str.encode(self.rand_str())
                client.filesystem.write(f, new_txt)

                self.lg('Check file (F1) is truncated and contains only (T2) text')
                file_text = client.bash('cat {}'.format(file_name)).get().stdout
                self.assertEqual(file_text, '{}'.format(new_txt.decode('utf-8')), mode)

                self.lg('Try to read the file (F1), should fail')
                with self.assertRaises(RuntimeError):
                    client.filesystem.read(f)

            if mode == 'w+':

                self.lg('Open file (F1) in (w+) mode')

                self.lg('Write text to file (F1), should succeed')
                new_txt = str.encode(self.rand_str())
                client.filesystem.write(f, new_txt)

                self.lg('Read text from file (F1), should succeed')
                client.filesystem.read(f)

            # read/write(at begin)
            if mode == 'r+':

                self.lg('Open file (F1) in (r+) mode')

                self.lg('Write text to file (F1), should success')
                new_txt = str.encode(self.rand_str())
                l = len(new_txt)
                client.filesystem.write(f, new_txt)

                self.lg('Check file (F1) content, should success')
                file_text = client.bash('cat {}'.format(file_name)).get().stdout
                self.assertEqual(file_text.replace('\n', ''), '{}{}'.format(new_txt.decode('utf-8'), txt[l:]).replace('\n', ''))
                file_text = client.filesystem.read(f).decode('utf-8')
                self.assertEqual(file_text, '{}\n'.format(txt[l:]))
                with self.assertRaises(RuntimeError):
                    client.filesystem.open(self.rand_str(), mode=mode)

            if mode == 'a':

                self.lg('Open file (F1) in (a) mode')

                self.lg('Write text to file (F1), should succeed')
                new_txt = str.encode(self.rand_str())
                client.filesystem.write(f, new_txt)
                file_text = client.bash('cat {}'.format(file_name)).get().stdout

                self.lg('Check file (F1) text , should succeed')
                final_text = '{}{}'.format(txt, new_txt.decode('utf-8'))
                self.assertEqual(file_text.replace('\n', ''), final_text.replace('\n', ''))

            if mode == 'x':
                self.lg('Create file (F2) using open in (x) mode, should succeed')
                file_name_2 = '{}.txt'.format(self.rand_str())
                client.filesystem.open(file_name_2, mode=mode)

                self.lg('Check if  file (F2) exists, should succeed')
                ls = client.bash('ls').get().stdout.splitlines()
                self.assertIn(file_name_2, ls)

            client.filesystem.close(f)

        if client_type == 'container':
            self.remove_container()

        self.lg('{} ENDED'.format(self._testID))

    @parameterized.expand(['client', 'container'])
    def test010_upload_download_file(self, client_type):

        """ zos-018
        *Test case for test filesystem upload, download, upload_file, download_file methods*

        **Test Scenario:**
        #. Create local file (LF1) and write data to it, should succeed
        #. Upload file (LF1) to zos/container
        #. Check file (LF1) exists in zos/container and check its data
        #. Upload buffer data to the remote file
        #. Check the remote file content equal to buffer content
        #. Create remote file (RF1) and write data to it, should succeed
        #. Download file (RF1) to localhost
        #. Check file (RF1) exists in localhost and check its data
        #. Download file (RF1) to buffer
        #. Check buffer data equal to file (RF1) content
        """

        if client_type == 'client':
            client = self.client
        else:
            self.container_create()
            client = self.client_container

        self.lg('{} STARTED'.format(self._testID))

        self.lg('Create local file (LF1) and write data to it, should succeed')
        local_file_name = '{}.txt'.format(self.rand_str())
        remote_file_name = '{}.txt'.format(self.rand_str())

        test_txt = self.rand_str()
        with open(local_file_name, 'w+') as f:
            f.write(test_txt)

        self.lg('Upload file (LF1) to zos/container')
        client.filesystem.upload_file('/{}'.format(remote_file_name), local_file_name)

        self.lg('Check file (LF1) is exists in zos/container and check its data')
        ls = client.bash('ls').get().stdout.splitlines()
        self.assertIn(remote_file_name, ls)
        file_text = client.bash('cat {}'.format(remote_file_name)).get().stdout.strip()
        self.assertEqual(file_text, test_txt)

        self.lg('Upload buffer data to the remote file')
        client.bash('echo "" > {}'.format(remote_file_name))
        time.sleep(1)
        buff = io.BytesIO(bytes(self.rand_str().encode('utf-8')))
        time.sleep(1)
        client.filesystem.upload('/{}'.format(remote_file_name), buff)
        time.sleep(1)
        self.lg('Check the remote file content equal to buffer content')
        file_text = client.bash('cat {}'.format(remote_file_name)).get().stdout.strip()
        time.sleep(1)
        self.assertEqual(buff.getvalue().decode('utf-8'), file_text)

        self.lg('Create remote file (RF1) and write data to it, should succeed')
        remote_file_name = '{}.txt'.format(self.rand_str())
        local_file_name = '{}.txt'.format(self.rand_str())
        test_txt = self.rand_str()
        client.bash('echo "{}" > {}'.format(test_txt, remote_file_name))
        time.sleep(1)

        self.lg('Download file (RF1) to localhost')
        client.filesystem.download_file('/{}'.format(remote_file_name), local_file_name)
        time.sleep(1)
        self.lg('Check file (RF1) is exists in localhost and check its data')
        ls = os.listdir()
        self.assertIn(local_file_name, ls)
        with open(local_file_name, 'r') as f:
            self.assertEqual(f.read().strip(), test_txt)

        self.lg('Download file (RF1) to buffer')
        buff = io.BytesIO()
        client.filesystem.download('/{}'.format(remote_file_name), buff)

        self.lg('Check buffer data equal to file (RF1) content')
        self.assertEqual(buff.getvalue().decode('utf-8').strip(), test_txt)

        if client_type == 'container':
            self.remove_container()

        self.lg('{} ENDED'.format(self._testID))

    def test011_kill_list_jobs(self):

        """ zos-032
        *Test case for testing killing and listing jobs*

        **Test Scenario:**
        #. Create job that runs for long time
        #. List the job, should be found
        #. Kill the job
        #. List the job, shouldn't be found
        """

        self.lg('{} STARTED'.format(self._testID))

        self.lg('Create job that runs for long time')
        cmd = 'core.system'
        match = 'sleep'
        self.client.system('sleep 40')

        self.lg('List the job, should be found')
        id = self.get_job_id(cmd, match)
        self.assertIsNotNone(id)

        self.lg('Kill the job')
        self.client.job.kill(id)

        self.lg('List the job, shouldn\'t be found')
        id = self.get_job_id(cmd, match)
        self.assertIsNone(id)

        self.lg('{} ENDED'.format(self._testID))

    def test012_add_delete_list_addr(self):

        """ zos-038
        *Test case for testing adding, deleteing and listing ip addresses*

        **Test Scenario:**
        #. Create a bridge and get its interface/Link L1
        #. Add ip for non existing link, should fail
        #. Add invalid cidr address to L1, should fail
        #. Add ip1 for link L1, should succeed
        #. List ips of that link, ip1 should be there
        #. Delete the added ip (ip1) of link L1, should succeed
        #. Delete ip1 again, should fail
        """

        self.lg('{} STARTED'.format(self._testID))

        self.lg('Create a bridge and get its interface/Link L1')
        L1 = self.rand_str()
        self.client.bridge.create(L1)

        self.lg('Add ip for non existing link, should fail')
        cidr = '192.168.34.1/16'
        with self.assertRaises(RuntimeError):
            self.client.ip.addr.add(self.rand_str(), cidr)

        self.lg('Add invalid cidr address to L1, should fail')
        with self.assertRaises(RuntimeError):
            self.client.ip.addr.add(self.rand_str(), '192.168.34.1')

        self.lg('Add ip1 for link L1, should succeed')
        self.client.ip.addr.add(L1, cidr)

        self.lg('Add ip1 for link L1, should fail')
        with self.assertRaises(RuntimeError):
            self.client.ip.addr.add(L1, cidr)

        self.lg('List ips of that link, ip1 should be there')
        self.assertEqual(self.client.ip.addr.list(L1)[0], cidr)

        self.lg('Delete the added ip (ip1) of link L1, should succeed')
        self.client.ip.addr.delete(L1, cidr)
        self.assertNotEqual(self.client.ip.addr.list(L1)[0], cidr)

        self.lg('Delete ip1 again, should fail')
        with self.assertRaises(RuntimeError):
            self.client.ip.addr.delete(L1, cidr)

        self.lg('{} ENDED'.format(self._testID))

    def test013_link_up_down_list_rename(self):

        """ zos-039
        *Test case for testing adding, deleteing and listing ip addresses*

        **Test Scenario:**
        #. Create a bridge and get its interface/Link L1
        #. Set the interface L1 down, should succeed
        #. Set the interface L1 up , should succeed
        #. Rename the interface L1 while it is up, should fail
        #. Set the interface L1 down then rename it, should succeed
        """

        self.lg('{} STARTED'.format(self._testID))

        self.lg('Create a bridge and get its interface/Link L1')
        L1 = self.rand_str()
        self.client.bridge.create(L1)

        self.lg('Set the interface L1 down, should succeed')
        self.client.ip.link.down(L1)
        self.client.ip.link.list()[-1]['up']

        self.lg('Set the interface L1 up , should succeed')
        self.client.ip.link.up(L1)
        self.client.ip.link.list()[-1]['up']

        self.lg('Rename the interface L1 while it is up, should fail')
        new_L1 = self.rand_str()
        with self.assertRaises(RuntimeError):
            self.client.ip.link.name(L1, new_L1)

        self.lg('Set the interface L1 down then rename it, should succeed')
        self.client.ip.link.down(L1)
        self.client.ip.link.name(L1, new_L1)
        self.assertEqual(self.client.ip.link.list()[-1]['name'], new_L1)

        self.lg('{} ENDED'.format(self._testID))

    def test014_add_delete_interface_bridge(self):

        """ zos-040
        *Test case for testing adding, deleteing bridges and their interfaces*

        **Test Scenario:**
        #. Add a bridge B1, should succeed
        #. Add interface to B1, should succeed
        #. Delete the added interface , should succeed
        #. Delete added interface again, should fail
        #. Delete a fake interface, should fail
        #. Delete the bridge, should succeed
        #. Delete the bridge again, should fail
        """

        self.lg('{} STARTED'.format(self._testID))

        self.lg('Add a bridge B1, should succeed')
        br = self.rand_str()
        self.client.ip.bridge.add(br)
        self.assertEqual(self.client.bridge.list()[-1], br)

        self.lg('Add interface to B1, should succeed')
        inf = self.rand_str()
        self.client.bash('ip l a {} type dummy'.format(inf)).get()
        self.client.ip.bridge.addif(br, inf)
        out = self.client.bash('brctl show | grep {} | grep -F -o {}'.format(br, inf))
        self.assertEqual(self.stdout(out), inf)

        self.lg('Delete the added interface , should succeed')
        self.client.ip.bridge.delif(br, inf)
        out = self.client.bash('brctl show | grep {} | grep -F -o {}'.format(br, inf))
        self.assertEqual(self.stdout(out), '')

        self.lg('Delete added interface again, should fail')
        with self.assertRaises(RuntimeError):
            self.client.ip.bridge.delif(br, inf)

        self.lg('Delete a fake interface, should fail')
        with self.assertRaises(RuntimeError):
            self.client.ip.bridge.delif(br, self.rand_str())

        self.lg('Delete the bridge, should succeed')
        self.client.ip.bridge.delete(br)
        self.assertNotIn(br, self.client.bridge.list())

        self.lg('Delete the bridge again, should fail')
        with self.assertRaises(RuntimeError):
            self.client.ip.bridge.delete(br)

        self.lg('{} ENDED'.format(self._testID))

    def test015_add_delete_list_route(self):

        """ zos-041
        *Test case for testing adding, deleteing and listing routes*

        **Test Scenario:**
        #. list all routes then get the etho route R1, should succeed
        #. Delete route R1, should succeed
        #. Delete route R1 again, should fail
        #. Delete route for non existing link, should fail
        #. Add route R1, should succeed
        #. Add route R1 again , should fail
        #. Add route for non existing link, should fail
        """

        self.lg('{} STARTED'.format(self._testID))

        self.lg('list all routes then get the etho route, should succeed')
        l = self.client.ip.route.list()
        new_l = [i for i in l if i['dst'] != '' and ':' not in i['dst'] and i['dev'].startswith('e')]

        self.lg('Delete route R1, should succeed')
        self.client.ip.route.delete(new_l[0]['dev'], new_l[0]['dst'])

        self.lg('Delete route R1 again, should fail')
        with self.assertRaises(RuntimeError):
            self.client.ip.route.delete(new_l[0]['dev'], new_l[0]['dst'])

        self.lg('Delete route for non existing link, should fail')
        with self.assertRaises(RuntimeError):
            self.client.ip.route.delete(self.rand_str(), new_l[0]['dst'])

        self.lg('Add route R1, should succeed')
        self.client.ip.route.add(new_l[0]['dev'], new_l[0]['dst'])

        self.lg('Add route R1 again, should fail')
        with self.assertRaises(RuntimeError):
            self.client.ip.route.add(new_l[0]['dev'], new_l[0]['dst'])

        self.lg('Add route for non existing link, should fail')
        with self.assertRaises(RuntimeError):
            self.client.ip.route.add(self.rand_str(), new_l[0]['dst'])

        self.lg('{} ENDED'.format(self._testID))

    def test016_open_drop_list_nft(self):

        """ zos-042
        *Test case for testing opening, droping ports and listing rules*

        **Test Scenario:**
        #. Open ssh port, should succeed
        #. List the ssh port and check if the rule exist, should succeed
        #. Open ssh port again should fail
        #. Drop the ssh port, should succeed
        #. Open fake port which is out of range, should fail
        #. List the ports and make sure the fake port is not there, should succeed
        """

        self.lg('{} STARTED'.format(self._testID))

        self.lg('Open ssh port, should succeed')
        self.client.nft.open_port(22)
        out = self.client.bash('nft list ruleset -a | grep -F -o "ssh accept"')
        self.assertEqual(self.stdout(out), 'ssh accept')

        self.lg('List the ssh port and check if the rule exist, should succeed')
        self.assertIn('tcp dport 22 accept', self.client.nft.list())
        self.assertEqual(self.client.nft.rule_exists(22), True)

        self.lg('Open ssh port again should fail')
        with self.assertRaises(RuntimeError):
            self.client.nft.open_port(22)

        self.lg('Drop the ssh port, should succeed')
        self.client.nft.drop_port(22)
        out = self.client.bash('nft list ruleset -a | grep -F -o "ssh accept"')
        self.assertEqual(self.stdout(out), '')

        self.lg('Open fake port which is out of range, should fail')
        port = randint(666666, 777777)
        with self.assertRaises(RuntimeError):
            self.client.nft.open_port(port)

        self.lg('List the ports and make sure the fake port is not there, should succeed')
        self.assertNotIn('tcp dport {} accept'.format(port), self.client.nft.list())
        with self.assertRaises(RuntimeError):
            self.client.nft.rule_exists(port)

        self.lg('{} ENDED'.format(self._testID))

    def test017_add_list_remove_tasks_cgroups(self):

        """ zos-046
        *Test case for adding, listing and removing tasks for cgroups*

        **Test Scenario:**
        #. Create a cgroup (CG1), should succeed.
        #. list tasks, should be empty.
        #. Create a process (P1).
        #. Add P1 as a task (T1) for cgroup (CG1).
        #. List tasks, task1 should be found.
        #. Remove task (T1), should succeed.
        """

        self.lg('{} STARTED'.format(self._testID))

        self.lg('Create a cgroup, should succeed')
        cg_name = '100m'
        subsystem = 'memory'
        self.client.cgroup.ensure(subsystem, cg_name)

        self.lg('list tasks, should be empty')
        res = self.client.cgroup.tasks(subsystem, cg_name)
        self.assertFalse(res)

        self.lg('Create a process P1')
        command = 'sleep 200'
        self.client.bash(command)
        pid = self.get_process_id(command)

        self.lg('Add P1 as a task (T1) for cgroup (CG1)')
        self.client.cgroup.task_add(subsystem, cg_name, pid)

        self.lg('List tasks, T1 should be found')
        res = self.client.cgroup.tasks(subsystem, cg_name)
        self.assertEqual(res[0], pid)

        self.lg('Remove task (T1), should succeed.')
        self.client.cgroup.task_remove(subsystem, cg_name, pid)
        res = self.client.cgroup.tasks(subsystem, cg_name)
        self.assertFalse(res)
        self.lg('Remove cgroup')
        self.client.cgroup.remove(subsystem, cg_name)

        self.lg('{} ENDED'.format(self._testID))

    def test018_add_list_remove_cgroup_subsystem(self):

        """ zos-047
        *Test case for adding, listing and removing subsystem*

        **Test Scenario:**
        #. list cgroups, should be empty
        #. Create two cgroups, should succeed
        #. List cgroups, should find two.
        #. Remove 1st cgroup, should succeed.
        #. Set the memory limit for the 2nd cgroup to L1, should succeed.
        #. Reset the 2nd cgroup, should succeed.
        """

        self.lg('{} STARTED'.format(self._testID))

        self.lg('list cgroups, should be empty')
        cg_list = self.client.cgroup.list()
        self.assertFalse(cg_list)

        self.lg('Create two cgroups, should succeed')
        cg1_name = '100m'
        cg2_name = '200m'
        subsystem = 'memory'
        self.client.cgroup.ensure(subsystem, cg1_name)
        self.client.cgroup.ensure(subsystem, cg2_name)

        self.lg('List cgroups, should find two.')
        cg_list = self.client.cgroup.list()[subsystem]
        self.assertEqual(len(cg_list), 2)
        self.assertIn(cg1_name, cg_list)
        self.assertIn(cg2_name, cg_list)

        self.lg('Remove 1st cgroup, should succeed')
        self.client.cgroup.remove(subsystem, cg1_name)
        cg_list = self.client.cgroup.list()[subsystem]
        self.assertNotIn(cg1_name, cg_list)
        self.assertEqual(len(cg_list), 1)

        self.lg('Set the memory limit for the 2nd cgroup to L1, should succeed')
        L1 = 200 * 1024 * 1024
        res = self.client.cgroup.memory(cg2_name, L1)
        mem = self.stdout(self.client.bash('cat /sys/fs/cgroup/{}/{}/memory.limit_in_bytes'.format(subsystem, cg2_name)))
        self.assertEqual(res['mem'], int(mem))

        self.lg('Reset the 2nd cgroup, should succeed')
        self.client.cgroup.reset(subsystem, cg2_name)
        res = self.client.cgroup.memory(cg2_name)
        self.assertGreater(res['mem'], L1 * 1024)
        self.assertEqual(res['swap'], 0)

        self.lg('{} ENDED'.format(self._testID))

    def test019_dmi_bios_info(self):
        """ zos-048
        *Test case for checking on the dmi BIOS information*

        **Test Scenario:**
        #. Get the dmi bios information using zos client
        #. Get the information using bash
        #. Compare dmi zos results to that of the bash results, should be the same
        """
        self.lg('get dmi_bios info using linux bash command (dmidecode -t bios)')
        dmi_bios_info = self.get_dmi_bios_info(self.client)

        self.lg('get dmi_bios info using zos client')
        bios_info = self.client.info.dmi('bios')['BIOS Information']['properties']

        for k in dmi_bios_info:
            self.assertEqual(dmi_bios_info[k],bios_info[k]['value'])


    @parameterized.expand(['client', 'container'])
    def test020_mkdir_chown_remove_directory(self, client_type):
        """ zos-049
        *Test case for testing filesystem mkdir and chown methods*

        **Test Scenario:**
        #. Make new directory (D1), should succeed
        #. Check directory (D1) exists, should be there
        #. Get user owner and group owner using bash
        #. Check directory (D1) user owner and group owner belong to (root) , should succeed
        #. Add new user (U1) using bash and make sure that it's created
        #. Change directory (D1) user owner and group owner to the new user (U1) using filesystem chown method
        #. Get user owner and group owner using bash
        #. Check if directory (D1) user owner and group owner belong to the new user (U1), should succeed
        #. Remove user (U1), should succeed
        #. Remove directory (D1), should succeed
        """
        if client_type == 'client':
            client = self.client
            adduser_options = '-D'
        else:
            self.container_create()
            client = self.client_container
            adduser_options = '--disabled-password --force-badname'

        self.lg('{} STARTED'.format(self._testID))

        self.lg('Make new directory (D1), should succeed')
        dir_name = self.rand_str()
        client.filesystem.mkdir(dir_name)

        self.lg('Check if directory (D1) is exists, should be there')
        ## using bash filesystem.exists
        self.assertTrue(client.filesystem.exists(dir_name))

        self.lg('Add new user (U1) using bash and make sure that it is created')
        new_user = self.rand_str()
        state = client.bash('adduser {} {} '.format(adduser_options,new_user)).get().state
        self.assertEqual(state, 'SUCCESS')

        self.lg('Get user owner and group owner using bash')
        user_owner_1 = client.bash('stat -c %U {}'.format(dir_name)).get().stdout.splitlines()[0]
        group_owner_1 = client.bash('stat -c %G {}' .format(dir_name)).get().stdout.splitlines()[0]

        self.lg('Check directory (D1) user owner and group owner belong to (root), should succeed')
        self.assertEqual(user_owner_1, 'root')
        self.assertEqual(group_owner_1, 'root')

        self.lg('Change directory (D1) user owner and group owner to the new user (U1) using filesystem chown method')
        client.filesystem.chown(dir_name, new_user, new_user)

        self.lg('Get user owner and group owner using bash')
        user_owner_2 = client.bash('stat -c %U {}'.format(dir_name)).get().stdout.splitlines()[0]
        group_owner_2 = client.bash('stat -c %G {}' .format(dir_name)).get().stdout.splitlines()[0]

        self.lg('Check directory (D1) user owner and group owner belong to the new user (U1), should succeed')
        self.assertEqual(user_owner_2, new_user)
        self.assertEqual(group_owner_2, new_user)

        self.lg('Remove directory (D1), should succeed')
        client.filesystem.remove(dir_name)
        ls = client.bash('ls').get().stdout.splitlines()
        self.assertNotIn(dir_name, ls)

        self.lg('Remove user (U1), should succeed')
        state = client.bash('deluser {}'.format(new_user)).get().state
        self.assertEqual(state, 'SUCCESS')

    @parameterized.expand(['tcp', 'udp'])
    def test021_port_info(self, protocol):
        """ zos-051
        *Test case for checking on port information*

        **Test Scenario:**
        #. Get port information using zos client
        #. Get port information using bash
        #. Compare port zos results to that of the bash results, should be the same
        """
        self.lg('Get port information using zos client')
        port_info_1 = self.client.info.port()

        self.lg('Get port information using bash (netstat)')
        port_info_2 = self.get_port_info(self.client, protocol)

        if protocol == 'tcp':
            start = 0
        else:
            for i,n in enumerate(port_info_1):
                if n['network'] == 'udp':
                    start = i
                    break

        self.lg('Compare port zos results to that of the bash results, should be the same')
        # check ip, port and pid are the same
        for idx in range(0 ,len(port_info_2['ip'])):
            self.assertEqual(port_info_1[idx + start]['ip'], port_info_2['ip'][idx])
            self.assertEqual(port_info_1[idx + start]['port'], int(port_info_2['port'][idx]))
            self.assertEqual(port_info_1[idx + start]['pid'], int(port_info_2['pid'][idx]))
