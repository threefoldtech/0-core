#!/bin/bash

# install zerotier and join the network
curl -s https://install.zerotier.com/ | sudo bash
zerotier-cli join ${ZT_NET_ID}
# fix for travis - zerotier issue
sudo ifconfig "$(ls /sys/class/net | grep zt)" mtu 1280

# generate a key
mkdir -p ~/.ssh
ssh-keygen -f ~/.ssh/id_rsa -P ''
eval `ssh-agent -s`
ssh-add ~/.ssh/id_rsa
