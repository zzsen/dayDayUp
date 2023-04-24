#!/bin/bash
docker-compose down -v 
echo -e '\033[36m docker-compose down, done\033[0m'

docker ps -a