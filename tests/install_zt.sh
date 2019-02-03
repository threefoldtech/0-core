#!/bin/bash

# install zerotier and join the network
curl -s https://install.zerotier.com/ | sudo bash
sudo zerotier-cli join ${ZT_NET_ID}
# fix for travis - zerotier issue
sudo ifconfig "$(ls /sys/class/net | grep zt)" mtu 1280

# generate a key
sudo ssh-keygen -t rsa -N "" -f /root/.ssh/id
export pub_key=$(sudo cat /root/.ssh/id_rsa.pub)
