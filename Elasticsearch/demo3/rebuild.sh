#!/bin/bash
sh clear.sh

# 修改vm.max_map_count, 也可以通过vim修改/etc/sysctl.conf, 加上"vm.max_map_count=262144", 然后再通过sysctl -p重启服务
sudo sysctl -w vm.max_map_count=262144

docker-compose up -d
echo -e '\033[32m docker-compose up, done\033[0m'

docker ps -a