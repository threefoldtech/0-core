## `memory_vs_containers.py` script:
     It plots the number of running containers vs memory consumption
    
### Usage
```usage: memory_vs_containers.py [-h] [--conts_num CONTS_NUM]
                               [--time_bet_cont TIME_BET_CONT] --zos_ip ZOS_IP
                               [--jwt JWT] [--teardown]

optional arguments:
  -h, --help            show this help message and exit
  --conts_num CONTS_NUM
                        Number of containers that need to be created
  --time_bet_cont TIME_BET_CONT
                        Time between creating containers
  --zos_ip ZOS_IP       IP for zos machine
  --jwt JWT             JWT for authentiicating ZOS client
  --teardown            Deleting the containers used for this test 
 ```


## More statstics:
To have more statstics beside memory consumtion like cpu load, network behaviours, etc.., You can use [node exporter](https://github.com/prometheus/node_exporter)
There is a ready flist which has node exporter installed and you can use it directly to get all information needed for zero-os.

1- Create container using the given flist in which `host_network=True` and `privileged=True`
   ```
   cl.container.create('https://hub.grid.tf/kheirj/prometheus_ubuntu_16.flist', host_network=True, nics=[{'type':'default'}], privileged=True, tags=['zos_statstics'])
   ```

2- Get a client for the container and start node_exporter

```
   cont_cl.bash('./opt/node_exporter-0.18.0.linux-amd64/node_exporter')
   cont_cl.bash('cd /opt/prometheus-2.10.0-rc.0.linux-amd64; ./prometheus')
```
