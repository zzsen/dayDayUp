#!/bin/bash
docker-compose down 
echo '\033[36m docker-compose down, done\033[0m'

rm -rf mysql/master/db*
rm -rf mysql/slave1/db*
rm -rf mysql/slave2/db*
echo '\033[36m data rm, done, done\033[0m'

docker-compose up -d
echo "\033[32m docker-compose up, done\033[0m"

docker image rm read_replication_demo4_mysql_master
docker image rm read_replication_demo4_mysql_slave1
docker image rm read_replication_demo4_mysql_slave2
echo '\033[36m docker image rm, done\033[0m'

echo "\033[32m master log\033[0m"
docker logs read_replication_demo4_mysql_master -n 100

echo "\033[32m slave1 log\033[0m"
docker logs read_replication_demo4_mysql_slave1 -n 100

echo "\033[32m slave2 log\033[0m"
docker logs read_replication_demo4_mysql_slave2 -n 100

docker ps -a