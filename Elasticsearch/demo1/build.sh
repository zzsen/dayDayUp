#!/bin/bash
sh clear.sh

# 修改vm.max_map_count, 也可以通过vim修改/etc/sysctl.conf, 加上"vm.max_map_count=262144", 然后再通过sysctl -p重启服务
sudo sysctl -w vm.max_map_count=262144

# 创建docker网络
docker network create elastic-network

# 分别启动多个容器
docker run -itd --name esn01 -p 9200:9200 -v esdata01:/usr/share/elasticsearch/data --network elastic-network -e "node.name=esn01" -e "cluster.name=liuxg-docker-cluster" -e "cluster.initial_master_nodes=esn01" -e "bootstrap.memory_lock=true" --ulimit memlock=-1:-1 -e ES_JAVA_OPTS="-Xms512m -Xmx512m" docker.elastic.co/elasticsearch/elasticsearch:7.5.0
docker run -itd --name esn02 -p 9201:9200 -v esdata02:/usr/share/elasticsearch/data --network elastic-network -e "node.name=esn02" -e "cluster.name=liuxg-docker-cluster" -e "discovery.seed_hosts=esn01" -e "bootstrap.memory_lock=true" --ulimit memlock=-1:-1 -e ES_JAVA_OPTS="-Xms512m -Xmx512m" docker.elastic.co/elasticsearch/elasticsearch:7.5.0
docker run -itd --name esn03 -p 9202:9200 -v esdata03:/usr/share/elasticsearch/data --network elastic-network -e "node.name=esn03" -e "cluster.name=liuxg-docker-cluster" -e "discovery.seed_hosts=esn01,esn02" -e "bootstrap.memory_lock=true" --ulimit memlock=-1:-1 -e ES_JAVA_OPTS="-Xms512m -Xmx512m" docker.elastic.co/elasticsearch/elasticsearch:7.5.0


docker run -itd --link esn01:elasticsearch --name kibana --network elastic-network -p 5601:5601 docker.elastic.co/kibana/kibana:7.5.0