#!/bin/bash
set -e

sudo ssh-keygen -t rsa -N "" -f /root/.ssh/id_rsa
export SSHKEYNAME=id_rsa

export JUMPSCALEBRANCH=${JUMPSCALEBRANCH:-development}
export JSFULL=1

curl https://raw.githubusercontent.com/threefoldtech/jumpscale_core/$JUMPSCALEBRANCH/install.sh?$RANDOM > /tmp/install_jumpscale.sh;sudo -HE bash -c 'bash /tmp/install_jumpscale.sh'

# create ssh key for jumpscale config manager
mkdir -p ~/.ssh
ssh-keygen -f ~/.ssh/id_rsa -P ''
eval `ssh-agent -s`
ssh-add ~/.ssh/id_rsa

# initialize jumpscale config manager
mkdir -p /opt/code/config_test
git init /opt/code/config_test
touch /opt/code/config_test/.jsconfig
js_config init --silent --path /opt/code/config_test/ --key ~/.ssh/id_rsa
sleep 10
