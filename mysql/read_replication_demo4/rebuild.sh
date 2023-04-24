#!/bin/bash

sh clear.sh

docker-compose up -d
echo -e '\033[32m docker-compose up, done\033[0m'

echo -e '\033[32m master log\033[0m'
docker logs read_replication_demo4_mysql_master -n 100

echo -e '\033[32m slave1 log\033[0m'
docker logs read_replication_demo4_mysql_slave1 -n 100

echo -e '\033[32m slave2 log\033[0m'
docker logs read_replication_demo4_mysql_slave2 -n 100

docker ps -a