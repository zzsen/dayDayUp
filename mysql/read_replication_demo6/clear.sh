#!/bin/bash
sh mycat/buildImage.sh

docker-compose down 
echo -e '\033[36m docker-compose down, done\033[0m'

docker image rm read_replication_demo6_mysql_master
docker image rm read_replication_demo6_mysql_slave1
docker image rm read_replication_demo6_mysql_slave2
docker image rm read_replication_demo6_mycat1
docker image rm read_replication_demo6_mycat2
echo -e '\033[36m docker image rm, done\033[0m'

rm -rf mysql/master/db*
rm -rf mysql/slave1/db*
rm -rf mysql/slave2/db*
rm -rf mycat/mycat1*
rm -rf mycat/mycat2*
echo -e '\033[36m data rm, done, done\033[0m'

docker ps -a