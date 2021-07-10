#!/bin/bash
docker-compose down
echo -e '\033[36m docker-compose down, done\033[0m'

docker image rm read_replication_demo4_mysql_master
docker image rm read_replication_demo4_mysql_slave1
docker image rm read_replication_demo4_mysql_slave2
echo -e '\033[36m docker image rm, done\033[0m'

rm -rf mysql/master/db*
rm -rf mysql/slave1/db*
rm -rf mysql/slave2/db*
echo -e '\033[36m data rm, done, done\033[0m'

docker-compose up -d
echo -e "\033[32m docker-compose up, done\033[0m"

docker ps -a