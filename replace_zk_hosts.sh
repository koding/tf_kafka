#!/bin/bash -x

zk_hosts=`cat ZK_HOSTS`

sed -i "s/localhost:2181/${zk_hosts}/g" "server.properties"
