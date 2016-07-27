#!/bin/bash -x

zk_hosts=`cat ZK_HOSTS`

echo "zookeeper.connect=$zk_hosts" >> server.properties

echo "export ZK_HOSTS=$zk_hosts" >> /opt/ami-scripts/env/zookeeper.sh
