#!/bin/bash
docker-compose down 
echo '\033[36m docker-compose down, done\033[0m'

rm -rf mysql/master/db*
rm -rf mysql/slave/db*
echo '\033[36m data rm, done, done\033[0m'

docker ps -a