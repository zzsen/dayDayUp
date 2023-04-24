#!/bin/bash
sh clear.sh

docker-compose up -d
echo -e '\033[32m docker-compose up, done\033[0m'

docker ps -a