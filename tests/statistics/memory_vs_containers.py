# script that plots the number of running containers vs memory consumption
from zeroos.core0 import client
import time
from argparse import ArgumentParser
import  matplotlib.pyplot as plt


def get_container_client(tag):
    cont = cl.container.find(tag)
    cont_id = list(cont.keys())[0]
    cont_cl = cl.container.client(cont_id)
    return cont_cl

def create_containers():
    mem_used = []
    flist = 'https://hub.grid.tf/thabet/redis.flist'
    for i in range(options.conts_num):
        print('Creating container No:{}'.format(i+1))
        cl.container.create(flist, nics=[{'type':'default'}], tags=[tag])
        cont_cl = get_container_client(tag)
        cont_cl.system('/usr/bin/redis-server')
        time.sleep(options.time_bet_cont)
        mem_used.append(cl.info.mem()['used'])

    cont_list = list(range(1,options.conts_num + 1))
    return cont_list, mem_used

def delete_containers():
    conts = cl.container.find(tag)
    print('Deleting containers ..')
    for cont_id in list(conts.keys()):
        cl.container.terminate(cont_id)
                                                 
    

if __name__ == "__main__":

    parser = ArgumentParser()
    parser.add_argument("--conts_num", dest="conts_num", type=int, default=100,
                        help="Number of containers that need to be created")
    parser.add_argument("--time_bet_cont", dest="time_bet_cont", type=int, default=4,
                        help="Time between creating containers")
    parser.add_argument("--zos_ip", dest="zos_ip",
                        help="IP for zos machine", required=True)
    parser.add_argument("--jwt", dest="jwt", default='',
                        help="JWT for authentiicating ZOS client")
    parser.add_argument("--teardown", dest="teardown",  action='store_true',
                        help="Deleting the containers used for this test")

    
    options = parser.parse_args()
    cl = client.Client(options.zos_ip, password=options.jwt, testConnectionAttempts=0)
    cl.ping()
    tag = 'test_cont'


    if options.teardown:
        delete_containers()
    else:
        cont_list, mem_used = create_containers()
        plt.plot(cont_list, mem_used)
        plt.xlabel('Number of Containers')
        plt.ylabel('Memory Usage')
        plt.savefig('ZOS_Memory_Usage.png')

