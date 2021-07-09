#!/bin/bash
docker-compose down
echo 'docker-compose down, done'

docker image rm read_replication_demo4_mysql_master
docker image rm read_replication_demo4_mysql_slave1
docker image rm read_replication_demo4_mysql_slave2
echo 'docker image rm, done'

rm -rf mysql/master/db*
rm -rf mysql/slave1/db*
rm -rf mysql/slave2/db*
echo 'data rm, done'

docker-compose up -d
echo 'docker-compose up, done'